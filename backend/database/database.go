// Copyright (C) 2024 (Andreas Gajdosik) <andreas@gajdosik.org>
// This file is part of project.
//
// project is non-violent software: you can use, redistribute,
// and/or modify it under the terms of the CNPLv7+ as found
// in the LICENSE file in the source code root directory or
// at <https://git.pixie.town/thufie/npl-builder>.
//
// project comes with ABSOLUTELY NO WARRANTY, to the extent
// permitted by applicable law. See the CNPL for details.

package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed default.db
var defaultDB embed.FS

var database *sql.DB

const (
	defaultPlayerName string = "anonymous"
	numSuspect        int    = 15 // How many suspects are in one investigation - there were 12 in original board game.
	emoDB             string = "💾"
)

// MARK: GENERAL DATABASE

// Ensure that database is ready to be used. First, check if gamesDir exists, if not create it.
// Then, check if database file exists, if not create it and initialize it.
// Returns the database connection.
func EnsureDBAvailable(gameDBPath string) error {
	log.Printf("%s Checking the database file at: %s\n", emoDB, gameDBPath)
	_, err := os.Stat(gameDBPath)
	if os.IsNotExist(err) {
		log.Printf("%s Database file %s does not exist, creating it...\n", emoDB, gameDBPath)
		parentDir := filepath.Dir(gameDBPath)
		err = os.MkdirAll(parentDir, 0755)
		if err != nil {
			return err
		}

		log.Printf("%s Opening default.db", emoDB)
		defaultDBFile, err := defaultDB.Open("default.db")
		if err != nil {
			return err
		}
		defer defaultDBFile.Close()

		log.Printf("%s Creating new database file at %s", emoDB, gameDBPath)
		newDBFile, err := os.Create(gameDBPath)
		if err != nil {
			return err
		}
		defer newDBFile.Close()

		log.Printf("%s Copying default.db to %s", emoDB, gameDBPath)
		_, err = io.Copy(newDBFile, defaultDBFile)
		if err != nil {
			return err
		}
		log.Printf("%s Database successfully created from default.db!", emoDB)
	}

	db, err := sql.Open("sqlite3", gameDBPath)
	if err != nil {
		log.Fatal(err)
	}
	database = db
	log.Printf("%s Database successfully opened!", emoDB)

	return nil
}

// MARK: SUSPECT

type Suspect struct {
	UUID      string `json:"UUID"`
	Image     string `json:"Image"`
	Free      bool   `json:"Free"`
	Fled      bool   `json:"Fled"`
	Timestamp string `json:"Timestamp"`
}

func SaveSuspect(suspect Suspect) error {
	var exists bool
	if suspect.UUID == "" {
		return fmt.Errorf("suspect.UUID cannot be empty")
	}

	checkQuery := "SELECT EXISTS(SELECT 1 FROM suspects WHERE image = ?)"
	err := database.QueryRow(checkQuery, suspect.Image).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	timestamp := TimestampNow()
	query := "INSERT into suspects (uuid, image, timestamp) VALUES (?, ?, ?)"
	_, err = database.Exec(query, suspect.UUID, suspect.Image, timestamp)
	if err != nil {
		log.Printf("Could not save Suspect %s (%s): %v", suspect.Image, suspect.UUID, err)
		return err
	}

	return nil
}

// Get the basic suspect data from the Database without Suspect.Free field!
// Because Suspect.Free and Suspect.Fled needs information from table Investigation->Rounds->Eliminations.
func GetSuspect(suspectUUID string) (Suspect, error) {
	var suspect Suspect
	row := database.QueryRow("SELECT uuid, image, timestamp FROM suspects WHERE uuid = $1 LIMIT 1", suspectUUID)
	err := row.Scan(&suspect.UUID, &suspect.Image, &suspect.Timestamp)
	if err != nil {
		log.Printf("Could not load Suspect (%s): %v", suspectUUID, err)
		return suspect, err
	}

	return suspect, nil
}

