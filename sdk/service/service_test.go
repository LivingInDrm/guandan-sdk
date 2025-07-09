package service

import (
	"sync"
	"testing"
	"time"
	"guandan/sdk/domain"
	"guandan/sdk/engine"
	"guandan/sdk/event"
)

func TestGameServiceInitialization(t *testing.T) {
	service := NewGameService()
	
	if service == nil {
		t.Error("GameService should not be nil")
	}
	
	// Test by trying to get a non-existent match
	_, err := service.GetMatchState("nonexistent")
	if err == nil {
		t.Error("Should return error for non-existent match")
	}
}

func TestGameServiceCreateMatch(t *testing.T) {
	service := NewGameService()
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	options := &MatchOptions{
		DealLimit: 0,
		Seed:      12345,
	}
	
	matchID, err := service.CreateMatch(players, options)
	if err != nil {
		t.Errorf("Failed to create match: %v", err)
	}
	
	if matchID == "" {
		t.Error("MatchID should not be empty")
	}
	
	matchState, err := service.GetMatchState(matchID)
	if err != nil {
		t.Errorf("Failed to get match state: %v", err)
	}
	
	if matchState.MatchID != matchID {
		t.Errorf("Expected match ID %s, got %s", matchID, matchState.MatchID)
	}
}

func TestGameServiceStartNextDeal(t *testing.T) {
	service := NewGameService()
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	options := &MatchOptions{
		DealLimit: 0,
		Seed:      12345,
	}
	
	matchID, err := service.CreateMatch(players, options)
	if err != nil {
		t.Errorf("Failed to create match: %v", err)
	}
	
	err = service.StartNextDeal(matchID)
	if err != nil {
		t.Errorf("Failed to start next deal: %v", err)
	}
	
	matchState, err := service.GetMatchState(matchID)
	if err != nil {
		t.Errorf("Failed to get match state: %v", err)
	}
	
	if matchState.Phase != engine.PhaseFirstPlay {
		t.Errorf("Expected phase %s, got %s", engine.PhaseFirstPlay, matchState.Phase)
	}
}

func TestGameServicePlayCards(t *testing.T) {
	service := NewGameService()
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	options := &MatchOptions{
		DealLimit: 0,
		Seed:      12345,
	}
	
	matchID, err := service.CreateMatch(players, options)
	if err != nil {
		t.Errorf("Failed to create match: %v", err)
	}
	
	err = service.StartNextDeal(matchID)
	if err != nil {
		t.Errorf("Failed to start next deal: %v", err)
	}
	
	// Get the current player
	currentPlayer, err := service.GetCurrentPlayer(matchID)
	if err != nil {
		t.Errorf("Failed to get current player: %v", err)
	}
	
	// Get valid plays for the current player
	validPlays, err := service.GetValidPlays(matchID, currentPlayer)
	if err != nil {
		t.Errorf("Failed to get valid plays: %v", err)
	}
	
	if len(validPlays) == 0 {
		t.Error("Should have at least some valid plays")
		return
	}
	
	// Play a card
	err = service.PlayCards(matchID, currentPlayer, validPlays[0])
	if err != nil {
		t.Errorf("Failed to play cards: %v", err)
	}
	
	matchState, err := service.GetMatchState(matchID)
	if err != nil {
		t.Errorf("Failed to get match state: %v", err)
	}
	
	if matchState.Phase != engine.PhaseInProgress {
		t.Errorf("Expected phase %s, got %s", engine.PhaseInProgress, matchState.Phase)
	}
}

func TestGameServiceInvalidOperations(t *testing.T) {
	service := NewGameService()
	
	err := service.StartNextDeal("nonexistent")
	if err == nil {
		t.Error("Should not be able to start deal on nonexistent match")
	}
	
	err = service.PlayCards("nonexistent", domain.SeatEast, []domain.Card{})
	if err == nil {
		t.Error("Should not be able to play cards on nonexistent match")
	}
	
	err = service.Pass("nonexistent", domain.SeatEast)
	if err == nil {
		t.Error("Should not be able to pass on nonexistent match")
	}
	
	_, err = service.GetSnapshot("nonexistent")
	if err == nil {
		t.Error("Should return error for nonexistent match snapshot")
	}
	
	_, err = service.GetValidPlays("nonexistent", domain.SeatEast)
	if err == nil {
		t.Error("Should return error for nonexistent match")
	}
}

