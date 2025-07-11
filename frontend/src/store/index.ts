import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';
import { subscribeWithSelector } from 'zustand/middleware';
import { 
  RoomState, 
  GameState, 
  Card, 
  SeatID, 
  EventMessage, 
  WSMessage,
  CONNECTION_STATUS 
} from '../types';
import { WSClient, isSnapshotMessage, isEventMessage, isErrorMessage } from '../utils/ws';

// Convert server seat ID (0,1,2,3) to frontend seat ID (east,south,west,north)
function convertSeatIDToString(seatID: number): SeatID {
  console.log('convertSeatIDToString called with:', seatID, 'type:', typeof seatID);
  
  if (seatID === undefined || seatID === null) {
    console.error('convertSeatIDToString: seatID is undefined or null!');
    return 'east';
  }
  
  const seatMap: Record<number, SeatID> = {
    0: 'east',
    1: 'south',
    2: 'west',
    3: 'north'
  };
  
  const result = seatMap[seatID];
  if (result === undefined) {
    console.error('convertSeatIDToString: Invalid seatID:', seatID);
    return 'east';
  }
  
  return result;
}

interface RoomStore extends RoomState {
  wsClient: WSClient | null;
}

export const useRoomStore = create<RoomStore>()(
  subscribeWithSelector(
    immer((set, get) => ({
      // Connection state
      wsUrl: null,
      isConnected: false,
      connectionStatus: CONNECTION_STATUS.DISCONNECTED,
      
      // Game state
      matchId: '',
      players: [],
      currentDeal: undefined,
      status: 'waiting',
      version: 0,
      
      // Player state
      mySeat: null,
      myHand: [],
      
      // UI state
      selectedCards: [],
      isMyTurn: false,
      canPlay: false,
      showValidationError: false,
      errorMessage: '',
      isDragging: false,
      
      // WebSocket client
      wsClient: null,
      
      actions: {
        // Connection actions
        connect: (wsUrl: string) => {
          const state = get();
          if (state.wsClient) {
            state.wsClient.disconnect();
          }
          
          const wsClient = new WSClient({
            url: wsUrl,
            onOpen: () => {
              set((state) => {
                state.isConnected = true;
                state.connectionStatus = CONNECTION_STATUS.CONNECTED;
                state.wsUrl = wsUrl;
              });
            },
            onMessage: (message: WSMessage) => {
              const actions = get().actions;
              
              if (isSnapshotMessage(message)) {
                actions.handleSnapshot(message.payload);
              } else if (isEventMessage(message)) {
                actions.handleEvent(message);
              } else if (isErrorMessage(message)) {
                actions.setError(message.error);
              }
            },
            onClose: () => {
              set((state) => {
                state.isConnected = false;
                state.connectionStatus = CONNECTION_STATUS.DISCONNECTED;
              });
            },
            onError: (_error) => {
              set((state) => {
                state.connectionStatus = CONNECTION_STATUS.ERROR;
                state.errorMessage = 'Connection error occurred';
              });
            },
            onReconnect: (_attempt) => {
              set((state) => {
                state.connectionStatus = CONNECTION_STATUS.CONNECTING;
              });
            }
          });
          
          set((state) => {
            state.wsClient = wsClient;
            state.connectionStatus = CONNECTION_STATUS.CONNECTING;
          });
          
          wsClient.connect().catch((_error) => {
            set((state) => {
              state.connectionStatus = CONNECTION_STATUS.ERROR;
              state.errorMessage = 'Failed to connect to server';
            });
          });
        },
        
        disconnect: () => {
          const state = get();
          if (state.wsClient) {
            state.wsClient.disconnect();
          }
          
          set((state) => {
            state.wsClient = null;
            state.isConnected = false;
            state.connectionStatus = CONNECTION_STATUS.DISCONNECTED;
            state.wsUrl = null;
          });
        },
        
        // Game actions
        selectCard: (card: Card) => {
          set((state) => {
            const index = state.selectedCards.findIndex(c => 
              c.suit === card.suit && c.rank === card.rank
            );
            
            if (index === -1) {
              state.selectedCards.push(card);
            }
          });
        },
        
        deselectCard: (card: Card) => {
          set((state) => {
            state.selectedCards = state.selectedCards.filter(c => 
              !(c.suit === card.suit && c.rank === card.rank)
            );
          });
        },
        
        playCards: (cards: Card[]) => {
          const state = get();
          if (!state.wsClient || !state.isConnected) {
            return;
          }
          
          // Validate play
          const validation = validatePlay(cards, state);
          if (!validation.isValid) {
            set((draft) => {
              draft.errorMessage = validation.error || 'Invalid play';
              draft.showValidationError = true;
            });
            return;
          }
          
          // Convert cards to string format
          const cardStrings = cards.map(card => formatCardForServer(card));
          
          state.wsClient.send({
            t: 'PlayCards',
            data: { cards: cardStrings }
          }).catch((_error) => {
            set((draft) => {
              draft.errorMessage = 'Failed to send play';
              draft.showValidationError = true;
            });
          });
          
          // Clear selected cards
          set((state) => {
            state.selectedCards = [];
          });
        },
        
        pass: () => {
          const state = get();
          if (!state.wsClient || !state.isConnected) {
            return;
          }
          
          state.wsClient.send({
            t: 'Pass'
          }).catch((_error) => {
            set((draft) => {
              draft.errorMessage = 'Failed to pass';
              draft.showValidationError = true;
            });
          });
        },
        
        // State management
        handleSnapshot: (snapshot: GameState) => {
          set((state) => {
            // Update game state
            state.matchId = snapshot.matchId;
            state.players = snapshot.players;
            state.currentDeal = snapshot.currentDeal;
            state.status = snapshot.status;
            state.version = snapshot.version;
            
            // Update player state
            if (state.mySeat && snapshot.currentDeal) {
              state.myHand = snapshot.currentDeal.playerHands[state.mySeat] || [];
            }
            
            // Update UI state
            state.isMyTurn = snapshot.currentDeal?.currentTurn === state.mySeat;
            state.canPlay = state.isMyTurn && state.status === 'playing';
            
            // Clear any errors
            state.showValidationError = false;
            state.errorMessage = '';
          });
        },
        
        handleEvent: (event: EventMessage) => {
          set((state) => {
            // Check version for synchronization
            if (event.version !== state.version + 1) {
              // Version mismatch, request snapshot
              console.warn('Version mismatch, requesting snapshot');
              return;
            }
            
            // Debug log for event structure
            console.log('=== EVENT DEBUG START ===');
            console.log('event.e:', JSON.stringify(event.e));
            console.log('typeof event.e:', typeof event.e);
            console.log('event.e === "MatchCreated":', event.e === "MatchCreated");
            console.log('Full event object:', JSON.stringify(event, null, 2));
            console.log('=== EVENT DEBUG END ===');
            
            // Update version
            state.version = event.version;
            
            // Handle different event types
            switch (event.e) {
              case 'MatchCreated':
                console.log('Handling MatchCreated event');
                handleMatchCreatedEvent(state, event.data);
                break;
              case 'CardsPlayed':
                console.log('Handling CardsPlayed event');
                handleCardsPlayedEvent(state, event.data);
                break;
              case 'PlayerPassed':
                console.log('Handling PlayerPassed event');
                handlePlayerPassedEvent(state, event.data);
                break;
              case 'TrickWon':
                console.log('Handling TrickWon event');
                handleTrickWonEvent(state, event.data);
                break;
              case 'DealStarted':
                console.log('Handling DealStarted event');
                handleDealStartedEvent(state, event.data);
                break;
              case 'CardsDealt':
                console.log('Handling CardsDealt event');
                handleCardsDealtEvent(state, event.data);
                break;
              default:
                console.log('Unknown event type:', event.e);
                console.log('Event object keys:', Object.keys(event));
            }
          });
        },
        
        setError: (message: string) => {
          set((state) => {
            state.errorMessage = message;
            state.showValidationError = true;
          });
        },
        
        clearError: () => {
          set((state) => {
            state.errorMessage = '';
            state.showValidationError = false;
          });
        }
      }
    }))
  )
);