// Get all Suspects and their complete data for specified Investigation.
// It needs Investigation because we need to iterate over its Rounds and Rounds' Eliminations
// to set Suspect.Free and Suspect.Fled booleans.
func getSuspectsInInvestigation(suspectUUIDs []string, investigation Investigation) ([]Suspect, error) {
	var suspects []Suspect
	eliminatedSuspectUUIDs := make(map[string]struct{})
	for i := range investigation.Rounds {
		round := investigation.Rounds[i]
		for x := range round.Eliminations {
			elimination := round.Eliminations[x]
			eliminatedSuspectUUIDs[elimination.SuspectUUID] = struct{}{}
		}
	}

	var err error
	for x := range suspectUUIDs {
		var suspect Suspect
		suspect, err = GetSuspect(suspectUUIDs[x])
		if err != nil {
			log.Printf("Error iterating over suspects: %v", err)
		}

		if _, found := eliminatedSuspectUUIDs[suspect.UUID]; found {
			if suspect.UUID == investigation.CriminalUUID {
				suspect.Fled = true
			} else {
				suspect.Free = true
			}
		}

		suspects = append(suspects, suspect)
	}

	return suspects, err
}

func GetAllSuspects() ([]Suspect, error) {
	var suspects []Suspect
	rows, err := database.Query("SELECT uuid, image, timestamp FROM suspects", numSuspect)
	if err != nil {
		log.Printf("Could not get random suspects: %v\n", err)
		return suspects, err
	}
	defer rows.Close()

	for rows.Next() {
		var suspect Suspect
		err := rows.Scan(&suspect.UUID, &suspect.Image, &suspect.Timestamp)
		if err != nil {
			log.Printf("Could not scan suspect: %v\n", err)
			return suspects, err
		}
		suspects = append(suspects, suspect)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during suspects rows iteration: %v\n", err)
		return suspects, err
	}

	return suspects, nil
}

// Get suspects and filter them by number of their descriptions by selected model.
// Only suspects with less than limit number of descriptions will be returned.
// Mostly used just for generation of descriptions in dev.go.
func GetSuspectsByDescriptions(limit int, serviceName, modelName string) ([]Suspect, error) {
	var suspects []Suspect
	allSuspects, err := GetAllSuspects()
	if err != nil {
		return suspects, err
	}
	for _, suspect := range allSuspects {
		descriptions, err := GetDescriptionsForSuspect(suspect.UUID, modelName, true) // strictly get only descriptions for this model
		if err != nil {
			fmt.Printf("Error getting descriptions for suspect (%s): %v", suspect.UUID, err)
			continue
		}
		if len(descriptions) >= limit {
			continue
		}
		suspects = append(suspects, suspect)
	}
	return suspects, nil
}

func randomSuspects() ([]Suspect, error) {
	var suspects []Suspect
	rows, err := database.Query("SELECT uuid, image, timestamp FROM suspects ORDER BY RANDOM() LIMIT $1", numSuspect)
	if err != nil {
		log.Printf("Could not get random suspects: %v\n", err)
		return suspects, err
	}
	defer rows.Close()

	for rows.Next() {
		var suspect Suspect
		err := rows.Scan(&suspect.UUID, &suspect.Image, &suspect.Timestamp)
		if err != nil {
			log.Printf("Could not scan suspect: %v\n", err)
			return suspects, err
		}
		suspects = append(suspects, suspect)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during suspects rows iteration: %v\n", err)
		return suspects, err
	}

	return suspects, nil
}

// MARK: PLAYER

// Instance of a Player who plays the Game. Right now it can be only the Investigator.
// Player UUID is generated by the frontend and stored in the browser's localStorage.
type Player struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// MARK: GAME

// User clicks on start and plays until they make a mistake, can be several cases. This is the Game.
type Game struct {
	UUID          string        `json:"uuid"`
	Score         int           `json:"Score"`         // TODO: implement
	Investigator  Player        `json:"Investigator"`  // The human player, right now can play only as investigator
	Timestamp     string        `json:"Timestamp"`     // when game was created
	Model         string        `json:"Model"`         // LLM model used for generating descriptions and answers
	Investigation Investigation `json:"investigation"` // TODO: actually this could be Investigations []Investigation
	Level         int           `json:"level"`         // aka number of Investigations done + 1
	GameOver      bool          `json:"GameOver"`      // TODO: when true, Game is over

}

