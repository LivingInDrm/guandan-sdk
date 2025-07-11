// Card types
export interface Card {
  suit: 'Hearts' | 'Diamonds' | 'Clubs' | 'Spades' | 'Joker';
  rank: string;
}

// Player types
export interface Player {
  id: string;
  name: string;
  seat: SeatID;
  handCount: number;
  level: number;
  connected: boolean;
}

export type SeatID = 'east' | 'south' | 'west' | 'north';

// Game state types
export interface GameState {
  matchId: string;
  players: Player[];
  currentDeal?: DealState;
  status: 'waiting' | 'playing' | 'finished';
  version: number;
}

export interface DealState {
  dealId: string;
  trump: string;
  phase: string;
  currentTurn: SeatID;
  tablePlay?: CardGroup;
  lastPlayer?: SeatID;
  playerHands: Record<SeatID, Card[]>;
}

export interface CardGroup {
  cards: Card[];
  type: string;
  value: number;
}

// WebSocket message types
export interface WSMessage {
  t: string;
  data?: any;
}

export interface SnapshotMessage extends WSMessage {
  t: 'Snapshot';
  version: number;
  payload: GameState;
}

export interface EventMessage extends WSMessage {
  t: 'Event';
  e: string;
  data: any;
  version: number;
}

export interface PlayCardsMessage extends WSMessage {
  t: 'PlayCards';
  cards: string[];
}

export interface PassMessage extends WSMessage {
  t: 'Pass';
}

// API types
export interface CreateRoomRequest {
  roomName: string;
}

export interface CreateRoomResponse {
  roomId: string;
}

export interface JoinRoomRequest {
  seat: number;
}

export interface JoinRoomResponse {
  wsUrl: string;
}

export interface RoomInfo {
  roomId: string;
  playerCount: number;
  maxPlayers: number;
  isEmpty: boolean;
}

// UI state types
export interface UIState {
  selectedCards: Card[];
  isMyTurn: boolean;
  canPlay: boolean;
  showValidationError: boolean;
  errorMessage: string;
  isDragging: boolean;
}

// Store types
export interface RoomState extends GameState, UIState {
  // Connection state
  wsUrl: string | null;
  isConnected: boolean;
  connectionStatus: 'disconnected' | 'connecting' | 'connected' | 'error';
  
  // Player state
  mySeat: SeatID | null;
  myHand: Card[];
  
  // Actions
  actions: {
    // Connection actions
    connect: (wsUrl: string) => void;
    disconnect: () => void;
    
    // Game actions
    selectCard: (card: Card) => void;
    deselectCard: (card: Card) => void;
    playCards: (cards: Card[]) => void;
    pass: () => void;
    
    // State management
    handleSnapshot: (snapshot: GameState) => void;
    handleEvent: (event: EventMessage) => void;
    setError: (message: string) => void;
    clearError: () => void;
  };
}

// Drag and drop types
export interface DragItem {
  type: string;
  card: Card;
  index: number;
}

export interface DropResult {
  type: string;
  action: 'play' | 'return';
}

// Validation types
export interface ValidationResult {
  isValid: boolean;
  error?: string;
}

// Animation types
export interface AnimationProps {
  duration?: number;
  delay?: number;
  easing?: string;
}

// Theme types
export interface Theme {
  colors: {
    primary: string;
    secondary: string;
    background: string;
    surface: string;
    text: string;
    textSecondary: string;
    border: string;
    error: string;
    success: string;
    warning: string;
  };
  spacing: {
    xs: string;
    sm: string;
    md: string;
    lg: string;
    xl: string;
  };
  borderRadius: {
    sm: string;
    md: string;
    lg: string;
  };
  shadows: {
    sm: string;
    md: string;
    lg: string;
  };
}

// Utility types
export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};

export type Optional<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;

export type RequireAtLeastOne<T, Keys extends keyof T = keyof T> = 
  Pick<T, Exclude<keyof T, Keys>> & 
  { [K in Keys]-?: Required<Pick<T, K>> & Partial<Pick<T, Exclude<Keys, K>>> }[Keys];

// Constants
export const SEATS: SeatID[] = ['east', 'south', 'west', 'north'];

export const CARD_SUITS = ['Hearts', 'Diamonds', 'Clubs', 'Spades'] as const;

export const CARD_RANKS = [
  '2', '3', '4', '5', '6', '7', '8', '9', '10', 
  'J', 'Q', 'K', 'A', '小王', '大王'
] as const;

export const WS_MESSAGE_TYPES = {
  SNAPSHOT: 'Snapshot',
  EVENT: 'Event',
  PLAY_CARDS: 'PlayCards',
  PASS: 'Pass',
  PING: 'ping',
  PONG: 'pong',
  ERROR: 'Error'
} as const;

export const GAME_PHASES = {
  IDLE: 'idle',
  CREATED: 'created',
  CARDS_DEALT: 'cards_dealt',
  TRIBUTE: 'tribute',
  FIRST_PLAY: 'first_play',
  IN_PROGRESS: 'in_progress',
  FINISHED: 'finished'
} as const;

export const CONNECTION_STATUS = {
  DISCONNECTED: 'disconnected',
  CONNECTING: 'connecting',
  CONNECTED: 'connected',
  ERROR: 'error'
} as const;