// Helper functions
function validatePlay(cards: Card[], state: any): { isValid: boolean; error?: string } {
  if (!state.isMyTurn) {
    return { isValid: false, error: 'Not your turn' };
  }
  
  if (cards.length === 0) {
    return { isValid: false, error: 'No cards selected' };
  }
  
  if (state.status !== 'playing') {
    return { isValid: false, error: 'Game not in progress' };
  }
  
  // Check if player has all selected cards
  for (const card of cards) {
    const hasCard = state.myHand.some((handCard: any) => 
      handCard.suit === card.suit && handCard.rank === card.rank
    );
    if (!hasCard) {
      return { isValid: false, error: 'You don\'t have this card' };
    }
  }
  
  // TODO: Add more validation rules based on game logic
  return { isValid: true };
}

function formatCardForServer(card: Card): string {
  if (card.suit === 'Joker') {
    return card.rank;
  }
  
  const suitMap: Record<string, string> = {
    'Hearts': '♥',
    'Diamonds': '♦',
    'Clubs': '♣',
    'Spades': '♠'
  };
  
  return suitMap[card.suit] + card.rank;
}

function parseCardFromServer(cardData: any): Card {
  // Handle card objects from server (JSON format)
  if (typeof cardData === 'object' && cardData !== null && 'Suit' in cardData && 'Rank' in cardData) {
    const { Suit, Rank } = cardData;
    
    // Map suit numbers to strings
    const suitMap: Record<number, Card['suit']> = {
      0: 'Hearts',
      1: 'Diamonds', 
      2: 'Clubs',
      3: 'Spades',
      4: 'Joker'
    };
    
    const suit = suitMap[Suit] || 'Hearts';
    
    // Map rank numbers to strings
    if (suit === 'Joker') {
      return { suit: 'Joker', rank: Rank === 51 ? '小王' : '大王' };
    }
    
    // Map rank numbers to display strings
    const rankMap: Record<number, string> = {
      1: '2', 2: '3', 3: '4', 4: '5', 5: '6', 6: '7', 7: '8', 8: '9', 9: '10',
      10: 'J', 11: 'Q', 12: 'K', 13: 'A'
    };
    
    const rank = rankMap[Rank] || Rank.toString();
    return { suit, rank };
  }
  
  // Handle string format (fallback)
  if (typeof cardData === 'string') {
    const cardStr = cardData;
    if (cardStr === '小王') {
      return { suit: 'Joker', rank: '小王' };
    }
    if (cardStr === '大王') {
      return { suit: 'Joker', rank: '大王' };
    }
    
    const suitMap: Record<string, string> = {
      '♥': 'Hearts',
      '♦': 'Diamonds',
      '♣': 'Clubs',
      '♠': 'Spades'
    };
    
    const suit = suitMap[cardStr[0]] || 'Hearts';
    const rank = cardStr.slice(1);
    
    return { suit: suit as Card['suit'], rank };
  }
  
  // Fallback for invalid input
  return { suit: 'Hearts', rank: '2' };
}