// Create a new game for the current player identified by their playerUUID.
// Multiple players can play the game at the same time, so we need to identify the player by their playerUUID.
func NewGame(playerUUID, model string) (Game, error) {
	var game Game
	game.UUID = uuid.New().String()
	game.Timestamp = TimestampNow()
	game.Score = 0
	game.Model = model
	game.Investigator = Player{
		UUID: playerUUID,
		Name: defaultPlayerName, // TODO: also pass from the frontend
	}
	err := saveGame(game)
	if err != nil {
		return game, err
	}

	game.Investigation, err = NewInvestigation(game.UUID)
	if err != nil {
		return game, err
	}
	game.Level, err = GetLevel(game.UUID)
	if err != nil {
		return game, err
	}

	return game, err
}

// Get the current game for the current player identified by their playerUUID.
// Multiple players can play the game at the same time, so we need to identify the game by playerUUID.
func GetCurrentGame(playerUUID string) (Game, error) {
	var game Game
	row := database.QueryRow("SELECT uuid, timestamp, score, model FROM games WHERE player_uuid = $1 ORDER BY timestamp DESC LIMIT 1", playerUUID)
	err := row.Scan(&game.UUID, &game.Timestamp, &game.Score, &game.Model)

	// No game found - first play
	if err == sql.ErrNoRows {
		log.Println("Warning: No games in DB, creating new game")
		return NewGame("", "") // TODO: PlayerUUID should be passed from frontend
	}
	if err != nil {
		return game, err
	}

	log.Printf("Got game: %v | %v", game.UUID, game.Timestamp)

	game.Investigation, err = getCurrentInvestigation(game.UUID)
	if err != nil {
		fmt.Println("GetGame()->getCurrentInvestigation(): ", err)
		return game, err
	}

	game.Level, err = GetLevel(game.UUID)
	if err != nil {
		log.Printf("GetCurrentGame() could not get Level: %v\n", err)
		return game, err
	}

	game.GameOver = isGameOver(game)

	return game, nil
}