func TestGameServiceConcurrentMatchCreation(t *testing.T) {
	service := NewGameService()
	
	const numMatches = 10
	var wg sync.WaitGroup
	var mu sync.Mutex
	var matchIDs []domain.MatchID
	var errors []error
	
	for i := 0; i < numMatches; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			players := []*domain.Player{
				domain.NewPlayer("p1", "Player1", domain.SeatEast),
				domain.NewPlayer("p2", "Player2", domain.SeatSouth),
				domain.NewPlayer("p3", "Player3", domain.SeatWest),
				domain.NewPlayer("p4", "Player4", domain.SeatNorth),
			}
			
			options := &MatchOptions{
				DealLimit: 0,
				Seed:      int64(12345 + index),
			}
			
			matchID, err := service.CreateMatch(players, options)
			
			mu.Lock()
			if err != nil {
				errors = append(errors, err)
			} else {
				matchIDs = append(matchIDs, matchID)
			}
			mu.Unlock()
		}(i)
	}
	
	wg.Wait()
	
	if len(errors) > 0 {
		t.Errorf("Got %d errors during concurrent match creation: %v", len(errors), errors[0])
	}
	
	if len(matchIDs) != numMatches {
		t.Errorf("Expected %d matches, got %d", numMatches, len(matchIDs))
	}
	
	// Verify all matches can be accessed
	for _, matchID := range matchIDs {
		_, err := service.GetMatchState(matchID)
		if err != nil {
			t.Errorf("Match %s should be accessible: %v", matchID, err)
		}
	}
}

func TestGameServiceConcurrentGameOperations(t *testing.T) {
	service := NewGameService()
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	options := &MatchOptions{
		DealLimit: 0,
		Seed:      12345,
	}
	
	matchID, err := service.CreateMatch(players, options)
	if err != nil {
		t.Errorf("Failed to create match: %v", err)
	}
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error
	
	// Concurrent read operations
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			_, err := service.GetMatchState(matchID)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
			
			_, err = service.GetCurrentPlayer(matchID)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}()
	}
	
	wg.Wait()
	
	if len(errors) > 0 {
		t.Errorf("Got %d errors during concurrent read operations: %v", len(errors), errors[0])
	}
}

func TestGameServiceDeleteMatch(t *testing.T) {
	service := NewGameService()
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	options := &MatchOptions{
		DealLimit: 0,
		Seed:      12345,
	}
	
	matchID, err := service.CreateMatch(players, options)
	if err != nil {
		t.Errorf("Failed to create match: %v", err)
	}
	
	_, err = service.GetMatchState(matchID)
	if err != nil {
		t.Errorf("Match should exist before deletion: %v", err)
	}
	
	err = service.DeleteMatch(matchID)
	if err != nil {
		t.Errorf("Failed to delete match: %v", err)
	}
	
	_, err = service.GetMatchState(matchID)
	if err == nil {
		t.Error("Match should not exist after deletion")
	}
}

