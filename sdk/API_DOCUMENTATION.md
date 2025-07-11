# Guandan SDK API Documentation

## Overview

The Guandan SDK provides a comprehensive implementation of the Chinese card game "Guandan" (掼蛋). The SDK is structured in 4 layers following clean architecture principles:

1. **Domain Layer** - Core business entities and rules
2. **Engine Layer** - Game engine and state machine
3. **Event Layer** - Event sourcing and messaging
4. **Service Layer** - High-level game service and snapshot management

## Architecture

```
┌─────────────┐
│   Service   │  ← High-level game management
├─────────────┤
│    Event    │  ← Event sourcing & messaging
├─────────────┤
│   Engine    │  ← Game engine & state machine
├─────────────┤
│   Domain    │  ← Core entities & business rules
└─────────────┘
```

---

## Domain Layer (`sdk/domain/`)

The domain layer contains core business entities and game rules for Guandan.

### Core Entities

#### Card (`card.go`)

Represents a playing card in the Guandan game.

**Types:**
```go
type Suit int
const (
    Hearts Suit = iota    // ♥
    Diamonds              // ♦
    Clubs                 // ♣
    Spades                // ♠
    Joker                 // 🃏
)

type Rank int
const (
    Two Rank = iota + 1   // 2-13, SmallJoker, BigJoker
    Three
    // ... up to Ace, SmallJoker, BigJoker
)

type Card struct {
    Suit Suit
    Rank Rank
}
```

**Key Functions:**
- `NewCard(suit, rank)` - Create a new card
- `NewJoker(rank)` - Create a joker card
- `ParseCard(string)` - Parse card from string representation
- `(c Card) IsJoker()` - Check if card is a joker
- `(c Card) IsRedSuit()` - Check if card is red suit (Hearts/Diamonds)
- `(c Card) ID()` - Get unique card ID for comparison

#### CardGroup (`cardgroup.go`)

Represents a valid combination of cards that can be played.

**Types:**
```go
type CardCategory int
const (
    InvalidCategory CardCategory = iota
    Single                  // Single card
    Pair                   // Two of same rank
    Triple                 // Three of same rank
    Straight               // 5+ consecutive cards
    PairStraight          // 3+ consecutive pairs
    TripleStraight        // 2+ consecutive triples
    Bomb                  // Four of same rank
    JokerBomb             // 2+ jokers
)

type CardGroup struct {
    Cards    []Card
    Category CardCategory
    Size     int
    Rank     Rank
}
```

**Key Functions:**
- `NewCardGroup(cards)` - Analyze and create card group
- `(cg *CardGroup) IsValid()` - Check if combination is valid
- `(cg *CardGroup) IsBomb()` - Check if it's a bomb
- `(cg *CardGroup) ComparisonKey()` - Get key for comparing groups

#### Card Comparison (`compare.go`)

Functions for comparing cards and card groups according to Guandan rules.

**Key Functions:**
- `CompareCards(a, b Card, trump Rank)` - Compare two cards
- `CompareCardGroups(a, b *CardGroup, trump Rank)` - Compare card groups
- `CanBeat(hand, tablePlay *CardGroup, trump Rank)` - Check if hand can beat table play
- `CanFollow(hand, tablePlay *CardGroup, trump Rank)` - Check if hand can follow table play
- `GetPlayableCards(hand []Card, tablePlay *CardGroup, trump Rank)` - Get all valid plays

**Trump Card Rules:**
- `IsTrump(card, trump)` - Check if card is trump
- `GetTrumpCards(cards, trump)` - Filter trump cards
- `CountTrumps(cards, trump)` - Count trump cards in hand

#### Player Management (`player.go`)

**Types:**
```go
type SeatID int
const (
    SeatEast SeatID = iota
    SeatSouth
    SeatWest
    SeatNorth
)

type TeamID int
const (
    TeamEastWest TeamID = iota
    TeamSouthNorth
)

type Player struct {
    ID       string
    Name     string
    SeatID   SeatID
    TeamID   TeamID
    Level    Rank
    Hand     []Card
    IsOnline bool
}

type Team struct {
    ID      TeamID
    Players [2]*Player
    Level   Rank
}
```