func saveGame(game Game) error {
	query := `INSERT INTO games (uuid, timestamp, score, investigator, player_uuid, model) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := database.Exec(
		query,
		game.UUID,
		game.Timestamp,
		game.Score,
		game.Investigator.Name,
		game.Investigator.UUID,
		game.Model,
	)
	return err
}

func isGameOver(game Game) bool {
	for x := range game.Investigation.Rounds {
		round := game.Investigation.Rounds[x]
		for y := range round.Eliminations {
			elimination := round.Eliminations[y]
			if elimination.SuspectUUID == game.Investigation.CriminalUUID {
				return true
			}
		}
	}
	return false
}

// MARK: INVESTIGATION

// Investigation is a set of X Suspects, User needs to find a Criminal among them.
type Investigation struct {
	UUID              string    `json:"uuid"`
	GameUUID          string    `json:"game_uuid"`
	Suspects          []Suspect `json:"suspects"`
	Rounds            []Round   `json:"rounds"`            // Ordered from oldest (first) to newest (last), 1st round is [0], 2nd [1] etc.
	CriminalUUID      string    `json:"-"`                 // Do not expose in JSON!
	InvestigationOver bool      `json:"InvestigationOver"` // Last standing is the Criminal
	Timestamp         string    `json:"Timestamp"`
}

func saveInvestigation(investigation Investigation) error {
	if len(investigation.Suspects) != 15 {
		err := fmt.Errorf("Investigation does not have 15 suspects, has %d", (len(investigation.Suspects)))
		log.Printf("Cannot save investigation: %v\n", err)
		return err
	}

	query := `INSERT OR REPLACE INTO investigations
		(uuid, game_uuid, timestamp,
		sus1_uuid,
		sus2_uuid,
		sus3_uuid,
		sus4_uuid,
		sus5_uuid,
		sus6_uuid,
		sus7_uuid,
		sus8_uuid,
		sus9_uuid,
		sus10_uuid,
		sus11_uuid,
		sus12_uuid,
		sus13_uuid,
		sus14_uuid,
		sus15_uuid,
		criminal_uuid
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := database.Exec(query, investigation.UUID, investigation.GameUUID, investigation.Timestamp,
		investigation.Suspects[0].UUID,
		investigation.Suspects[1].UUID,
		investigation.Suspects[2].UUID,
		investigation.Suspects[3].UUID,
		investigation.Suspects[4].UUID,
		investigation.Suspects[5].UUID,
		investigation.Suspects[6].UUID,
		investigation.Suspects[7].UUID,
		investigation.Suspects[8].UUID,
		investigation.Suspects[9].UUID,
		investigation.Suspects[10].UUID,
		investigation.Suspects[11].UUID,
		investigation.Suspects[12].UUID,
		investigation.Suspects[13].UUID,
		investigation.Suspects[14].UUID,
		investigation.CriminalUUID,
	)
	if err != nil {
		log.Printf("Could not save investigation: %v", err)
		return err
	}

	return nil
}

// Create a new Investigation, save it into the database and return it.
// Usage on New Game for initial first Investigation,
// or when Investigation is successfully solved and we need new one.
func NewInvestigation(gameUUID string) (Investigation, error) {
	var i Investigation
	i.UUID = uuid.New().String()
	i.GameUUID = gameUUID
	i.Timestamp = TimestampNow()

	round, err := NewRound(i.UUID)
	if err != nil {
		return i, err
	}
	i.Rounds = append(i.Rounds, round)

	suspects, err := randomSuspects()
	if err != nil {
		return i, err
	}
	i.Suspects = suspects
	cn := rand.IntN(len(suspects))
	i.CriminalUUID = i.Suspects[cn].UUID

	log.Printf("NEW INVESTIGATION, criminal is: no. %d\n", cn+1)
	err = saveInvestigation(i)
	return i, err
}

func getCurrentInvestigation(gameUUID string) (Investigation, error) {
	var investigation = Investigation{GameUUID: gameUUID}
	var suspects_uuids = make([]string, 15)
	log.Printf("Getting investigation for game %s\n", gameUUID)
	row := database.QueryRow(`SELECT uuid, timestamp, criminal_uuid,
		sus1_uuid,
		sus2_uuid,
		sus3_uuid,
		sus4_uuid,
		sus5_uuid,
		sus6_uuid,
		sus7_uuid,
		sus8_uuid,
		sus9_uuid,
		sus10_uuid,
		sus11_uuid,
		sus12_uuid,
		sus13_uuid,
		sus14_uuid,
		sus15_uuid
		FROM investigations WHERE game_uuid = $1 ORDER BY timestamp DESC LIMIT 1`, gameUUID)
	err := row.Scan(&investigation.UUID, &investigation.Timestamp, &investigation.CriminalUUID,
		&suspects_uuids[0],
		&suspects_uuids[1],
		&suspects_uuids[2],
		&suspects_uuids[3],
		&suspects_uuids[4],
		&suspects_uuids[5],
		&suspects_uuids[6],
		&suspects_uuids[7],
		&suspects_uuids[8],
		&suspects_uuids[9],
		&suspects_uuids[10],
		&suspects_uuids[11],
		&suspects_uuids[12],
		&suspects_uuids[13],
		&suspects_uuids[14],
	)
	if err != nil {
		log.Printf("Could not get investigation: %v\n", err)
		return investigation, err
	}

	investigation.Rounds, err = getRounds(investigation.UUID)
	if err != nil {
		return investigation, err
	}

	investigation.Suspects, err = getSuspectsInInvestigation(suspects_uuids, investigation)
	if err != nil {
		return investigation, err
	}

	eliminated := 0
	for x := range investigation.Rounds {
		eliminated += len(investigation.Rounds[x].Eliminations)
	}
	if eliminated == (numSuspect - 1) {
		investigation.InvestigationOver = true
	}

	return investigation, nil
}

// MARK: ROUND

type Round struct {
	UUID              string        `json:"uuid"`
	InvestigationUUID string        `json:"InvestigationUUID"`
	Question          Question      `json:"Question"`
	AnswerUUID        string        `json:"AnswerUUID"`
	Answer            string        `json:"answer"` // TODO: Answer could be actually stored in table
	Eliminations      []Elimination `json:"Eliminations"`
	Timestamp         string        `json:"Timestamp"`
}

func saveRound(r Round) error {
	query := `
		INSERT OR REPLACE INTO rounds (uuid, investigation_uuid, question_uuid, answer, timestamp)
		VALUES (?, ?, ?, ?, ?)
		`
	_, err := database.Exec(query, r.UUID, r.InvestigationUUID, r.Question.UUID, r.Answer, r.Timestamp)
	return err
}

func NewRound(investigationUUID string) (Round, error) {
	var r Round
	r.UUID = uuid.New().String()
	r.InvestigationUUID = investigationUUID
	r.Timestamp = TimestampNow()
	question, err := GetRandomQuestion()
	if err != nil {
		return r, err
	}
	r.Question = question

	err = saveRound(r)
	return r, err
}

func getRounds(investigationUUID string) ([]Round, error) {
	var rounds []Round
	log.Println("Getting rounds for investigation", investigationUUID)

	rows, err := database.Query("SELECT uuid, investigation_uuid, question_uuid, answer, timestamp FROM rounds WHERE investigation_uuid = $1 ORDER BY timestamp ASC", investigationUUID)
	if err != nil {
		log.Printf("Could not get rounds: %v\n", err)
		return rounds, err
	}
	defer rows.Close()

	for rows.Next() {
		var round Round
		err := rows.Scan(&round.UUID, &round.InvestigationUUID, &round.Question.UUID, &round.Answer, &round.Timestamp)
		if err != nil {
			log.Printf("Could not scan round: %v\n", err)
			return rounds, err
		}

		question, err := getQuestion(round.Question.UUID)
		if err != nil {
			log.Printf("Could not get question text for question_uuid=%s: %v", round.Question.UUID, err)
			return rounds, err
		}
		round.Question = question

		round.Eliminations, err = getEliminationsForRound(round.UUID)
		if err != nil {
			log.Printf("Could not get Eliminations for Round (%s): %v\n", round.UUID, err)
			return rounds, err
		}

		rounds = append(rounds, round)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during rows iteration: %v\n", err)
		return rounds, err
	}

	log.Println("Got rounds:", rounds)

	return rounds, nil
}

// MARK: ELIMINATION

type Elimination struct {
	UUID        string `json:"UUID"`
	RoundUUID   string `json:"RoundUUID"`
	SuspectUUID string `json:"SuspectUUID"`
	Timestamp   string `json:"Timestamp"`
}

// Save the Elimination, check if Criminal was not released
// and if not update the Game.Score accordingly.
func SaveElimination(suspectUUID, roundUUID, investigationUUID string) error {
	UUID := uuid.New().String()
	timestamp := TimestampNow()
	query := `INSERT OR REPLACE INTO eliminations (UUID, RoundUUID, SuspectUUID, Timestamp) VALUES (?, ?, ?, ?)`
	_, err := database.Exec(query, UUID, roundUUID, suspectUUID, timestamp)
	if err != nil {
		log.Printf("Could not save elimination of Suspect (%s) on Round (%s): %v\n", suspectUUID, roundUUID, err)
		return err
	}

	var criminalUUID string
	var gameUUID string
	row := database.QueryRow(`SELECT criminal_uuid, game_uuid FROM investigations WHERE uuid = $1`, investigationUUID)
	err = row.Scan(&criminalUUID, &gameUUID)
	if err != nil {
		log.Printf("Could not get criminal_uuid on Investigation (%s): %v\n", investigationUUID, err)
	}

	if criminalUUID != suspectUUID {

		increaseScore(gameUUID, roundUUID)
	} else {
		log.Println("Guilty criminal was released :(")
	}

	return nil
}

func getEliminationsForRound(roundUUID string) ([]Elimination, error) {
	var eliminations []Elimination
	log.Printf("Getting Eliminations for Round (%s)\n", roundUUID)

	rows, err := database.Query("SELECT UUID, RoundUUID, SuspectUUID, Timestamp FROM eliminations WHERE RoundUUID = $1 ORDER BY timestamp DESC", roundUUID)
	if err != nil {
		log.Printf("Could not get Eliminations: %v\n", err)
		return eliminations, err
	}
	defer rows.Close()

	for rows.Next() {
		var elimination Elimination
		err := rows.Scan(&elimination.UUID, &elimination.RoundUUID, &elimination.SuspectUUID, &elimination.Timestamp)
		if err != nil {
			log.Printf("Could not scan Elimination: %v\n", err)
			return eliminations, err
		}

		eliminations = append(eliminations, elimination)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error during Eliminations rows iteration: %v\n", err)
		return eliminations, err
	}

	fmt.Println("GOT ELIMINATIONS:", eliminations)

	return eliminations, nil
}

// MARK: QUESTION

type Question struct {
	UUID    string `json:"UUID"`
	English string `json:"English"`
	Czech   string `json:"Czech"`
	Polish  string `json:"Polish"`
	Topic   string `json:"Topic"`
	Level   int    `json:"Level"`
}

func GetRandomQuestion() (Question, error) {
	var question Question
	row := database.QueryRow("SELECT UUID, English, Czech, Polish, Topic, Level FROM questions ORDER BY RANDOM() LIMIT 1")
	err := row.Scan(&question.UUID, &question.English, &question.Czech, &question.Polish, &question.Topic, &question.Level)
	return question, err
}

// English is the cannonical text. If question with same English version exists, it will not overwrite.
func SaveQuestion(q Question) error {
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM questions WHERE English = ?)"
	err := database.QueryRow(checkQuery, q.English).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	UUID := uuid.New().String()
	query := "INSERT into questions (UUID, English, Czech, Polish, Topic, Level) VALUES (?, ?, ?, ?, ?, ?)"
	_, err = database.Exec(query, UUID, q.English, q.Czech, q.Polish, q.Topic, q.Level)
	if err != nil {
		log.Printf("Could not save Question %s (%s): %v", q.English, UUID, err)
		return err
	}

	return nil
}

