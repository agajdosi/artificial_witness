import { writable } from 'svelte/store';
import type { ErrorMessage, Game, Player } from '$lib/main';


// GAME STATE
const storedGame = localStorage.getItem('currentGame');
const defaultGame: Game = {
    uuid: '',
    level: 0,
    Score: 0,
    investigation: {
        uuid: '',
        game_uuid: '',
        suspects: [],
        rounds: [],
        CriminalUUID: '',
        InvestigationOver: false,
        Timestamp: ''
    },
    GameOver: false,
    Investigator: '',
    Model: '',
    Timestamp: ''
};
export const currentGame = writable<Game>(storedGame ? JSON.parse(storedGame) : defaultGame);
currentGame.subscribe((value) => {
    localStorage.setItem('currentGame', JSON.stringify(value));
});

// ErrorMessage
const defaultErrorMessage: ErrorMessage = {
    Severity: '',
    Title: '',
    Message: '',
    Actions: []
};
export const errorMessage = writable<ErrorMessage>(defaultErrorMessage);

// Hint
export const hint = writable<string>("");


// MARK: Stored PLAYER
// TODO: actually we can use `import { v4 as uuidv4 } from 'uuid'`;
const generateUUID = (): string => {
    if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
        return crypto.randomUUID();
    }
    return `player-${Date.now()}-${Math.random().toString(16).slice(2)}`;
};

const createNewPlayer = (): Player => ({
    UUID: generateUUID(),
    Name: '',
    SeenIntro: false
});

const storedPlayer = localStorage.getItem('player');
let initialPlayer: Player;
if (storedPlayer) {
    try {
        const parsed = JSON.parse(storedPlayer) as Player;
        if (!parsed.UUID) {
            initialPlayer = createNewPlayer();
            localStorage.setItem('player', JSON.stringify(initialPlayer));
        } else {
            initialPlayer = parsed;
        }
    } catch (error) {
        console.error('Failed to parse stored player, creating new one.', error);
        initialPlayer = createNewPlayer();
        localStorage.setItem('player', JSON.stringify(initialPlayer));
    }
} else {
    initialPlayer = createNewPlayer();
    localStorage.setItem('player', JSON.stringify(initialPlayer));
}

export const currentPlayer = writable<Player>(initialPlayer);
currentPlayer.subscribe((value) => {
    localStorage.setItem('player', JSON.stringify(value));
});

// Selected model (AI) for a new game - persisted in localStorage so navigation/reloads keep selection
const storedSelectedModel = localStorage.getItem('selectedModel');
let initialSelectedModel: string | null = storedSelectedModel ? JSON.parse(storedSelectedModel) : null;
export const selectedModel = writable<string | null>(initialSelectedModel);
selectedModel.subscribe((value) => {
    localStorage.setItem('selectedModel', JSON.stringify(value));
});