// Event handlers
function handleCardsPlayedEvent(state: any, data: any) {
  const { Player, Cards } = data;
  const seatID = convertSeatIDToString(Player);
  
  // Update current deal if it exists
  if (state.currentDeal) {
    // Update table play
    state.currentDeal.tablePlay = {
      cards: Cards.map(parseCardFromServer),
      type: 'unknown', // TODO: get from server
      value: 0 // TODO: get from server
    };
    
    // Update last player
    state.currentDeal.lastPlayer = seatID;
    
    // Remove cards from player's hand if it's our turn
    if (seatID === state.mySeat) {
      const playedCards = Cards.map(parseCardFromServer);
      state.myHand = state.myHand.filter((handCard: any) => {
        return !playedCards.some((playedCard: any) => 
          playedCard.suit === handCard.suit && playedCard.rank === handCard.rank
        );
      });
    }
    
    // Update hand count for the player
    const playerIndex = state.players.findIndex((p: any) => p.seat === seatID);
    if (playerIndex !== -1 && state.players[playerIndex]) {
      state.players[playerIndex].handCount -= Cards.length;
    }
  }
  
  // Update turn (simplified logic)
  const currentPlayerIndex = state.players.findIndex((p: any) => p.seat === seatID);
  if (currentPlayerIndex !== -1 && state.players.length > 0) {
    const nextPlayerIndex = (currentPlayerIndex + 1) % state.players.length;
    const nextPlayer = state.players[nextPlayerIndex];
    if (state.currentDeal && nextPlayer && nextPlayer.seat) {
      state.currentDeal.currentTurn = nextPlayer.seat;
    }
  }
  
  // Update UI state
  state.isMyTurn = state.currentDeal?.currentTurn === state.mySeat;
  state.canPlay = state.isMyTurn && state.status === 'playing';
}

function handlePlayerPassedEvent(state: any, data: any) {
  const { Player } = data;
  const seatID = convertSeatIDToString(Player);
  
  // Update turn to next player
  const currentPlayerIndex = state.players.findIndex((p: any) => p.seat === seatID);
  if (currentPlayerIndex !== -1 && state.players.length > 0) {
    const nextPlayerIndex = (currentPlayerIndex + 1) % state.players.length;
    const nextPlayer = state.players[nextPlayerIndex];
    if (state.currentDeal && nextPlayer && nextPlayer.seat) {
      state.currentDeal.currentTurn = nextPlayer.seat;
    }
  }
  
  // Update UI state
  state.isMyTurn = state.currentDeal?.currentTurn === state.mySeat;
  state.canPlay = state.isMyTurn && state.status === 'playing';
}