func getQuestion(questionUUID string) (Question, error) {
	var question = Question{UUID: questionUUID}
	row := database.QueryRow("SELECT English, Czech, Polish, Topic, Level FROM questions WHERE UUID = $1 LIMIT 1", questionUUID)
	err := row.Scan(&question.English, &question.Czech, &question.Polish, &question.Topic, &question.Level)
	if err != nil {
		log.Printf("Could not scan question (%s): %v", questionUUID, err)
		return question, err
	}
	return question, nil
}

// MARK: ANSWER

type Answer struct {
	UUID      string `json:"UUID"`
	Text      string `json:"Text"`
	Timestamp string `json:"Timestamp"`
}

// Save the Answer to the Round record in the database. There is then func WaitForAnswer()
// which is called from frontend once new Round is found (and so Question can be shown ASAP).
// But Answer takes time and when it is saved here the WaitForAnswer() retrieves it later.
func SaveAnswer(answer, roundUUID string) error {
	query := "UPDATE rounds SET answer = $1 WHERE uuid = $2"
	result, err := database.Exec(query, answer, roundUUID)
	if err != nil {
		log.Printf("Error updating answer for round %s: %v", roundUUID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error fetching rows affected for round %s: %v", roundUUID, err)
		return err
	}
	if rowsAffected == 0 {
		log.Printf("No rows were updated for round %s", roundUUID)
	}

	return nil
}

// MARK: LEVEL & SCORE

func GetLevel(gameUUID string) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM investigations WHERE game_uuid = $1"

	err := database.QueryRow(query, gameUUID).Scan(&count)
	if err != nil {
		return -1, fmt.Errorf("error counting investigations records for game_uuid %s: %v", gameUUID, err)
	}

	return count, nil
}

