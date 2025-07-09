package engine

import (
	"testing"
	"time"
	"guandan/sdk/domain"
	"guandan/sdk/event"
)

func TestGameEngineInitialization(t *testing.T) {
	eventBus := event.NewEventBus(100)
	engine := NewGameEngine(eventBus)
	
	if engine.IsInitialized() {
		t.Error("Engine should not be initialized initially")
	}
	
	if engine.GetCurrentPhase() != PhaseIdle {
		t.Errorf("Expected phase %s, got %s", PhaseIdle, engine.GetCurrentPhase())
	}
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	err := engine.Initialize(matchCtx)
	if err != nil {
		t.Errorf("Failed to initialize engine: %v", err)
	}
	
	if !engine.IsInitialized() {
		t.Error("Engine should be initialized after Initialize call")
	}
	
	err = engine.Initialize(matchCtx)
	if err == nil {
		t.Error("Should not be able to initialize engine twice")
	}
}

func TestGameEngineReset(t *testing.T) {
	eventBus := event.NewEventBus(100)
	engine := NewGameEngine(eventBus)
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	err := engine.Initialize(matchCtx)
	if err != nil {
		t.Errorf("Failed to initialize engine: %v", err)
	}
	
	if !engine.IsInitialized() {
		t.Error("Engine should be initialized")
	}
	
	engine.Reset()
	
	if engine.IsInitialized() {
		t.Error("Engine should not be initialized after reset")
	}
	
	if engine.GetCurrentPhase() != PhaseIdle {
		t.Errorf("Expected phase %s after reset, got %s", PhaseIdle, engine.GetCurrentPhase())
	}
}