function handleTrickWonEvent(state: any, data: any) {
  const { Winner } = data;
  const seatID = convertSeatIDToString(Winner);
  
  // Clear table play
  if (state.currentDeal) {
    state.currentDeal.tablePlay = undefined;
    state.currentDeal.currentTurn = seatID;
  }
  
  // Update UI state
  state.isMyTurn = state.currentDeal?.currentTurn === state.mySeat;
  state.canPlay = state.isMyTurn && state.status === 'playing';
}

function handleDealStartedEvent(state: any, data: any) {
  const { DealNumber, Trump, FirstPlayer } = data;
  const seatID = convertSeatIDToString(FirstPlayer);
  
  // Update deal state
  if (state.currentDeal) {
    state.currentDeal.dealId = `deal_${DealNumber}`;
    state.currentDeal.trump = Trump;
    state.currentDeal.currentTurn = seatID;
  }
  
  state.status = 'playing';
  
  // Update UI state
  state.isMyTurn = state.currentDeal?.currentTurn === state.mySeat;
  state.canPlay = state.isMyTurn && state.status === 'playing';
}

function handleMatchCreatedEvent(state: any, data: any) {
  // MatchCreated event initializes the match
  console.log('handleMatchCreatedEvent called with data:', data);
  console.log('Data type:', typeof data);
  console.log('Data keys:', Object.keys(data || {}));
  
  const { Players } = data;
  console.log('Players:', Players);
  
  if (Players && Array.isArray(Players)) {
    // Update players list from the MatchCreated event
    console.log('Processing players array, length:', Players.length);
    state.players = Players.map((player: any, index: number) => {
      console.log(`Processing player ${index}:`, player);
      console.log(`Player SeatID before conversion:`, player.SeatID);
      
      if (player.SeatID === undefined || player.SeatID === null) {
        console.error(`Player ${index} has invalid SeatID:`, player.SeatID);
        return null;
      }
      
      const convertedSeat = convertSeatIDToString(player.SeatID);
      console.log(`Player seat after conversion:`, convertedSeat);
      
      return {
        id: player.ID || `player_${index}`,
        name: player.Name || `Player ${index}`,
        seat: convertedSeat,
        handCount: 0,
        level: player.Level || 0,
        connected: player.IsOnline !== false
      };
    }).filter(player => player !== null);
  } else {
    console.error('Players is not a valid array:', Players);
    state.players = [];
  }
  
  state.status = 'playing';
  console.log('Match created with players:', state.players);
}

function handleCardsDealtEvent(state: any, data: any) {
  const { Hands } = data;
  
  // Ensure Hands exists before accessing it
  if (!Hands) {
    console.warn('CardsDealt event received without Hands data');
    return;
  }
  
  // Convert Hands keys from integer seat IDs to string seat IDs
  const convertedHands: Record<string, any> = {};
  for (const [seatID, cards] of Object.entries(Hands)) {
    const stringSeatID = convertSeatIDToString(parseInt(seatID));
    convertedHands[stringSeatID] = cards;
  }
  
  // Update player hands
  if (state.mySeat && convertedHands[state.mySeat]) {
    state.myHand = convertedHands[state.mySeat].map(parseCardFromServer);
  }
  
  // Update player hand counts
  state.players.forEach((player: any) => {
    const hand = convertedHands[player.seat];
    if (hand) {
      player.handCount = hand.length;
    }
  });
}

// Selector hooks for common state slices
export const useConnectionState = () => useRoomStore(state => ({
  isConnected: state.isConnected,
  connectionStatus: state.connectionStatus,
  connect: state.actions.connect,
  disconnect: state.actions.disconnect
}));

export const useGameState = () => useRoomStore(state => ({
  matchId: state.matchId,
  players: state.players,
  currentDeal: state.currentDeal,
  status: state.status,
  version: state.version
}));

export const usePlayerState = () => useRoomStore(state => ({
  mySeat: state.mySeat,
  myHand: state.myHand,
  selectedCards: state.selectedCards,
  isMyTurn: state.isMyTurn,
  canPlay: state.canPlay
}));

export const useGameActions = () => useRoomStore(state => ({
  selectCard: state.actions.selectCard,
  deselectCard: state.actions.deselectCard,
  playCards: state.actions.playCards,
  pass: state.actions.pass
}));

export const useErrorState = () => useRoomStore(state => ({
  showValidationError: state.showValidationError,
  errorMessage: state.errorMessage,
  setError: state.actions.setError,
  clearError: state.actions.clearError
}));

// Initialize store with seat information
export const initializePlayerSeat = (seat: SeatID) => {
  useRoomStore.setState((state) => {
    state.mySeat = seat;
  });
};