// Increase the game.Score in the database after successful Elimination.
// Amount of increase is based on in which level we are and if it is 1st, 2nd or Nth
// Elimination in this round. Players are rewarded for risky behaviour - eliminating more than one suspect.
// But also they are rewarded for longevity - how much investigations they have solved.
func increaseScore(gameUUID string, roundUUID string) {
	level, err := GetLevel(gameUUID)
	if err != nil {
		log.Println("Could not get level and increase score:", err)
		return
	}
	eliminations, err := getEliminationsForRound(roundUUID)
	if err != nil {
		log.Println("Could not get eliminations for this round and increase score:", err)
		return
	}

	amount := level * len(eliminations)

	query := "UPDATE games SET score = score + $1 WHERE uuid = $2"
	_, err = database.Exec(query, amount, gameUUID)
	if err != nil {
		log.Printf("error increasing score for gameUUID %s: %v", gameUUID, err)
		return
	}
	fmt.Printf("Score increased by %d\n", amount)
}

// This is used for High Scores list.
type FinalScore struct {
	Score        int    `json:"Score"`
	Position     int    `json:"Position"`
	Investigator string `json:"Investigator"`
	GameUUID     string `json:"GameUUID"`
	Timestamp    string `json:"Timestamp"`
}

func GetScores() ([]FinalScore, error) {
	var scores []FinalScore
	query := "SELECT uuid, score, investigator FROM games ORDER BY score DESC"
	rows, err := database.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get scores: %w", err)
	}
	defer rows.Close()

	// Loop through the result set and scan into the games slice
	var position int
	for rows.Next() {
		position++
		var finalScore FinalScore
		var investigator sql.NullString
		var score sql.NullInt64
		err := rows.Scan(&finalScore.GameUUID, &score, &investigator)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		finalScore.Position = position
		finalScore.Investigator = investigator.String
		if score.Valid {
			finalScore.Score = int(score.Int64)
		}
		scores = append(scores, finalScore)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return scores, nil
}

