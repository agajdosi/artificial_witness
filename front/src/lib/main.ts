import { currentGame, currentPlayer } from '$lib/stores';
import { get } from 'svelte/store';

// MARK: CONSTANTS

const API_URL = import.meta.env.PROD ? 'https://api.artificialwitness.com/' : 'http://localhost:8080';
const initGET = {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json',
    },
}
const initPOST = {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
    },
}

// MARK: TYPES

export interface Answer {
    UUID: string;
    Text: string;
    Timestamp: string;
}

export interface Elimination {
    UUID: string;
    RoundUUID: string;
    SuspectUUID: string;
    Timestamp: string;
}

export interface ErrorMessage {
    Severity: string;
    Title: string;
    Message: string;
    Actions: string[];
}

export interface FinalScore {
    GameUUID: string;
    Score: number;
    Investigator: string; // AKA player name
}

export interface Player {
    UUID: string;
    Name: string;
}

export interface Game {
    uuid: string;
    investigation: Investigation;
    level: number;
    Score: number;
    GameOver: boolean;
    Investigator: string;
    Model: string;
    Timestamp: string;
}

export interface Investigation {
    uuid: string;
    game_uuid: string;
    suspects: Suspect[];
    rounds: Round[];
    CriminalUUID: string;
    InvestigationOver: boolean;
    Timestamp: string;
}

export interface Model {
    Name: string;
    Service: string;
    Visual: boolean;
    Allowed: boolean;
    Historical: boolean;
}

export interface Round {
    uuid: string;
    InvestigationUUID: string;
    Question: Question;
    AnswerUUID: string;
    answer: string;
    Eliminations: Elimination[];
    Timestamp: string;
}

export interface Service {
    Name: string;
    API_style: string; // Style of the API (openAI, anthropic, etc.)
    Type: string; // API or local
    URL: string;
    Token: string;
    Active: boolean
}

export interface Suspect {
    UUID: string;
    Image: string;
    Free: boolean;
    Fled: boolean;
    Timestamp: string;
}

export interface Question {
    UUID: string;
    English: string;
    Czech: string;
    Polish: string;
    Topic: string;
    Level: number;
}


// MARK: FUNCTIONS

export async function NewGame(model: string): Promise<Game> {
    console.log("NEW GAME requested!");
    let newGame: Game;
    try {
        const player = get(currentPlayer);
        const response = await fetch(`${API_URL}/new_game?player_uuid=${player.UUID}&model=${model}`, initGET);
        if (!response.ok) {
            throw new Error('Failed to create new game');
        }
        newGame = await response.json();
        console.log(`NewGame() response: ${newGame}`);
        currentGame.set(newGame);
    } catch (error) {
        console.log(`NewGame() has failed: ${error}`);
        throw error;
    }

    const lastRoundUUID = newGame.investigation.rounds.at(-1)?.uuid;
    if (!lastRoundUUID) {
        throw new Error('Last Round UUID not found in new game');
    }
    const answer = await generateAnswer(lastRoundUUID);

    if (newGame.investigation.rounds.at(-1)) {
        const answerText = answer?.Text;
        if (!answerText) {
            throw new Error('Generated answer is empty');
        }
        if (!newGame.investigation.rounds[newGame.investigation.rounds.length - 1]) {
            throw new Error('Last round not found in new game');
        } 
        newGame.investigation.rounds[newGame.investigation.rounds.length - 1].answer = answerText;
    }

    currentGame.set(newGame);
    return newGame;
}

export async function GetGame(): Promise<Game> {
    const player = get(currentPlayer);
    const response = await fetch(`${API_URL}/get_game?player_uuid=${player.UUID}`, initGET);
    if (!response.ok) {
        throw new Error('Failed to fetch game');
    }
    
    return await response.json();
}

export async function NextRound() {
    // FIRST GET THE NEW ROUND`
    const player = get(currentPlayer);
    const response = await fetch(`${API_URL}/next_round?player_uuid=${player.UUID}`, initGET);
    if (!response.ok) {
        throw new Error('Failed to fetch next round');
    }

    let game: Game = await response.json();
    console.log(`>>> NEW ROUND: ${game.investigation.rounds.at(-1)}`);
    currentGame.set(game);

    // THEN GENERATE ANSWER
    const lastRoundUUID = game.investigation.rounds.at(-1)?.uuid;
    if (!lastRoundUUID) {
        throw new Error('Last Round UUID not found in new game');
    }
    const answer = await generateAnswer(lastRoundUUID);

    if (game.investigation.rounds.at(-1)) {
        const answerText = answer?.Text;
        if (!answerText) {
            throw new Error('Generated answer is empty');
        }
        if (!game.investigation.rounds[game.investigation.rounds.length - 1]) {
            throw new Error('Last round not found in new game');
        } 
        game.investigation.rounds[game.investigation.rounds.length - 1].answer = answerText;
    }
    currentGame.set(game);
}

export async function NextInvestigation(): Promise<Game> {
    const player = get(currentPlayer);
    const response = await fetch(`${API_URL}/next_investigation?player_uuid=${player.UUID}`, initGET);

    if (!response.ok) {
        throw new Error('Failed to fetch next investigation');
    }

    return await response.json();
}

export async function EliminateSuspect(suspectUUID: string, roundUUID: string, investigationUUID: string): Promise<void> {
    const response = await fetch(`${API_URL}/eliminate_suspect?suspect_uuid=${suspectUUID}&round_uuid=${roundUUID}&investigation_uuid=${investigationUUID}`, initPOST);
    if (!response.ok) {
        throw new Error('Failed to eliminate suspect');
    }
}

export async function WaitForAnswer(roundUUID: string): Promise<string> {
    const response = await fetch(`${API_URL}/wait_for_answer?round_uuid=${roundUUID}`, initGET);
    if (!response.ok) {
        throw new Error('Failed to wait for answer');
    }

    return await response.json();
}

export async function GetScores(): Promise<FinalScore[]> {
    const response = await fetch(`${API_URL}/get_scores`, initGET);

    if (!response.ok) {
        throw new Error('Failed to fetch scores');
    }

    return await response.json();
}

export async function SaveScore(playerName: string, gameUUID: string) {
    const response = await fetch(`${API_URL}/save_score?player_name=${playerName}&game_uuid=${gameUUID}`, initPOST);
    if (!response.ok) {
        throw new Error('Failed to save score');
    }
}


export async function ListModelsOllama(): Promise<Model[]> {
    let models: Model[] = [];
    return models;
}

export async function ListAvailableModels(allowedOnly: boolean, orderBy: string): Promise<Model[]> {
    const response = await fetch(`${API_URL}/get_models?allowed_only=${allowedOnly}&order_by=${orderBy}`, initGET);
    if (!response.ok) {
        throw new Error('Failed to fetch models');
    }
    return await response.json();
}

export async function generateAnswer(roundUUID: string): Promise<Answer|undefined> {
    const player = get(currentPlayer);
    console.log(`>>> generateAnswer called! roundUUID=${roundUUID}`);
    try {
        let answer: Answer; 
        const response = await fetch(`${API_URL}/get_or_generate_answer?player_uuid=${player.UUID}`, initGET);
        if (!response.ok) {
            throw new Error('Failed to /get_or_generate_answer');
        }

        answer = await response.json() as Answer;
        console.log(`Got answer: ${answer}`);
        return answer;
    } catch (error) {
        console.error(`generateAnswer error for round ${roundUUID}:`, error);
        // TODO: communicate failure to the user and GUI
    }
}