func TestGameEngineFullDealFlow(t *testing.T) {
	eventBus := event.NewEventBus(100)
	engine := NewGameEngine(eventBus)
	
	var receivedEvents []event.DomainEvent
	eventBus.Start()
	defer eventBus.Stop()
	
	eventChan, unsubscribe := eventBus.Subscribe("test-match")
	defer unsubscribe()
	
	go func() {
		for e := range eventChan {
			receivedEvents = append(receivedEvents, e)
		}
	}()
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	err := engine.Initialize(matchCtx)
	if err != nil {
		t.Errorf("Failed to initialize engine: %v", err)
	}
	
	err = engine.StartDeal(1, domain.Two, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	if engine.GetCurrentPhase() != PhaseCreated {
		t.Errorf("Expected phase %s, got %s", PhaseCreated, engine.GetCurrentPhase())
	}
	
	err = engine.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	if engine.GetCurrentPhase() != PhaseCardsDealt {
		t.Errorf("Expected phase %s, got %s", PhaseCardsDealt, engine.GetCurrentPhase())
	}
	
	for seat := domain.SeatEast; seat <= domain.SeatNorth; seat++ {
		hand := engine.GetPlayerHand(seat)
		if len(hand) != 27 {
			t.Errorf("Expected 27 cards in hand for seat %s, got %d", seat, len(hand))
		}
	}
	
	err = engine.StartTribute()
	if err != nil {
		t.Errorf("Failed to start tribute: %v", err)
	}
	
	if engine.GetCurrentPhase() != PhaseFirstPlay {
		t.Errorf("Expected phase %s (tribute skipped for first deal), got %s", PhaseFirstPlay, engine.GetCurrentPhase())
	}
	
	if engine.GetCurrentPlayer() != domain.SeatEast {
		t.Errorf("Expected current player to be %s, got %s", domain.SeatEast, engine.GetCurrentPlayer())
	}
	
	// Give some time for events to be processed
	time.Sleep(50 * time.Millisecond)
	
	if len(receivedEvents) < 2 {
		t.Errorf("Expected at least 2 events, got %d", len(receivedEvents))
	}
}

func TestGameEnginePlayCards(t *testing.T) {
	eventBus := event.NewEventBus(100)
	engine := NewGameEngine(eventBus)
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	err := engine.Initialize(matchCtx)
	if err != nil {
		t.Errorf("Failed to initialize engine: %v", err)
	}
	
	err = engine.StartDeal(1, domain.Two, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	err = engine.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	err = engine.StartTribute()
	if err != nil {
		t.Errorf("Failed to start tribute: %v", err)
	}
	
	hand := engine.GetPlayerHand(domain.SeatEast)
	if len(hand) == 0 {
		t.Error("Player hand should not be empty")
	}
	
	singleCard := []domain.Card{hand[0]}
	
	err = engine.PlayCards(domain.SeatEast, singleCard)
	if err != nil {
		t.Errorf("Failed to play cards: %v", err)
	}
	
	if engine.GetCurrentPhase() != PhaseInProgress {
		t.Errorf("Expected phase %s, got %s", PhaseInProgress, engine.GetCurrentPhase())
	}
	
	if engine.GetCurrentPlayer() != domain.SeatSouth {
		t.Errorf("Expected current player to be %s, got %s", domain.SeatSouth, engine.GetCurrentPlayer())
	}
	
	lastPlay := engine.GetLastPlay()
	if lastPlay == nil {
		t.Error("Last play should not be nil")
	}
	
	if lastPlay.Category != domain.Single {
		t.Errorf("Expected single card play, got %s", lastPlay.Category)
	}
	
	newHand := engine.GetPlayerHand(domain.SeatEast)
	if len(newHand) != len(hand)-1 {
		t.Errorf("Expected hand size to decrease by 1, got %d", len(newHand))
	}
}

func TestGameEnginePass(t *testing.T) {
	eventBus := event.NewEventBus(100)
	engine := NewGameEngine(eventBus)
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	err := engine.Initialize(matchCtx)
	if err != nil {
		t.Errorf("Failed to initialize engine: %v", err)
	}
	
	err = engine.StartDeal(1, domain.Two, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	err = engine.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	err = engine.StartTribute()
	if err != nil {
		t.Errorf("Failed to start tribute: %v", err)
	}
	
	hand := engine.GetPlayerHand(domain.SeatEast)
	singleCard := []domain.Card{hand[0]}
	
	err = engine.PlayCards(domain.SeatEast, singleCard)
	if err != nil {
		t.Errorf("Failed to play cards: %v", err)
	}
	
	err = engine.Pass(domain.SeatSouth)
	if err != nil {
		t.Errorf("Failed to pass: %v", err)
	}
	
	if engine.GetCurrentPlayer() != domain.SeatWest {
		t.Errorf("Expected current player to be %s, got %s", domain.SeatWest, engine.GetCurrentPlayer())
	}
	
	passedPlayers := engine.GetPassedPlayers()
	if len(passedPlayers) != 1 {
		t.Errorf("Expected 1 passed player, got %d", len(passedPlayers))
	}
	
	if passedPlayers[0] != domain.SeatSouth {
		t.Errorf("Expected passed player to be %s, got %s", domain.SeatSouth, passedPlayers[0])
	}
}

func TestGameEngineInvalidActions(t *testing.T) {
	eventBus := event.NewEventBus(100)
	engine := NewGameEngine(eventBus)
	
	err := engine.StartDeal(1, domain.Two, domain.SeatEast)
	if err == nil {
		t.Error("Should not be able to start deal on uninitialized engine")
	}
	
	err = engine.DealCards()
	if err == nil {
		t.Error("Should not be able to deal cards on uninitialized engine")
	}
	
	err = engine.PlayCards(domain.SeatEast, []domain.Card{})
	if err == nil {
		t.Error("Should not be able to play cards on uninitialized engine")
	}
	
	err = engine.Pass(domain.SeatEast)
	if err == nil {
		t.Error("Should not be able to pass on uninitialized engine")
	}
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	err = engine.Initialize(matchCtx)
	if err != nil {
		t.Errorf("Failed to initialize engine: %v", err)
	}
	
	err = engine.DealCards()
	if err == nil {
		t.Error("Should not be able to deal cards before starting deal")
	}
	
	err = engine.StartTribute()
	if err == nil {
		t.Error("Should not be able to start tribute before dealing cards")
	}
	
	err = engine.PlayCards(domain.SeatEast, []domain.Card{})
	if err == nil {
		t.Error("Should not be able to play cards before first play phase")
	}
}

func TestGameEngineValidPlays(t *testing.T) {
	eventBus := event.NewEventBus(100)
	engine := NewGameEngine(eventBus)
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	err := engine.Initialize(matchCtx)
	if err != nil {
		t.Errorf("Failed to initialize engine: %v", err)
	}
	
	err = engine.StartDeal(1, domain.Two, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	err = engine.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	err = engine.StartTribute()
	if err != nil {
		t.Errorf("Failed to start tribute: %v", err)
	}
	
	validPlays := engine.GetValidPlays(domain.SeatEast)
	if len(validPlays) == 0 {
		t.Error("Should have at least some valid plays")
	}
	
	hand := engine.GetPlayerHand(domain.SeatEast)
	if len(hand) > 0 {
		canPlay := engine.CanPlayCards(domain.SeatEast, []domain.Card{hand[0]})
		if !canPlay {
			t.Error("Should be able to play a single card in first play")
		}
	}
}

func TestGameEnginePlayerTurn(t *testing.T) {
	eventBus := event.NewEventBus(100)
	engine := NewGameEngine(eventBus)
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	err := engine.Initialize(matchCtx)
	if err != nil {
		t.Errorf("Failed to initialize engine: %v", err)
	}
	
	err = engine.StartDeal(1, domain.Two, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	err = engine.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	err = engine.StartTribute()
	if err != nil {
		t.Errorf("Failed to start tribute: %v", err)
	}
	
	if !engine.IsPlayerTurn(domain.SeatEast) {
		t.Error("Should be East's turn")
	}
	
	if engine.IsPlayerTurn(domain.SeatSouth) {
		t.Error("Should not be South's turn")
	}
	
	hand := engine.GetPlayerHand(domain.SeatEast)
	if len(hand) > 0 {
		err = engine.PlayCards(domain.SeatEast, []domain.Card{hand[0]})
		if err != nil {
			t.Errorf("Failed to play cards: %v", err)
		}
	}
	
	if engine.IsPlayerTurn(domain.SeatEast) {
		t.Error("Should not be East's turn after playing")
	}
	
	if !engine.IsPlayerTurn(domain.SeatSouth) {
		t.Error("Should be South's turn after East plays")
	}
}

func TestGameEngineActionPermissions(t *testing.T) {
	eventBus := event.NewEventBus(100)
	engine := NewGameEngine(eventBus)
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	err := engine.Initialize(matchCtx)
	if err != nil {
		t.Errorf("Failed to initialize engine: %v", err)
	}
	
	err = engine.StartDeal(1, domain.Two, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	err = engine.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	err = engine.StartTribute()
	if err != nil {
		t.Errorf("Failed to start tribute: %v", err)
	}
	
	engine.SetAllowedActions(domain.SeatEast, []string{"play"})
	
	hand := engine.GetPlayerHand(domain.SeatEast)
	if len(hand) > 0 {
		err = engine.PlayCards(domain.SeatEast, []domain.Card{hand[0]})
		if err != nil {
			t.Errorf("Should be able to play cards when allowed: %v", err)
		}
	}
	
	engine.SetAllowedActions(domain.SeatSouth, []string{"tribute"})
	
	hand = engine.GetPlayerHand(domain.SeatSouth)
	if len(hand) > 0 {
		err = engine.PlayCards(domain.SeatSouth, []domain.Card{hand[0]})
		if err == nil {
			t.Error("Should not be able to play cards when not allowed")
		}
	}
	
	engine.ClearAllowedActions(domain.SeatSouth)
	
	hand = engine.GetPlayerHand(domain.SeatSouth)
	if len(hand) > 0 {
		err = engine.PlayCards(domain.SeatSouth, []domain.Card{hand[0]})
		if err != nil {
			t.Errorf("Should be able to play cards when actions are cleared: %v", err)
		}
	}
}

func TestGameEngineEventPublishing(t *testing.T) {
	eventBus := event.NewEventBus(100)
	engine := NewGameEngine(eventBus)
	
	var receivedEvents []event.DomainEvent
	eventBus.Start()
	defer eventBus.Stop()
	
	eventChan, unsubscribe := eventBus.Subscribe("test-match")
	defer unsubscribe()
	
	go func() {
		for e := range eventChan {
			receivedEvents = append(receivedEvents, e)
		}
	}()
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	err := engine.Initialize(matchCtx)
	if err != nil {
		t.Errorf("Failed to initialize engine: %v", err)
	}
	
	err = engine.StartDeal(1, domain.Two, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	if len(receivedEvents) < 1 {
		t.Error("Should have received DealStarted event")
	}
	
	err = engine.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	if len(receivedEvents) < 2 {
		t.Error("Should have received CardsDealt event")
	}
	
	err = engine.StartTribute()
	if err != nil {
		t.Errorf("Failed to start tribute: %v", err)
	}
	
	hand := engine.GetPlayerHand(domain.SeatEast)
	if len(hand) > 0 {
		err = engine.PlayCards(domain.SeatEast, []domain.Card{hand[0]})
		if err != nil {
			t.Errorf("Failed to play cards: %v", err)
		}
	}
	
	if len(receivedEvents) < 3 {
		t.Error("Should have received CardsPlayed event")
	}
	
	err = engine.Pass(domain.SeatSouth)
	if err != nil {
		t.Errorf("Failed to pass: %v", err)
	}
	
	if len(receivedEvents) < 4 {
		t.Error("Should have received PlayerPassed event")
	}
}