func SaveScore(name, gameUUID string) error {
	query := "UPDATE games SET investigator = $1 WHERE uuid = $2"
	_, err := database.Exec(query, name, gameUUID)
	if err != nil {
		log.Printf("error saving investigator for gameUUID %s: %v", gameUUID, err)
	}
	return err
}

// MARK: AI SERVICES

// Service is an LLM provider. It can be OpenAI, Anthropic, DeepSeek, or local model served via LiteLLM.
type Service struct {
	Name      string         `json:"Name"`      // Name presented to the user
	API_style sql.NullString `json:"API_style"` // What is the style of the API (openai, deepseek, etc) - we can have DeepSeek provided via LiteLLM (which uses openai API style)
	Type      string         `json:"Type"`      // API or local
	URL       sql.NullString `json:"URL"`
	Token     string         `json:"Token"`
	Active    bool           `json:"Active"`
}

func GetService(name string) (Service, error) {
	var service Service
	query := "SELECT Name, API_style, Type, URL, Token, Active FROM services WHERE name = $1"
	err := database.QueryRow(query, name).Scan(&service.Name, &service.API_style, &service.Type, &service.URL, &service.Token, &service.Active)
	if err != nil {
		return service, fmt.Errorf("error geting Service for name %s: %v", name, err)
	}
	return service, nil
}

func GetServices() ([]Service, error) {
	var services []Service
	query := "SELECT Name, API_style, Type, URL, Token, Active FROM services"
	rows, err := database.Query(query)
	if err != nil {
		return services, err
	}
	defer rows.Close()

	for rows.Next() {
		var service Service
		err := rows.Scan(&service.Name, &service.API_style, &service.Type, &service.URL, &service.Token, &service.Active)
		if err != nil {
			return services, err
		}
		services = append(services, service)
	}

	if err = rows.Err(); err != nil {
		return services, err
	}
	return services, nil
}

// Get Service by name of the Model which it provides.
// We know Model name, and we need to get the Service which provides it.
func GetServiceForModel(modelName string) (Service, error) {
	var service Service
	model, err := GetModel(modelName)
	if err != nil {
		return service, fmt.Errorf("could not get model %s for service lookup: %v", modelName, err)
	}

	query := "SELECT Name, API_style, Type, URL, Token, Active FROM services WHERE name = $1"
	row := database.QueryRow(query, model.Service)
	err = row.Scan(&service.Name, &service.API_style, &service.Type, &service.URL, &service.Token, &service.Active)
	if err != nil {
		return service, fmt.Errorf("error geting Service for model %s: %v", modelName, err)
	}

	return service, nil
}

