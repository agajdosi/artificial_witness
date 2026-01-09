package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/agajdosi/artificial_suspects/backend/database"
	"github.com/google/uuid"
)

func main() {
	port := flag.String("port", "8080", "Port to run the server on")
	host := flag.String("host", "localhost", "Host to run the server on, for production use 0.0.0.0")
	db_path := flag.String("db-path", "./data/artsus.db", "Path to the database file")
	flag.Parse()

	err := database.EnsureDBAvailable(*db_path)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	// gameplay
	mux.HandleFunc("/new_game", enableCORS(NewGameHandler))
	mux.HandleFunc("/get_game", enableCORS(GetGameHandler))
	mux.HandleFunc("/eliminate_suspect", enableCORS(EliminateSuspectHandler))
	mux.HandleFunc("/next_round", enableCORS(NextRoundHandler))
	mux.HandleFunc("/next_investigation", enableCORS(NextInvestigationHandler))
	// scores
	mux.HandleFunc("/get_scores", enableCORS(GetScoresHandler))
	mux.HandleFunc("/save_score", enableCORS(SaveScoreHandler))
	// AI
	mux.HandleFunc("/get_models", enableCORS(GetModelsHandler))
	mux.HandleFunc("/get_or_generate_answer", enableCORS(GetOrGenerateAnswerHandler))
	// utils
	mux.HandleFunc("/status", enableCORS(statusHandler))

	url := fmt.Sprintf("%s:%s", *host, *port)
	log.Printf("🚀 Starting server on: http://%s", url)
	err = http.ListenAndServe(url, mux)
	if err != nil {
		log.Fatal(err)
	}
}