**Key Functions:**
- `NewPlayer(id, name, seat)` - Create new player
- `(p *Player) AddCards(cards)` - Add cards to hand
- `(p *Player) RemoveCards(cards)` - Remove cards from hand
- `(p *Player) HasCard(card)` - Check if player has card
- `GetTeamFromSeat(seat)` - Get team from seat position

#### Game Context (`context.go`)

Context objects for tracking game state across different levels.

**Types:**
```go
type MatchCtx struct {
    ID          MatchID
    State       MatchState
    Players     PlayerArray
    Teams       [2]*Team
    StartTime   time.Time
    EndTime     *time.Time
    CurrentDeal int
    MaxDeals    int
    Winner      *TeamID
    Seed        int64
}

type DealCtx struct {
    DealNumber    int
    State         DealState
    Trump         Rank
    CurrentLevel  Rank
    StartTime     time.Time
    FirstPlayer   SeatID
    RankList      []SeatID
    TrickCount    int
    IsFirstDeal   bool
    TributeGiven  bool
    TributeCards  map[SeatID][]Card
}

type TrickCtx struct {
    TrickNumber   int
    State         TrickState
    StartPlayer   SeatID
    CurrentPlayer SeatID
    LastPlay      *CardGroup
    LastPlayer    SeatID
    PassedPlayers map[SeatID]bool
    PlayHistory   []TrickPlay
    Winner        SeatID
}
```

#### Deck Management (`deck.go`)

**Type:**
```go
type Deck struct {
    Cards []Card
    rng   *rand.Rand
}
```

**Key Functions:**
- `NewDeck()` - Create new shuffled deck (108 cards: 2×54)
- `NewDeckWithSeed(seed)` - Create deck with specific seed
- `(d *Deck) Shuffle()` - Shuffle the deck
- `(d *Deck) Deal(numCards)` - Deal specified number of cards
- `(d *Deck) DealToHands(numPlayers)` - Deal cards to players

---

## Engine Layer (`sdk/engine/`)

The engine layer manages game flow and state transitions.

### GameEngine (`engine.go`)

The main game engine coordinates all game operations.

**Type:**
```go
type GameEngine struct {
    mu              sync.RWMutex
    stateMachine    *DealStateMachine
    eventBus        *event.EventBus
    isInitialized   bool
    allowedActions  map[domain.SeatID][]string
}
```

**Core Operations:**
- `NewGameEngine(eventBus)` - Create new engine
- `Initialize(matchCtx)` - Initialize engine with match context
- `StartDeal(dealNumber, trump, firstPlayer)` - Start a new deal
- `DealCards()` - Deal cards to all players
- `PlayCards(seat, cards)` - Player plays cards
- `Pass(seat)` - Player passes turn

**Query Functions:**
- `GetCurrentPhase()` - Get current game phase
- `GetCurrentPlayer()` - Get current player's turn
- `GetValidPlays(seat)` - Get valid plays for player
- `CanPlayCards(seat, cards)` - Check if cards can be played
- `IsPlayerTurn(seat)` - Check if it's player's turn
- `GetPlayerHand(seat)` - Get player's current hand

**State Management:**
- `IsGameFinished()` - Check if game is complete
- `GetGameWinner()` - Get winning team
- `Reset()` - Reset engine state

### DealStateMachine (`statemachine.go`)

Manages the state transitions within a single deal.

**Deal Phases:**
```go
type DealPhase int
const (
    PhaseIdle DealPhase = iota
    PhaseCreated
    PhaseCardsDealt
    PhaseTribute
    PhaseFirstPlay
    PhaseInProgress
    PhaseRankList
    PhaseFinished
)
```

**Type:**
```go
type DealStateMachine struct {
    currentPhase DealPhase
    matchCtx     *domain.MatchCtx
    dealCtx      *domain.DealCtx
    trickCtx     *domain.TrickCtx
    eventBus     *event.EventBus
    deck         *domain.Deck
}
```

**State Transitions:**
- `StartDeal()` - Idle → Created
- `DealCards()` - Created → CardsDealt
- `StartTribute()` - CardsDealt → Tribute
- `StartFirstPlay()` - Tribute → FirstPlay
- `TransitionToInProgress()` - FirstPlay → InProgress
- `finishDeal()` - InProgress → Finished

---

## Event Layer (`sdk/event/`)

Event-driven architecture for game state changes and notifications.

### Event System (`events.go`)

