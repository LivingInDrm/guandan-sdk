# Guandan Game SDK

A complete implementation of the Guandan card game SDK with 4-layer architecture, following domain-driven design principles.

## Architecture

The SDK follows a 4-layer architecture:

- **Domain Layer** (`sdk/domain/`): Core game logic, card definitions, and business rules
- **Event Layer** (`sdk/event/`): Event-driven architecture with domain events and event bus
- **Engine Layer** (`sdk/engine/`): Game state machine and game engine
- **Service Layer** (`sdk/service/`): High-level game service interface and match management

## Features

- ✅ Complete Guandan game rules implementation
- ✅ Thread-safe concurrent game management
- ✅ Event-driven architecture with publish/subscribe pattern
- ✅ Deterministic replay with seeded random number generation
- ✅ Game state snapshots and serialization
- ✅ Comprehensive unit and integration tests
- ✅ HTTP + WebSocket demo server

## Getting Started

### Prerequisites

- Go 1.22 or later

### Installation

```bash
git clone https://github.com/LivingInDrm/guandan-sdk.git
cd guandan-sdk
go mod tidy
```

### Running Tests

```bash
# Run all tests
go test ./sdk/...

# Run specific layer tests
go test ./sdk/domain/...
go test ./sdk/engine/...
go test ./sdk/service/...

# Run with verbose output
go test -v ./sdk/...
```

### Demo Server

Start the demo server with HTTP + WebSocket support:

```bash
go run cmd/guandan-server/main.go
```

The server provides:
- WebSocket endpoint: `ws://localhost:8080/ws`
- Web interface: `http://localhost:8080/`
- Health check: `http://localhost:8080/api/health`
- REST API: `http://localhost:8080/api/matches`

## Usage Examples

### Basic Game Service Usage

```go
package main

import (
    "fmt"
    "guandan/sdk/domain"
    "guandan/sdk/service"
)

func main() {
    // Create game service
    gameService := service.NewGameService()
    
    // Create players
    players := []*domain.Player{
        domain.NewPlayer("p1", "Player1", domain.SeatEast),
        domain.NewPlayer("p2", "Player2", domain.SeatSouth),
        domain.NewPlayer("p3", "Player3", domain.SeatWest),
        domain.NewPlayer("p4", "Player4", domain.SeatNorth),
    }
    
    // Create match
    options := &service.MatchOptions{
        DealLimit: 0,
        Seed:      12345,
    }
    
    matchID, err := gameService.CreateMatch(players, options)
    if err != nil {
        panic(err)
    }
    
    // Start a deal
    err = gameService.StartNextDeal(matchID)
    if err != nil {
        panic(err)
    }
    
    // Get current player
    currentPlayer, err := gameService.GetCurrentPlayer(matchID)
    if err != nil {
        panic(err)
    }
    
    // Get valid plays
    validPlays, err := gameService.GetValidPlays(matchID, currentPlayer)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Current player: %s\n", currentPlayer)
    fmt.Printf("Valid plays: %d\n", len(validPlays))
}
```

### Event Subscription

```go
// Subscribe to game events
unsubscribe, err := gameService.Subscribe(matchID, func(event event.DomainEvent) {
    fmt.Printf("Event: %s\n", event.EventType())
})
if err != nil {
    panic(err)
}
defer unsubscribe()
```

### Game State Snapshot

```go
// Get game state snapshot
snapshot, err := gameService.GetSnapshot(matchID)
if err != nil {
    panic(err)
}

// Serialize to JSON
data, err := json.Marshal(snapshot)
if err != nil {
    panic(err)
}

fmt.Printf("Snapshot: %s\n", string(data))
```

## WebSocket API

### Message Format

**Client to Server:**
```json
{
  "type": "create_match",
  "match_id": "optional",
  "player_id": "optional",
  "data": {},
  "timestamp": "2024-01-01T00:00:00Z"
}
```

**Server to Client:**
```json
{
  "type": "create_match_response",
  "match_id": "match_12345",
  "data": {},
  "error": "optional error message",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Supported Message Types

- `create_match`: Create a new game match
- `join_match`: Join an existing match
- `start_deal`: Start a new deal
- `play_cards`: Play cards
- `pass`: Pass turn
- `get_game_state`: Get current game state
- `get_valid_plays`: Get valid plays for a player

## Game Rules

The SDK implements the standard Guandan rules:

- **Players**: 4 players in 2 teams (East-West vs South-North)
- **Cards**: 108 cards (2 standard decks + 4 jokers)
- **Trump**: Changes each deal based on current level
- **Card Types**: Single, Pair, Triple, Straight, Pair Straight, Triple Straight, Bomb, Joker Bomb
- **Winning**: First team to finish with Ace level wins

## Development

### Project Structure

```
guandan/
├── sdk/
│   ├── domain/         # Core game logic
│   ├── event/          # Event system
│   ├── engine/         # Game engine
│   └── service/        # Service layer
├── cmd/
│   └── guandan-server/ # Demo server
├── go.mod
└── README.md
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Testing

The project includes comprehensive tests:

- Unit tests for domain logic
- Integration tests for engine layer
- Service layer tests with concurrency
- All tests pass with 100% coverage of critical paths

## License

This project is licensed under the MIT License.

## Acknowledgments

- Guandan game rules and regulations
- Go community for excellent libraries
- Domain-driven design principles