// Wait until non-empty Answer appears on the Round record in Rounds table.
// Timeouts in 60 seconds, retries every 1 second. Answer is not modified anyhow,
// needs to be handled after return. On error returned answer is "". On error during
// generation of the answer, it is something like "failed OpenAI()". If everything was
// ok, the answer should be YES or NO, or lowercase variant - parse it later!
func WaitForAnswer(roundUUID string) string {
	pollInterval := 1 * time.Second
	timeout := 60 * time.Second
	start := time.Now()
	var answer string
	for {
		err := database.QueryRow("SELECT answer FROM rounds WHERE uuid = $1", roundUUID).Scan(&answer)
		if err == sql.ErrNoRows {
			log.Printf("Answer not available yet for Round (%s). Retrying...\n", roundUUID)
		} else if err != nil {
			log.Printf("Error querying answer for Round (%s), err: %v\n", roundUUID, err)
			return ""
		} else if answer == "" {
			log.Printf("Answer is still empty, lets sleep for a while...")
		} else {
			log.Printf("Answer found: %s", answer)
			return answer
		}
		if time.Since(start) > timeout {
			log.Printf("timed out waiting for answer to be available on Round (%s)\n", roundUUID)
			return ""
		}
		time.Sleep(pollInterval) // Wait for the polling interval before checking again
	}
}

// MARK: AI MODELS

type Model struct {
	Name       string `json:"Name"`
	Service    string `json:"Service"`    // Service  which provides this model (OpenAI, Anthropic, DeepSeek)
	Visual     bool   `json:"Visual"`     // Model has visual capabilities
	Allowed    bool   `json:"Allowed"`    // Model can be used to play the Game right now
	Historical bool   `json:"Historical"` // Model can be shown in the historical statistics
}

// Get all available Models from the database.
func GetModels(allowedOnly bool, orderBy string) ([]Model, error) {
	var models []Model
	var query string
	order := ""
	if orderBy == "price" {
		order = "ORDER BY price"
	}
	if orderBy == "weight" {
		order = "ORDER BY weight"
	}
	where := ""
	if allowedOnly {
		where = "WHERE Allowed = 1"
	}

	query = fmt.Sprintf("SELECT Name, Service, Visual, Allowed, Historical FROM models %s %s", where, order)

	fmt.Println("QUERY:", query)
	rows, err := database.Query(query)
	if err != nil {
		return models, err
	}
	defer rows.Close()

	for rows.Next() {
		var model Model
		err := rows.Scan(&model.Name, &model.Service, &model.Visual, &model.Allowed, &model.Historical)
		if err != nil {
			return models, err
		}
		models = append(models, model)
	}

	if err = rows.Err(); err != nil {
		return models, err
	}
	return models, nil
}

// Get Model specified by its name from the database.
func GetModel(name string) (Model, error) {
	var model Model
	query := "SELECT Name, Service, Visual, Allowed, Historical FROM models WHERE Name = $1"
	err := database.QueryRow(query, name).Scan(&model.Name, &model.Service, &model.Visual, &model.Allowed, &model.Historical)
	if err != nil {
		return model, fmt.Errorf("error geting Model for name %s: %v", name, err)
	}
	return model, nil
}

// MARK: DESCRIPTIONS

// Holds description of the Suspect image. There can be multiple descriptions for one Suspect.
// Descriptions can be made by different Services and different Models.
type Description struct {
	UUID        string `json:"UUID"`
	SuspectUUID string `json:"SuspectUUID"`
	Service     string `json:"Service"`
	Model       string `json:"Model"`
	Description string `json:"Description"`
	Prompt      string `json:"Prompt"`
	Timestamp   string `json:"Timestamp"`
}

func SaveDescription(d Description) error {
	query := `
		INSERT OR REPLACE INTO descriptions (UUID, SuspectUUID, Service, Model, Description, Prompt, Timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	timestamp := TimestampNow()
	if d.UUID == "" {
		d.UUID = uuid.New().String()
	}
	_, err := database.Exec(query, d.UUID, d.SuspectUUID, d.Service, d.Model, d.Description, d.Prompt, timestamp)
	return err
}