**Base Event Interface:**
```go
type DomainEvent interface {
    EventType() string
    Timestamp() time.Time
    MatchID() domain.MatchID
}
```

**Event Types:**
- `MatchCreatedEvent` - Match creation
- `DealStartedEvent` - Deal initialization
- `CardsDealtEvent` - Cards distributed to players
- `TributeRequestedEvent` - Tribute phase started
- `TributeGivenEvent` - Tribute cards exchanged
- `CardsPlayedEvent` - Player played cards
- `PlayerPassedEvent` - Player passed turn
- `TrickWonEvent` - Trick completed
- `PlayerFinishedEvent` - Player finished all cards
- `DealEndedEvent` - Deal completed
- `MatchEndedEvent` - Match completed

### EventBus

**Type:**
```go
type EventBus struct {
    mu           sync.RWMutex
    subscribers  map[domain.MatchID][]chan<- DomainEvent
    globalChan   chan DomainEvent
    bufferSize   int
    isRunning    bool
    stopChan     chan struct{}
}
```

**Key Functions:**
- `NewEventBus(bufferSize)` - Create new event bus
- `Start()` - Start event processing
- `Stop()` - Stop event processing
- `Publish(event)` - Publish event to all subscribers
- `Subscribe(matchID)` - Subscribe to match events
- `SubscribeWithCallback(matchID, callback)` - Subscribe with callback function

---

## Service Layer (`sdk/service/`)

High-level game service interface for application integration.

### GameService (`service.go`)

The main service interface for managing Guandan games.

**Interface:**
```go
type GameService interface {
    CreateMatch(players []*domain.Player, opt *MatchOptions) (domain.MatchID, error)
    StartNextDeal(matchID domain.MatchID) error
    PlayCards(matchID domain.MatchID, seat domain.SeatID, cards []domain.Card) error
    Pass(matchID domain.MatchID, seat domain.SeatID) error
    GetSnapshot(matchID domain.MatchID) (*MatchSnapshot, error)
    Subscribe(matchID domain.MatchID, callback func(event.DomainEvent)) (func(), error)
    GetValidPlays(matchID domain.MatchID, seat domain.SeatID) ([][]domain.Card, error)
    GetCurrentPlayer(matchID domain.MatchID) (domain.SeatID, error)
    IsPlayerTurn(matchID domain.MatchID, seat domain.SeatID) (bool, error)
    GetMatchState(matchID domain.MatchID) (*MatchState, error)
    DeleteMatch(matchID domain.MatchID) error
}
```

**Implementation:**
```go
type GameServiceImpl struct {
    mu       sync.RWMutex
    matches  map[domain.MatchID]*MatchInstance
    eventBus *event.EventBus
    idSeed   int64
}
```

**Usage Example:**
```go
// Create service
gameService := NewGameService()

// Create players
players := []*domain.Player{
    domain.NewPlayer("player1", "Alice", domain.SeatEast),
    domain.NewPlayer("player2", "Bob", domain.SeatSouth),
    domain.NewPlayer("player3", "Charlie", domain.SeatWest),
    domain.NewPlayer("player4", "David", domain.SeatNorth),
}

// Create match
matchID, err := gameService.CreateMatch(players, &MatchOptions{
    DealLimit: 0,
    Seed: 12345,
})

// Start first deal
err = gameService.StartNextDeal(matchID)

// Play cards
cards := []domain.Card{domain.NewCard(domain.Hearts, domain.Ace)}
err = gameService.PlayCards(matchID, domain.SeatEast, cards)

// Subscribe to events
unsubscribe, err := gameService.Subscribe(matchID, func(event event.DomainEvent) {
    fmt.Printf("Event: %s\n", event.EventType())
})
defer unsubscribe()
```

### Snapshot System (`snapshot.go`)

Game state persistence and replay functionality.

**Types:**
```go
type MatchSnapshot struct {
    Version     int                            `json:"version"`
    MatchID     domain.MatchID                 `json:"match_id"`
    MatchCtx    domain.MatchCtx                `json:"match_ctx"`
    DealCtx     domain.DealCtx                 `json:"deal_ctx"`
    TrickCtx    domain.TrickCtx                `json:"trick_ctx"`
    Hands       map[domain.SeatID][]domain.Card `json:"hands"`
    History     []event.DomainEvent            `json:"history"`
    CreatedAt   time.Time                      `json:"created_at"`
    UpdatedAt   time.Time                      `json:"updated_at"`
}

type SnapshotManager struct {
    snapshots map[domain.MatchID]*MatchSnapshot
}

type ReplayManager struct {
    snapshotManager *SnapshotManager
}
```