func TestGameServiceMatchLifecycle(t *testing.T) {
	service := NewGameService()
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	options := &MatchOptions{
		DealLimit: 0,
		Seed:      12345,
	}
	
	// Create match
	matchID, err := service.CreateMatch(players, options)
	if err != nil {
		t.Errorf("Failed to create match: %v", err)
	}
	
	// Start deal
	err = service.StartNextDeal(matchID)
	if err != nil {
		t.Errorf("Failed to start next deal: %v", err)
	}
	
	// Verify we're in FirstPlay phase
	matchState, err := service.GetMatchState(matchID)
	if err != nil {
		t.Errorf("Failed to get match state: %v", err)
	}
	
	if matchState.Phase != engine.PhaseFirstPlay {
		t.Errorf("Expected phase %s, got %s", engine.PhaseFirstPlay, matchState.Phase)
	}
	
	// Verify current player
	currentPlayer, err := service.GetCurrentPlayer(matchID)
	if err != nil {
		t.Errorf("Failed to get current player: %v", err)
	}
	
	// Check if it's the player's turn
	isPlayerTurn, err := service.IsPlayerTurn(matchID, currentPlayer)
	if err != nil {
		t.Errorf("Failed to check player turn: %v", err)
	}
	
	if !isPlayerTurn {
		t.Error("It should be the current player's turn")
	}
	
	// Play a card
	validPlays, err := service.GetValidPlays(matchID, currentPlayer)
	if err != nil {
		t.Errorf("Failed to get valid plays: %v", err)
	}
	
	if len(validPlays) == 0 {
		t.Error("Should have at least some valid plays")
		return
	}
	
	err = service.PlayCards(matchID, currentPlayer, validPlays[0])
	if err != nil {
		t.Errorf("Failed to play cards: %v", err)
	}
	
	// Verify state change
	matchState, err = service.GetMatchState(matchID)
	if err != nil {
		t.Errorf("Failed to get match state after playing: %v", err)
	}
	
	if matchState.Phase != engine.PhaseInProgress {
		t.Errorf("Expected phase %s, got %s", engine.PhaseInProgress, matchState.Phase)
	}
	
	// Verify turn change
	newCurrentPlayer, err := service.GetCurrentPlayer(matchID)
	if err != nil {
		t.Errorf("Failed to get current player after turn: %v", err)
	}
	
	if newCurrentPlayer == currentPlayer {
		t.Error("Current player should have changed after playing")
	}
	
	// Pass
	err = service.Pass(matchID, newCurrentPlayer)
	if err != nil {
		t.Errorf("Failed to pass: %v", err)
	}
	
	// Verify turn change after pass
	nextPlayer, err := service.GetCurrentPlayer(matchID)
	if err != nil {
		t.Errorf("Failed to get current player after pass: %v", err)
	}
	
	if nextPlayer == newCurrentPlayer {
		t.Error("Current player should have changed after passing")
	}
}

func TestGameServiceEventHandling(t *testing.T) {
	service := NewGameService()
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	options := &MatchOptions{
		DealLimit: 0,
		Seed:      12345,
	}
	
	matchID, err := service.CreateMatch(players, options)
	if err != nil {
		t.Errorf("Failed to create match: %v", err)
	}
	
	var receivedEvents []event.DomainEvent
	var mu sync.Mutex
	
	unsubscribe, err := service.Subscribe(matchID, func(e event.DomainEvent) {
		mu.Lock()
		receivedEvents = append(receivedEvents, e)
		mu.Unlock()
	})
	if err != nil {
		t.Errorf("Failed to subscribe to events: %v", err)
	}
	defer unsubscribe()
	
	err = service.StartNextDeal(matchID)
	if err != nil {
		t.Errorf("Failed to start next deal: %v", err)
	}
	
	// Give time for event processing
	time.Sleep(50 * time.Millisecond)
	
	mu.Lock()
	eventCount := len(receivedEvents)
	mu.Unlock()
	
	if eventCount < 2 {
		t.Errorf("Expected at least 2 events, got %d", eventCount)
	}
}

func TestGameServiceSnapshot(t *testing.T) {
	service := NewGameService()
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	options := &MatchOptions{
		DealLimit: 0,
		Seed:      12345,
	}
	
	matchID, err := service.CreateMatch(players, options)
	if err != nil {
		t.Errorf("Failed to create match: %v", err)
	}
	
	err = service.StartNextDeal(matchID)
	if err != nil {
		t.Errorf("Failed to start next deal: %v", err)
	}
	
	snapshot, err := service.GetSnapshot(matchID)
	if err != nil {
		t.Errorf("Failed to get snapshot: %v", err)
	}
	
	if snapshot == nil {
		t.Error("Snapshot should not be nil")
	}
	
	if snapshot.MatchID != matchID {
		t.Errorf("Expected snapshot match ID %s, got %s", matchID, snapshot.MatchID)
	}
	
	if !snapshot.IsValid() {
		t.Error("Snapshot should be valid")
	}
	
	// Verify hands are preserved
	if len(snapshot.Hands) != 4 {
		t.Errorf("Expected 4 hands in snapshot, got %d", len(snapshot.Hands))
	}
	
	for seat := domain.SeatEast; seat <= domain.SeatNorth; seat++ {
		if _, exists := snapshot.Hands[seat]; !exists {
			t.Errorf("Hand for seat %s should exist in snapshot", seat)
		}
	}
}