package room

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"guandan/sdk/domain"
)

// PlayerConn represents a player connection in a room
type PlayerConn struct {
	PlayerID   string
	Seat       domain.SeatID
	Conn       *websocket.Conn
	LastPing   time.Time
	Connected  bool
	mutex      sync.RWMutex
}

// NewPlayerConn creates a new player connection
func NewPlayerConn(playerID string, seat domain.SeatID, conn *websocket.Conn) *PlayerConn {
	return &PlayerConn{
		PlayerID:  playerID,
		Seat:      seat,
		Conn:      conn,
		LastPing:  time.Now(),
		Connected: true,
	}
}

// Send sends a message to the player
func (pc *PlayerConn) Send(msg interface{}) error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	if !pc.Connected {
		return ErrPlayerDisconnected
	}
	
	return pc.Conn.WriteJSON(msg)
}

// Close closes the connection
func (pc *PlayerConn) Close() {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	pc.Connected = false
	if pc.Conn != nil {
		pc.Conn.Close()
	}
}

// IsConnected checks if the player is connected
func (pc *PlayerConn) IsConnected() bool {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	return pc.Connected
}

// Message types for WebSocket communication
type WSMessage struct {
	Type string      `json:"t"`
	Data interface{} `json:"data,omitempty"`
}

// Client to server messages
type PlayCardsMessage struct {
	Cards []string `json:"cards"`
}

type PassMessage struct{}

// Server to client messages
type SnapshotMessage struct {
	Type    string      `json:"t"`
	Version int         `json:"version"`
	Payload interface{} `json:"payload"`
}

type EventMessage struct {
	Type    string      `json:"t"`
	Event   string      `json:"e"`
	Data    interface{} `json:"data"`
	Version int         `json:"version"`
}

// Match snapshot for synchronization
type MatchSnapshot struct {
	MatchID      string                     `json:"matchId"`
	Players      []PlayerInfo               `json:"players"`
	CurrentDeal  *DealSnapshot              `json:"currentDeal"`
	Status       string                     `json:"status"`
	Version      int                        `json:"version"`
}

type PlayerInfo struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Seat      domain.SeatID `json:"seat"`
	HandCount int           `json:"handCount"`
	Level     int           `json:"level"`
	Connected bool          `json:"connected"`
}

type DealSnapshot struct {
	DealID       string                `json:"dealId"`
	Trump        domain.Rank           `json:"trump"`
	Phase        string                `json:"phase"`
	CurrentTurn  domain.SeatID         `json:"currentTurn"`
	TablePlay    *domain.CardGroup     `json:"tablePlay"`
	LastPlayer   domain.SeatID         `json:"lastPlayer"`
	TrickHistory []TrickInfo           `json:"trickHistory"`
	PlayerHands  map[domain.SeatID][]domain.Card `json:"playerHands"`
}

type TrickInfo struct {
	Winner domain.SeatID        `json:"winner"`
	Cards  []domain.Card        `json:"cards"`
	Player domain.SeatID        `json:"player"`
}

// Room configuration
type RoomConfig struct {
	MaxPlayers    int           `json:"maxPlayers"`
	IdleTimeout   time.Duration `json:"idleTimeout"`
	PingInterval  time.Duration `json:"pingInterval"`
	AllowReconnect bool         `json:"allowReconnect"`
}

// Default room configuration
var DefaultRoomConfig = RoomConfig{
	MaxPlayers:    4,
	IdleTimeout:   30 * time.Minute,
	PingInterval:  30 * time.Second,
	AllowReconnect: true,
}

// Room events
type RoomEvent struct {
	Type      string      `json:"type"`
	RoomID    string      `json:"roomId"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// Error types
type RoomError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e RoomError) Error() string {
	return e.Message
}

// Common errors
var (
	ErrRoomFull            = RoomError{"ROOM_FULL", "Room is full"}
	ErrRoomNotFound        = RoomError{"ROOM_NOT_FOUND", "Room not found"}
	ErrPlayerNotFound      = RoomError{"PLAYER_NOT_FOUND", "Player not found"}
	ErrPlayerDisconnected  = RoomError{"PLAYER_DISCONNECTED", "Player disconnected"}
	ErrInvalidSeat         = RoomError{"INVALID_SEAT", "Invalid seat"}
	ErrSeatTaken          = RoomError{"SEAT_TAKEN", "Seat is already taken"}
	ErrGameNotStarted     = RoomError{"GAME_NOT_STARTED", "Game not started"}
	ErrGameAlreadyStarted = RoomError{"GAME_ALREADY_STARTED", "Game already started"}
	ErrInvalidAction      = RoomError{"INVALID_ACTION", "Invalid action"}
	ErrNotPlayerTurn      = RoomError{"NOT_PLAYER_TURN", "Not player's turn"}
)