// CORS middleware for local development. TODO: Remove this for production.
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔍 statusHandler() request: %v", r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func NewGameHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🎮 NewGameHandler() request: %v", r)
	playerUUID := r.URL.Query().Get("player_uuid")
	model := r.URL.Query().Get("model")
	if model == "" {
		log.Printf("NewGameHandler() error: query parameter 'model' cannot be empty!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if playerUUID == "" {
		log.Println("NewGameHandler() warning: player_uuid is empty! Creating new game without player.UUID.")
	}

	game, err := database.NewGame(playerUUID, model)
	if err != nil {
		log.Printf("NewGame() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(game)
	if err != nil {
		log.Printf("NewGame() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("🎮 NewGameHandler() completed successfully.")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// Get the current game for the current player identified by required query parameter player_uuid.
func GetGameHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔍 GetGameHandler() request: %v", r)
	playerUUID := r.URL.Query().Get("player_uuid")
	if playerUUID == "" {
		log.Printf("GetGameHandler() error: player_uuid is empty!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	game, err := database.GetCurrentGame(playerUUID)
	if err != nil {
		log.Printf("GetGame() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(game)
	if err != nil {
		log.Printf("GetGame() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// Get the next investigation for the current game for the current player identified by required query parameter player_uuid.
func NextInvestigationHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔍 NextInvestigationHandler() request: %v", r)
	playerUUID := r.URL.Query().Get("player_uuid")
	if playerUUID == "" {
		log.Printf("NextInvestigationHandler() error: player_uuid is empty!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	game, err := database.GetCurrentGame(playerUUID)
	if err != nil {
		log.Printf("NextInvestigationHandler() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	game.Investigation, err = database.NewInvestigation(game.UUID)
	if err != nil {
		log.Printf("NextInvestigationHandler() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	game.Level, err = database.GetLevel(game.UUID)
	if err != nil {
		log.Printf("NextInvestigationHandler() could not get Level: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(game)
	if err != nil {
		log.Printf("NextInvestigation() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// Get the next round for the current the current player identified by required query parameter player_uuid.
func NextRoundHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔍 NextRoundHandler() request: %v", r)
	playerUUID := r.URL.Query().Get("player_uuid")
	if playerUUID == "" {
		log.Printf("NextRoundHandler() error: player_uuid is empty!")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	game, err := database.GetCurrentGame(playerUUID)
	if err != nil {
		log.Printf("NextRound() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	round, err := database.NewRound(game.Investigation.UUID)
	if err != nil {
		log.Printf("NextRound() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	game.Investigation.Rounds = append(game.Investigation.Rounds, round) // prepend
	log.Printf("New Round %d: %s", game.Level, game.Investigation.Rounds[len(game.Investigation.Rounds)-1].Question.English)

	resp, err := json.Marshal(game)
	if err != nil {
		log.Printf("NextRound() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func GetScoresHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔍 GetScoresHandler() request: %v", r)
	scores, err := database.GetScores()
	if err != nil {
		log.Printf("GetScores() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(scores)
	if err != nil {
		log.Printf("GetScores() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// Based on the UUID (of the current investigation) choose the index (of description) to be used.
// Be consistent across the one UUID, the investigation yet choose differently on next UUID (of investigation).
func randomForThisInvestigation(UUID string, choices int) int {
	if choices <= 0 {
		return 0
	}
	id, err := uuid.Parse(UUID)
	if err != nil {
		fmt.Printf("Error parsing UUID %s: %v", UUID, err)
		return 0
	}

	pseudoRandom := binary.BigEndian.Uint64(id[0:8])
	num := int(pseudoRandom % uint64(choices))

	return num
}

func EliminateSuspectHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🎯 EliminateSuspectHandler() request: %v", r)
	suspectUUID := r.URL.Query().Get("suspect_uuid")
	roundUUID := r.URL.Query().Get("round_uuid")
	investigationUUID := r.URL.Query().Get("investigation_uuid")

	err := database.SaveElimination(suspectUUID, roundUUID, investigationUUID)
	if err != nil {
		log.Printf("EliminateSuspect() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func SaveScoreHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("💰 SaveScoreHandler() request: %v", r)
	name := r.URL.Query().Get("player_name")
	gameUUID := r.URL.Query().Get("game_uuid")
	err := database.SaveScore(name, gameUUID)
	if err != nil {
		log.Printf("SaveScore() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Saved score: player_name: %s, game_uuid: %s", name, gameUUID)
	w.WriteHeader(http.StatusOK)
}

// Get all Models available in the database.
// Results can be ORDERed by price/weight or nothing - by default ID.
// And can be also filtered by allowed_only.
// WARNING: API keys must not leak in here, this goes to public frontend!
func GetModelsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔍 GetModelsHandler() request: %v", r)
	orderBy := r.URL.Query().Get("order_by") // order by price/weight or "" - default ID
	allowedOnly := false
	if r.URL.Query().Get("allowed_only") == "true" {
		allowedOnly = true
	}

	models, err := database.GetModels(allowedOnly, orderBy)
	if err != nil {
		log.Printf("GetModels() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(models)
	if err != nil {
		log.Printf("GetServices() error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// TODO: toto muzeme vlastne oddelat
// 1. generovat answer z newGame anebo z nextRound primo v Gocku
// 2. na frontend pak jen pockat skrze WaitForAnswer
func GetOrGenerateAnswerHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔍 GetOrGenerateAnswerHandler() request: %v", r)
	playerUUID := r.URL.Query().Get("player_uuid")
	if playerUUID == "" {
		errMsg := "query parameter 'player_uuid' cannot be empty!"
		log.Printf("GetOrGenerateAnswerHandler(): %s\n", errMsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMsg))
		return
	}
	game, err := database.GetCurrentGame(playerUUID)
	question := game.Investigation.Rounds[len(game.Investigation.Rounds)-1].Question.English
	if err != nil {
		errMsg := fmt.Sprintf("Error getting currentGame for player_uuid %s: %v", playerUUID, err)
		log.Printf("GetOrGenerateAnswerHandler(): %v\n", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errMsg))
		return
	}

	log.Printf("===> game.Model: %s\n", game.Model)

	// TODO: get the service based on the current game's Model
	service, err := database.GetServiceForModel(game.Model)
	if err != nil {
		errMsg := fmt.Sprintf("Error getting service for model %s: %v", game.Model, err)
		log.Printf("GetOrGenerateAnswerHandler(): %v\n", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errMsg))
		return
	}

	descriptions, err := database.GetDescriptionsForSuspect(
		game.Investigation.CriminalUUID,
		game.Model,
		false, // do not be strict, allow fallback to any description
	)
	if err != nil {
		errMsg := fmt.Sprintf("Error getting descriptions for suspect: %v", err)
		log.Printf("GetOrGenerateAnswerHandler(): %v\n", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errMsg))
		return
	}

	x := randomForThisInvestigation(game.Investigation.UUID, len(descriptions))
	answer, err := database.GenerateAnswer(question, descriptions[x].Description, game.Model, service)
	if err != nil {
		errMsg := fmt.Sprintf("Error generating answer: %v", err)
		log.Printf("GetOrGenerateAnswerHandler(): %v\n", errMsg)
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(errMsg))
		return
	}

	// TODO: move to database.GenerateAnswer()?
	err = database.SaveAnswer(answer, game.Investigation.Rounds[len(game.Investigation.Rounds)-1].UUID)
	if err != nil {
		errMsg := fmt.Sprintf("Error saving answer: %v", err)
		log.Printf("GetOrGenerateAnswerHandler(): %v\n", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errMsg))
		return
	}

	log.Printf("GetOrGenerateAnswerHandler() - generated answer: %s", answer)

	resp, err := json.Marshal(
		database.Answer{ // TODO: add UUID and Timestamp once Answer has its own table
			UUID:      "",
			Text:      answer,
			Timestamp: "",
		})
	if err != nil {
		errMsg := fmt.Sprintf("Error marshalling answer: %v", err)
		log.Printf("GetOrGenerateAnswerHandler(): %v\n", errMsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errMsg))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
