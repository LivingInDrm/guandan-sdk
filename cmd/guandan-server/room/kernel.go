package room

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"guandan/sdk/domain"
	"guandan/sdk/event"
	"guandan/sdk/service"
)

// RoomKernel manages a single room with multiple players
type RoomKernel struct {
	roomID       string
	gameService  service.GameService
	matchID      domain.MatchID
	players      map[domain.SeatID]*PlayerConn
	config       RoomConfig
	version      int
	mutex        sync.RWMutex
	eventSub     func() // unsubscribe function
	ctx          context.Context
	cancel       context.CancelFunc
	lastActivity time.Time
}

// NewRoomKernel creates a new room kernel
func NewRoomKernel(roomID string, gameService service.GameService, config RoomConfig) *RoomKernel {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &RoomKernel{
		roomID:       roomID,
		gameService:  gameService,
		players:      make(map[domain.SeatID]*PlayerConn),
		config:       config,
		version:      1,
		ctx:          ctx,
		cancel:       cancel,
		lastActivity: time.Now(),
	}
}

// Start starts the room kernel
func (rk *RoomKernel) Start() error {
	// Start ping routine
	go rk.pingRoutine()
	
	// Start idle timeout routine
	go rk.idleTimeoutRoutine()
	
	return nil
}

// Stop stops the room kernel
func (rk *RoomKernel) Stop() {
	rk.cancel()
	
	rk.mutex.Lock()
	defer rk.mutex.Unlock()
	
	// Close all connections
	for _, player := range rk.players {
		player.Close()
	}
	
	// Unsubscribe from events
	if rk.eventSub != nil {
		rk.eventSub()
	}
}

// AddPlayer adds a player to the room
func (rk *RoomKernel) AddPlayer(playerID string, seat domain.SeatID, conn *websocket.Conn) error {
	rk.mutex.Lock()
	defer rk.mutex.Unlock()
	
	// Check if room is full
	if len(rk.players) >= rk.config.MaxPlayers {
		return ErrRoomFull
	}
	
	// Check if seat is taken
	if _, exists := rk.players[seat]; exists {
		return ErrSeatTaken
	}
	
	// Create player connection
	playerConn := NewPlayerConn(playerID, seat, conn)
	rk.players[seat] = playerConn
	
	// Update activity
	rk.lastActivity = time.Now()
	
	// If we have all players, create the match
	log.Printf("Room %s now has %d/%d players", rk.roomID, len(rk.players), rk.config.MaxPlayers)
	if len(rk.players) == rk.config.MaxPlayers {
		log.Printf("Room %s is full, creating match...", rk.roomID)
		err := rk.createMatch()
		if err != nil {
			log.Printf("Failed to create match for room %s: %v", rk.roomID, err)
			return err
		}
		log.Printf("Match created successfully for room %s", rk.roomID)
	}
	
	// Send current state to new player
	go rk.sendSnapshotToPlayer(playerConn)
	
	log.Printf("Player %s joined room %s at seat %s", playerID, rk.roomID, seat)
	return nil
}

// RemovePlayer removes a player from the room
func (rk *RoomKernel) RemovePlayer(seat domain.SeatID) {
	rk.mutex.Lock()
	defer rk.mutex.Unlock()
	
	if player, exists := rk.players[seat]; exists {
		player.Close()
		delete(rk.players, seat)
		log.Printf("Player %s left room %s", player.PlayerID, rk.roomID)
	}
}

