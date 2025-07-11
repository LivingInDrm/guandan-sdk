package room

import (
	"testing"
	"time"

	"guandan/sdk/domain"
)

func TestPlayerConn_NewPlayerConn(t *testing.T) {
	playerID := "test-player"
	seat := domain.SeatEast
	
	// Create a nil connection for testing
	pc := &PlayerConn{
		PlayerID:  playerID,
		Seat:      seat,
		Conn:      nil,
		LastPing:  time.Now(),
		Connected: true,
	}
	
	if pc.PlayerID != playerID {
		t.Errorf("Expected player ID to be %s, got %s", playerID, pc.PlayerID)
	}
	
	if pc.Seat != seat {
		t.Errorf("Expected seat to be %s, got %s", seat, pc.Seat)
	}
	
	if !pc.Connected {
		t.Error("Expected player to be connected")
	}
}

func TestPlayerConn_IsConnected(t *testing.T) {
	pc := &PlayerConn{
		PlayerID:  "test-player",
		Seat:      domain.SeatEast,
		Conn:      nil,
		LastPing:  time.Now(),
		Connected: true,
	}
	
	if !pc.IsConnected() {
		t.Error("Expected player to be connected")
	}
	
	pc.Connected = false
	
	if pc.IsConnected() {
		t.Error("Expected player to not be connected")
	}
}

func TestPlayerConn_Close(t *testing.T) {
	pc := &PlayerConn{
		PlayerID:  "test-player",
		Seat:      domain.SeatEast,
		Conn:      nil,
		LastPing:  time.Now(),
		Connected: true,
	}
	
	pc.Close()
	
	if pc.IsConnected() {
		t.Error("Expected player to be disconnected after close")
	}
}

func TestRoomError_Error(t *testing.T) {
	err := RoomError{
		Code:    "TEST_ERROR",
		Message: "This is a test error",
	}
	
	if err.Error() != "This is a test error" {
		t.Errorf("Expected error message to be 'This is a test error', got %s", err.Error())
	}
}

func TestDefaultRoomConfig(t *testing.T) {
	config := DefaultRoomConfig
	
	if config.MaxPlayers != 4 {
		t.Errorf("Expected max players to be 4, got %d", config.MaxPlayers)
	}
	
	if config.IdleTimeout != 30*time.Minute {
		t.Errorf("Expected idle timeout to be 30 minutes, got %v", config.IdleTimeout)
	}
	
	if config.PingInterval != 30*time.Second {
		t.Errorf("Expected ping interval to be 30 seconds, got %v", config.PingInterval)
	}
	
	if !config.AllowReconnect {
		t.Error("Expected allow reconnect to be true")
	}
}

func TestPlayerInfo_Structure(t *testing.T) {
	playerInfo := PlayerInfo{
		ID:        "player1",
		Name:      "Test Player",
		Seat:      domain.SeatEast,
		HandCount: 27,
		Level:     5,
		Connected: true,
	}
	
	if playerInfo.ID != "player1" {
		t.Errorf("Expected ID to be 'player1', got %s", playerInfo.ID)
	}
	
	if playerInfo.Name != "Test Player" {
		t.Errorf("Expected name to be 'Test Player', got %s", playerInfo.Name)
	}
	
	if playerInfo.Seat != domain.SeatEast {
		t.Errorf("Expected seat to be %s, got %s", domain.SeatEast, playerInfo.Seat)
	}
	
	if playerInfo.HandCount != 27 {
		t.Errorf("Expected hand count to be 27, got %d", playerInfo.HandCount)
	}
	
	if playerInfo.Level != 5 {
		t.Errorf("Expected level to be 5, got %d", playerInfo.Level)
	}
	
	if !playerInfo.Connected {
		t.Error("Expected player to be connected")
	}
}

func TestMatchSnapshot_Structure(t *testing.T) {
	snapshot := MatchSnapshot{
		MatchID: "test-match",
		Players: []PlayerInfo{
			{
				ID:        "player1",
				Name:      "Test Player 1",
				Seat:      domain.SeatEast,
				HandCount: 27,
				Level:     5,
				Connected: true,
			},
		},
		Status:  "playing",
		Version: 1,
	}
	
	if snapshot.MatchID != "test-match" {
		t.Errorf("Expected match ID to be 'test-match', got %s", snapshot.MatchID)
	}
	
	if len(snapshot.Players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(snapshot.Players))
	}
	
	if snapshot.Status != "playing" {
		t.Errorf("Expected status to be 'playing', got %s", snapshot.Status)
	}
	
	if snapshot.Version != 1 {
		t.Errorf("Expected version to be 1, got %d", snapshot.Version)
	}
}

func TestWSMessage_Structure(t *testing.T) {
	msg := WSMessage{
		Type: "PlayCards",
		Data: map[string]interface{}{
			"cards": []string{"♠A", "♥K"},
		},
	}
	
	if msg.Type != "PlayCards" {
		t.Errorf("Expected type to be 'PlayCards', got %s", msg.Type)
	}
	
	if msg.Data == nil {
		t.Error("Expected data to not be nil")
	}
	
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		t.Error("Expected data to be a map")
	}
	
	cards, ok := data["cards"].([]string)
	if !ok {
		t.Error("Expected cards to be a string slice")
	}
	
	if len(cards) != 2 {
		t.Errorf("Expected 2 cards, got %d", len(cards))
	}
}

func TestSnapshotMessage_Structure(t *testing.T) {
	msg := SnapshotMessage{
		Type:    "Snapshot",
		Version: 42,
		Payload: map[string]interface{}{
			"matchId": "test-match",
		},
	}
	
	if msg.Type != "Snapshot" {
		t.Errorf("Expected type to be 'Snapshot', got %s", msg.Type)
	}
	
	if msg.Version != 42 {
		t.Errorf("Expected version to be 42, got %d", msg.Version)
	}
	
	if msg.Payload == nil {
		t.Error("Expected payload to not be nil")
	}
}

func TestEventMessage_Structure(t *testing.T) {
	msg := EventMessage{
		Type:    "Event",
		Event:   "CardsPlayed",
		Data:    map[string]interface{}{"seat": "east"},
		Version: 1,
	}
	
	if msg.Type != "Event" {
		t.Errorf("Expected type to be 'Event', got %s", msg.Type)
	}
	
	if msg.Event != "CardsPlayed" {
		t.Errorf("Expected event to be 'CardsPlayed', got %s", msg.Event)
	}
	
	if msg.Version != 1 {
		t.Errorf("Expected version to be 1, got %d", msg.Version)
	}
	
	if msg.Data == nil {
		t.Error("Expected data to not be nil")
	}
}