**Key Functions:**
- `(s *MatchSnapshot) ToJSON()` - Serialize snapshot
- `(s *MatchSnapshot) FromJSON()` - Deserialize snapshot
- `(s *MatchSnapshot) Validate()` - Validate snapshot integrity
- `(sm *SnapshotManager) SaveSnapshot()` - Save game snapshot
- `(sm *SnapshotManager) LoadSnapshot()` - Load game snapshot
- `(rm *ReplayManager) ReplayFromSnapshot()` - Replay game from snapshot

---

## Usage Patterns

### Basic Game Flow

1. **Create Game Service**
   ```go
   gameService := service.NewGameService()
   ```

2. **Create Match**
   ```go
   players := []*domain.Player{...}
   matchID, err := gameService.CreateMatch(players, nil)
   ```

3. **Start Deal**
   ```go
   err = gameService.StartNextDeal(matchID)
   ```

4. **Game Loop**
   ```go
   for !gameFinished {
       currentPlayer, _ := gameService.GetCurrentPlayer(matchID)
       validPlays, _ := gameService.GetValidPlays(matchID, currentPlayer)
       
       // Player chooses cards or passes
       if len(chosenCards) > 0 {
           err = gameService.PlayCards(matchID, currentPlayer, chosenCards)
       } else {
           err = gameService.Pass(matchID, currentPlayer)
       }
   }
   ```

### Event Handling

```go
unsubscribe, err := gameService.Subscribe(matchID, func(event event.DomainEvent) {
    switch e := event.(type) {
    case *event.CardsPlayedEvent:
        fmt.Printf("Player %s played cards\n", e.Player)
    case *event.DealEndedEvent:
        fmt.Printf("Deal ended, winner: %s\n", e.WinnerTeam)
    case *event.MatchEndedEvent:
        fmt.Printf("Match ended, winner: %s\n", e.WinnerTeam)
    }
})
```

### State Inspection

```go
// Get current match state
state, err := gameService.GetMatchState(matchID)
fmt.Printf("Phase: %s, Deal: %d, Trump: %s\n", 
    state.Phase, state.CurrentDeal, state.Trump)

// Get game snapshot
snapshot, err := gameService.GetSnapshot(matchID)
playerHand := snapshot.GetPlayerHand(domain.SeatEast)
```

---

## Error Handling

The SDK uses Go's standard error handling patterns. Common error scenarios:

- **Invalid card combinations** - `CardGroup.IsValid() == false`
- **Invalid plays** - `CanFollow() == false`
- **Turn violations** - Player playing out of turn
- **Match not found** - Invalid match ID
- **Engine not initialized** - Operations before initialization

All service methods return errors that should be checked and handled appropriately.

---

## Thread Safety

- `GameEngine` uses `sync.RWMutex` for concurrent access
- `EventBus` is thread-safe for publishing and subscribing
- `GameService` uses mutexes to protect match instances
- Domain objects are generally immutable with "With" methods returning new instances

---

## Testing

The SDK includes comprehensive tests:
- Unit tests for all domain logic
- Integration tests for engine state transitions
- Event system tests
- Service layer tests

Run tests with:
```bash
go test ./sdk/...
```

---

## Performance Considerations

- **Card Combinations**: Generating all possible plays can be expensive for large hands
- **Event Bus**: Uses buffered channels to prevent blocking
- **Snapshots**: Deep copying can be memory intensive for large game states
- **Concurrent Matches**: Service can handle multiple simultaneous matches

---

## Extension Points

The SDK is designed for extensibility:

1. **Custom Card Rules**: Extend comparison logic in `compare.go`
2. **Additional Events**: Add new event types in `events.go`
3. **Alternative Storage**: Implement snapshot persistence backends
4. **AI Integration**: Use `GetValidPlays()` for AI decision making
5. **Network Protocol**: Build on top of event system for multiplayer

This documentation provides a comprehensive guide to integrating and extending the Guandan SDK for your applications.