// HandleMessage handles a message from a player
func (rk *RoomKernel) HandleMessage(seat domain.SeatID, msg WSMessage) {
	rk.mutex.Lock()
	defer rk.mutex.Unlock()
	
	// Update activity
	rk.lastActivity = time.Now()
	
	// Check if player exists
	player, exists := rk.players[seat]
	if !exists {
		return
	}
	
	// Handle different message types
	switch msg.Type {
	case "PlayCards":
		rk.handlePlayCards(player, msg)
	case "Pass":
		rk.handlePass(player, msg)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// GetSnapshot returns the current game state snapshot
func (rk *RoomKernel) GetSnapshot() (*MatchSnapshot, error) {
	rk.mutex.RLock()
	defer rk.mutex.RUnlock()
	
	if rk.matchID == "" {
		return &MatchSnapshot{
			MatchID: rk.roomID,
			Players: rk.getPlayersInfo(),
			Status:  "waiting",
			Version: rk.version,
		}, nil
	}
	
	// Get match state from game service
	matchState, err := rk.gameService.GetMatchState(rk.matchID)
	if err != nil {
		return nil, err
	}
	
	// Convert to snapshot
	snapshot := &MatchSnapshot{
		MatchID: rk.roomID,
		Players: rk.getPlayersInfo(),
		Status:  "playing",
		Version: rk.version,
	}
	
	if matchState.CurrentDeal > 0 {
		// Get deal context from game service
		serviceSnapshot, err := rk.gameService.GetSnapshot(rk.matchID)
		if err == nil {
			snapshot.CurrentDeal = &DealSnapshot{
				DealID:      fmt.Sprintf("deal-%d", matchState.CurrentDeal),
				Trump:       matchState.Trump,
				Phase:       matchState.Phase.String(),
				CurrentTurn: domain.SeatEast, // Default, would need to get from trick context
				TablePlay:   nil,            // Would need to get from trick context
				LastPlayer:  domain.SeatEast, // Default
				PlayerHands: serviceSnapshot.Hands,
			}
		}
	}
	
	return snapshot, nil
}

// GetPlayerCount returns the number of players in the room
func (rk *RoomKernel) GetPlayerCount() int {
	rk.mutex.RLock()
	defer rk.mutex.RUnlock()
	return len(rk.players)
}

// IsEmpty returns true if the room is empty
func (rk *RoomKernel) IsEmpty() bool {
	return rk.GetPlayerCount() == 0
}

// Private methods

func (rk *RoomKernel) createMatch() error {
	// Create players array
	players := make([]*domain.Player, 0, len(rk.players))
	for seat, playerConn := range rk.players {
		player := domain.NewPlayer(playerConn.PlayerID, playerConn.PlayerID, seat)
		players = append(players, player)
	}
	
	// Create match
	matchID, err := rk.gameService.CreateMatch(players, &service.MatchOptions{
		DealLimit: 0,
		Seed:      time.Now().UnixNano(),
	})
	if err != nil {
		return err
	}
	
	rk.matchID = matchID
	
	// Subscribe to match events
	rk.eventSub, err = rk.gameService.Subscribe(matchID, rk.handleGameEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to match events: %w", err)
	}
	
	// Start first deal
	log.Printf("Starting first deal for match %s", matchID)
	err = rk.gameService.StartNextDeal(matchID)
	if err != nil {
		log.Printf("Failed to start deal: %v", err)
		return err
	}
	log.Printf("First deal started successfully for match %s", matchID)
	
	// Send updated snapshot to all players with dealt cards
	go func() {
		snapshot, err := rk.GetSnapshot()
		if err != nil {
			log.Printf("Failed to get snapshot after deal: %v", err)
			return
		}
		
		snapshotMsg := SnapshotMessage{
			Type:    "Snapshot",
			Version: snapshot.Version,
			Payload: snapshot,
		}
		
		log.Printf("Broadcasting updated snapshot with dealt cards to all players")
		rk.broadcastMessage(snapshotMsg)
	}()
	
	return nil
}

func (rk *RoomKernel) handlePlayCards(player *PlayerConn, msg WSMessage) {
	if rk.matchID == "" {
		return
	}
	
	// Parse cards from message
	var playMsg PlayCardsMessage
	if data, ok := msg.Data.(map[string]interface{}); ok {
		if cardsData, ok := data["cards"].([]interface{}); ok {
			cards := make([]domain.Card, len(cardsData))
			for i, cardStr := range cardsData {
				if str, ok := cardStr.(string); ok {
					card, err := domain.ParseCard(str)
					if err != nil {
						log.Printf("Invalid card: %s", str)
						return
					}
					cards[i] = card
				}
			}
			playMsg.Cards = make([]string, len(cards))
			for i, card := range cards {
				playMsg.Cards[i] = card.String()
			}
		}
	}
	
	// Convert string cards to domain cards
	cards := make([]domain.Card, len(playMsg.Cards))
	for i, cardStr := range playMsg.Cards {
		card, err := domain.ParseCard(cardStr)
		if err != nil {
			log.Printf("Invalid card: %s", cardStr)
			return
		}
		cards[i] = card
	}
	
	// Play cards
	err := rk.gameService.PlayCards(rk.matchID, player.Seat, cards)
	if err != nil {
		log.Printf("Failed to play cards: %v", err)
		// Send error to player
		errorMsg := map[string]interface{}{
			"t":     "Error",
			"error": err.Error(),
		}
		player.Send(errorMsg)
	}
}

func (rk *RoomKernel) handlePass(player *PlayerConn, msg WSMessage) {
	if rk.matchID == "" {
		return
	}
	
	// Pass
	err := rk.gameService.Pass(rk.matchID, player.Seat)
	if err != nil {
		log.Printf("Failed to pass: %v", err)
		// Send error to player
		errorMsg := map[string]interface{}{
			"t":     "Error",
			"error": err.Error(),
		}
		player.Send(errorMsg)
	}
}

func (rk *RoomKernel) handleGameEvent(event event.DomainEvent) {
	// Increment version
	rk.version++
	
	// Create event message
	eventMsg := EventMessage{
		Type:    "Event",
		Event:   event.EventType(),
		Data:    event,
		Version: rk.version,
	}
	
	// Debug log
	log.Printf("Broadcasting event: %s with data: %+v", event.EventType(), event)
	
	// Broadcast to all players
	rk.broadcastMessage(eventMsg)
}

func (rk *RoomKernel) broadcastMessage(msg interface{}) {
	rk.mutex.RLock()
	defer rk.mutex.RUnlock()
	
	for _, player := range rk.players {
		if player.IsConnected() {
			err := player.Send(msg)
			if err != nil {
				log.Printf("Failed to send message to player %s: %v", player.PlayerID, err)
			}
		}
	}
}

func (rk *RoomKernel) sendSnapshotToPlayer(player *PlayerConn) {
	snapshot, err := rk.GetSnapshot()
	if err != nil {
		log.Printf("Failed to get snapshot: %v", err)
		return
	}
	
	snapshotMsg := SnapshotMessage{
		Type:    "Snapshot",
		Version: snapshot.Version,
		Payload: snapshot,
	}
	
	err = player.Send(snapshotMsg)
	if err != nil {
		log.Printf("Failed to send snapshot to player %s: %v", player.PlayerID, err)
	}
}

func (rk *RoomKernel) getPlayersInfo() []PlayerInfo {
	playersInfo := make([]PlayerInfo, 0, len(rk.players))
	
	for seat, player := range rk.players {
		info := PlayerInfo{
			ID:        player.PlayerID,
			Name:      player.PlayerID,
			Seat:      seat,
			Connected: player.IsConnected(),
		}
		
		// Get hand count if match is active
		if rk.matchID != "" {
			matchState, err := rk.gameService.GetMatchState(rk.matchID)
			if err != nil {
				log.Printf("Failed to get match state for hand count: %v", err)
			} else {
				log.Printf("Match state retrieved, %d players found", len(matchState.Players))
				for _, p := range matchState.Players {
					if p.SeatID == seat {
						handSize := len(p.GetHand())
						info.HandCount = handSize
						log.Printf("Player %s has %d cards in hand", seat, handSize)
						break
					}
				}
			}
		}
		
		playersInfo = append(playersInfo, info)
	}
	
	return playersInfo
}

func (rk *RoomKernel) pingRoutine() {
	ticker := time.NewTicker(rk.config.PingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			rk.pingPlayers()
		case <-rk.ctx.Done():
			return
		}
	}
}

func (rk *RoomKernel) pingPlayers() {
	rk.mutex.RLock()
	defer rk.mutex.RUnlock()
	
	for _, player := range rk.players {
		if player.IsConnected() {
			err := player.Send(map[string]interface{}{"t": "ping"})
			if err != nil {
				log.Printf("Failed to ping player %s: %v", player.PlayerID, err)
			}
		}
	}
}

func (rk *RoomKernel) idleTimeoutRoutine() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if time.Since(rk.lastActivity) > rk.config.IdleTimeout {
				log.Printf("Room %s idle timeout", rk.roomID)
				rk.Stop()
				return
			}
		case <-rk.ctx.Done():
			return
		}
	}
}