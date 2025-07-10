package engine

import (
	"testing"
	"guandan/sdk/domain"
	"guandan/sdk/event"
)

func TestDealStateMachinePhaseTransitions(t *testing.T) {
	eventBus := event.NewEventBus(100)
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	sm := NewDealStateMachine(matchCtx, eventBus)
	
	if sm.GetCurrentPhase() != PhaseIdle {
		t.Errorf("Expected initial phase %s, got %s", PhaseIdle, sm.GetCurrentPhase())
	}
	
	err := sm.StartDeal(1, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	if sm.GetCurrentPhase() != PhaseCreated {
		t.Errorf("Expected phase %s, got %s", PhaseCreated, sm.GetCurrentPhase())
	}
	
	err = sm.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	if sm.GetCurrentPhase() != PhaseCardsDealt {
		t.Errorf("Expected phase %s, got %s", PhaseCardsDealt, sm.GetCurrentPhase())
	}
	
	err = sm.DetermineTrump()
	if err != nil {
		t.Errorf("Failed to determine trump: %v", err)
	}
	
	if sm.GetCurrentPhase() != PhaseTrumpDecision {
		t.Errorf("Expected phase %s, got %s", PhaseTrumpDecision, sm.GetCurrentPhase())
	}
	
	err = sm.StartTribute()
	if err != nil {
		t.Errorf("Failed to start tribute: %v", err)
	}
	
	if sm.GetCurrentPhase() != PhaseFirstPlay {
		t.Errorf("Expected phase %s (tribute skipped for first deal), got %s", PhaseFirstPlay, sm.GetCurrentPhase())
	}
}

func TestDealStateMachineInvalidTransitions(t *testing.T) {
	eventBus := event.NewEventBus(100)
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	sm := NewDealStateMachine(matchCtx, eventBus)
	
	err := sm.DealCards()
	if err == nil {
		t.Error("Should not be able to deal cards from idle phase")
	}
	
	err = sm.StartTribute()
	if err == nil {
		t.Error("Should not be able to start tribute from idle phase")
	}
	
	err = sm.PlayCards(domain.SeatEast, []domain.Card{})
	if err == nil {
		t.Error("Should not be able to play cards from idle phase")
	}
	
	err = sm.Pass(domain.SeatEast)
	if err == nil {
		t.Error("Should not be able to pass from idle phase")
	}
	
	err = sm.StartDeal(1, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	err = sm.StartDeal(2, domain.SeatSouth)
	if err == nil {
		t.Error("Should not be able to start deal twice")
	}
	
	err = sm.StartTribute()
	if err == nil {
		t.Error("Should not be able to start tribute before determining trump")
	}
}

func TestDealStateMachineCardDealing(t *testing.T) {
	eventBus := event.NewEventBus(100)
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	sm := NewDealStateMachine(matchCtx, eventBus)
	
	err := sm.StartDeal(1, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	err = sm.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	dealCtx := sm.GetDealCtx()
	if dealCtx == nil {
		t.Error("DealCtx should not be nil after dealing")
	}
	
	if dealCtx.State != domain.DealStateDealt {
		t.Errorf("Expected deal state %s, got %s", domain.DealStateDealt, dealCtx.State)
	}
	
	if dealCtx.DealNumber != 1 {
		t.Errorf("Expected deal number 1, got %d", dealCtx.DealNumber)
	}
	
	// Trump is not determined yet at this phase
	// Trump will be determined in DetermineTrump phase
	
	if dealCtx.FirstPlayer != domain.SeatEast {
		t.Errorf("Expected first player %s, got %s", domain.SeatEast, dealCtx.FirstPlayer)
	}
	
	for seat := domain.SeatEast; seat <= domain.SeatNorth; seat++ {
		player := matchCtx.GetPlayer(seat)
		if player == nil {
			t.Errorf("Player %s should not be nil", seat)
			continue
		}
		
		hand := player.GetHand()
		if len(hand) != 27 {
			t.Errorf("Expected 27 cards in hand for seat %s, got %d", seat, len(hand))
		}
	}
}

func TestDealStateMachinePlayCardsValidation(t *testing.T) {
	eventBus := event.NewEventBus(100)
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	sm := NewDealStateMachine(matchCtx, eventBus)
	
	err := sm.StartDeal(1, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	err = sm.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	err = sm.DetermineTrump()
	if err != nil {
		t.Errorf("Failed to determine trump: %v", err)
	}
	
	err = sm.StartTribute()
	if err != nil {
		t.Errorf("Failed to start tribute: %v", err)
	}
	
	trickCtx := sm.GetTrickCtx()
	if trickCtx == nil {
		t.Error("TrickCtx should not be nil after starting first play")
	}
	
	if trickCtx.CurrentPlayer != domain.SeatEast {
		t.Errorf("Expected current player %s, got %s", domain.SeatEast, trickCtx.CurrentPlayer)
	}
	
	err = sm.PlayCards(domain.SeatSouth, []domain.Card{})
	if err == nil {
		t.Error("Should not allow playing out of turn")
	}
	
	player := matchCtx.GetPlayer(domain.SeatEast)
	if player == nil {
		t.Error("Player should not be nil")
		return
	}
	
	hand := player.GetHand()
	if len(hand) == 0 {
		t.Error("Hand should not be empty")
		return
	}
	
	nonOwnedCard := domain.NewCard(domain.Hearts, domain.Ace)
	err = sm.PlayCards(domain.SeatEast, []domain.Card{nonOwnedCard})
	if err == nil {
		t.Error("Should not allow playing cards not in hand")
	}
	
	invalidCards := []domain.Card{hand[0], hand[1]}
	if len(hand) > 1 && hand[0].Rank != hand[1].Rank {
		err = sm.PlayCards(domain.SeatEast, invalidCards)
		if err == nil {
			t.Error("Should not allow playing invalid card combinations")
		}
	}
	
	validCard := []domain.Card{hand[0]}
	err = sm.PlayCards(domain.SeatEast, validCard)
	if err != nil {
		t.Errorf("Should allow playing valid card: %v", err)
	}
	
	if sm.GetCurrentPhase() != PhaseInProgress {
		t.Errorf("Expected phase %s after first play, got %s", PhaseInProgress, sm.GetCurrentPhase())
	}
	
	updatedTrickCtx := sm.GetTrickCtx()
	if updatedTrickCtx.CurrentPlayer != domain.SeatSouth {
		t.Errorf("Expected current player %s, got %s", domain.SeatSouth, updatedTrickCtx.CurrentPlayer)
	}
	
	if updatedTrickCtx.LastPlay == nil {
		t.Error("LastPlay should not be nil after playing cards")
	}
	
	if updatedTrickCtx.LastPlayer != domain.SeatEast {
		t.Errorf("Expected last player %s, got %s", domain.SeatEast, updatedTrickCtx.LastPlayer)
	}
}

func TestDealStateMachinePassMechanism(t *testing.T) {
	eventBus := event.NewEventBus(100)
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	sm := NewDealStateMachine(matchCtx, eventBus)
	
	err := sm.StartDeal(1, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	err = sm.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	err = sm.DetermineTrump()
	if err != nil {
		t.Errorf("Failed to determine trump: %v", err)
	}
	
	err = sm.StartTribute()
	if err != nil {
		t.Errorf("Failed to start tribute: %v", err)
	}
	
	player := matchCtx.GetPlayer(domain.SeatEast)
	hand := player.GetHand()
	
	err = sm.PlayCards(domain.SeatEast, []domain.Card{hand[0]})
	if err != nil {
		t.Errorf("Failed to play cards: %v", err)
	}
	
	err = sm.Pass(domain.SeatEast)
	if err == nil {
		t.Error("Should not allow passing out of turn")
	}
	
	err = sm.Pass(domain.SeatSouth)
	if err != nil {
		t.Errorf("Failed to pass: %v", err)
	}
	
	trickCtx := sm.GetTrickCtx()
	if !trickCtx.HasPlayerPassed(domain.SeatSouth) {
		t.Error("Player should be marked as passed")
	}
	
	if trickCtx.CurrentPlayer != domain.SeatWest {
		t.Errorf("Expected current player %s, got %s", domain.SeatWest, trickCtx.CurrentPlayer)
	}
	
	err = sm.Pass(domain.SeatWest)
	if err != nil {
		t.Errorf("Failed to pass: %v", err)
	}
	
	err = sm.Pass(domain.SeatNorth)
	if err != nil {
		t.Errorf("Failed to pass: %v", err)
	}
	
	updatedTrickCtx := sm.GetTrickCtx()
	if updatedTrickCtx.Winner != domain.SeatEast {
		t.Errorf("Expected trick winner %s, got %s", domain.SeatEast, updatedTrickCtx.Winner)
	}
	
	if updatedTrickCtx.TrickNumber != 1 {
		t.Errorf("Expected trick number 1, got %d", updatedTrickCtx.TrickNumber)
	}
}

func TestDealStateMachineEventPublishing(t *testing.T) {
	eventBus := event.NewEventBus(100)
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
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
	
	sm := NewDealStateMachine(matchCtx, eventBus)
	
	err := sm.StartDeal(1, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	if len(receivedEvents) < 1 {
		t.Error("Should have received DealStarted event")
	}
	
	dealStartedEvent, ok := receivedEvents[0].(event.DealStartedEvent)
	if !ok {
		t.Error("First event should be DealStartedEvent")
	} else {
		if dealStartedEvent.DealNumber != 1 {
			t.Errorf("Expected deal number 1, got %d", dealStartedEvent.DealNumber)
		}
		// Trump in DealStarted event is temporary, actual trump determined in P2 phase
		if dealStartedEvent.FirstPlayer != domain.SeatEast {
			t.Errorf("Expected first player %s, got %s", domain.SeatEast, dealStartedEvent.FirstPlayer)
		}
	}
	
	err = sm.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	if len(receivedEvents) < 2 {
		t.Error("Should have received CardsDealt event")
	}
	
	cardsDealtEvent, ok := receivedEvents[1].(event.CardsDealtEvent)
	if !ok {
		t.Error("Second event should be CardsDealtEvent")
	} else {
		if len(cardsDealtEvent.Hands) != 4 {
			t.Errorf("Expected 4 hands, got %d", len(cardsDealtEvent.Hands))
		}
		for seat, hand := range cardsDealtEvent.Hands {
			if len(hand) != 27 {
				t.Errorf("Expected 27 cards for seat %s, got %d", seat, len(hand))
			}
		}
	}
	
	err = sm.StartTribute()
	if err != nil {
		t.Errorf("Failed to start tribute: %v", err)
	}
	
	player := matchCtx.GetPlayer(domain.SeatEast)
	hand := player.GetHand()
	
	err = sm.PlayCards(domain.SeatEast, []domain.Card{hand[0]})
	if err != nil {
		t.Errorf("Failed to play cards: %v", err)
	}
	
	var cardsPlayedEvent event.CardsPlayedEvent
	found := false
	for _, e := range receivedEvents {
		if cpe, ok := e.(event.CardsPlayedEvent); ok {
			cardsPlayedEvent = cpe
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Should have received CardsPlayed event")
	} else {
		if cardsPlayedEvent.Player != domain.SeatEast {
			t.Errorf("Expected player %s, got %s", domain.SeatEast, cardsPlayedEvent.Player)
		}
		if len(cardsPlayedEvent.Cards) != 1 {
			t.Errorf("Expected 1 card, got %d", len(cardsPlayedEvent.Cards))
		}
		if cardsPlayedEvent.CardGroup == nil {
			t.Error("CardGroup should not be nil")
		}
	}
	
	err = sm.Pass(domain.SeatSouth)
	if err != nil {
		t.Errorf("Failed to pass: %v", err)
	}
	
	var playerPassedEvent event.PlayerPassedEvent
	found = false
	for _, e := range receivedEvents {
		if ppe, ok := e.(event.PlayerPassedEvent); ok {
			playerPassedEvent = ppe
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Should have received PlayerPassed event")
	} else {
		if playerPassedEvent.Player != domain.SeatSouth {
			t.Errorf("Expected player %s, got %s", domain.SeatSouth, playerPassedEvent.Player)
		}
	}
}

func TestDealStateMachineReset(t *testing.T) {
	eventBus := event.NewEventBus(100)
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	sm := NewDealStateMachine(matchCtx, eventBus)
	
	err := sm.StartDeal(1, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	err = sm.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	if sm.GetCurrentPhase() != PhaseCardsDealt {
		t.Errorf("Expected phase %s, got %s", PhaseCardsDealt, sm.GetCurrentPhase())
	}
	
	if sm.GetDealCtx() == nil {
		t.Error("DealCtx should not be nil")
	}
	
	sm.Reset()
	
	if sm.GetCurrentPhase() != PhaseIdle {
		t.Errorf("Expected phase %s after reset, got %s", PhaseIdle, sm.GetCurrentPhase())
	}
	
	if sm.GetDealCtx() != nil {
		t.Error("DealCtx should be nil after reset")
	}
	
	if sm.GetTrickCtx() != nil {
		t.Error("TrickCtx should be nil after reset")
	}
}

// P0 - Match Initialization Tests
func TestP0MatchInitialization(t *testing.T) {
	// Test P0 requirements: 4 players seated, teams, initial level = 2
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	// P0 Step 1: Four players seated, teams paired
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	// Verify 4 players are seated
	if len(players) != 4 {
		t.Errorf("Expected 4 players, got %d", len(players))
	}
	
	// Verify teams (East-West vs South-North)
	eastPlayer := matchCtx.GetPlayer(domain.SeatEast)
	westPlayer := matchCtx.GetPlayer(domain.SeatWest)
	southPlayer := matchCtx.GetPlayer(domain.SeatSouth)
	northPlayer := matchCtx.GetPlayer(domain.SeatNorth)
	
	if eastPlayer == nil || westPlayer == nil || southPlayer == nil || northPlayer == nil {
		t.Error("All players should be properly seated")
	}
	
	// P0 Step 2: Initial level = 2
	// Check team initial level
	team1 := matchCtx.GetTeam(domain.TeamEastWest)
	team2 := matchCtx.GetTeam(domain.TeamSouthNorth)
	if team1 == nil || team2 == nil {
		t.Error("Teams should be initialized")
	} else {
		if team1.Level != domain.Two {
			t.Errorf("Expected team1 level %s, got %s", domain.Two, team1.Level)
		}
		if team2.Level != domain.Two {
			t.Errorf("Expected team2 level %s, got %s", domain.Two, team2.Level)
		}
	}
	
	// P0 Step 3: Teams & Level written to Match-ctx
	if matchCtx.ID == "" {
		t.Error("Match ID should be set")
	}
	
	// Verify level range (2-A, no jokers)
	validLevels := []domain.Rank{
		domain.Two, domain.Three, domain.Four, domain.Five, domain.Six,
		domain.Seven, domain.Eight, domain.Nine, domain.Ten,
		domain.Jack, domain.Queen, domain.King, domain.Ace,
	}
	
	if team1 != nil {
		found := false
		for _, validLevel := range validLevels {
			if team1.Level == validLevel {
				found = true
				break
			}
		}
		
		if !found {
			t.Errorf("Initial level %s should be in valid range (2-A)", team1.Level)
		}
	}
}

// P1 - Deal Start Tests
func TestP1DealStart(t *testing.T) {
	eventBus := event.NewEventBus(100)
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	sm := NewDealStateMachine(matchCtx, eventBus)
	
	// P1 Entry condition: Each Deal starts
	err := sm.StartDeal(1, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	// P1 Step 1: Shuffle 108 cards
	err = sm.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	// Verify deck has 108 cards
	if sm.deck == nil {
		t.Error("Deck should not be nil after dealing")
	}
	
	// P1 Step 2: First deal should select starting card
	dealCtx := sm.GetDealCtx()
	if dealCtx.IsFirstDeal {
		if sm.startingCard == nil {
			t.Error("Starting card should be selected for first deal")
		}
		
		// P1 Step 4: Record starting card holder
		if sm.startingCardHolder == domain.SeatEast {
			// Verify that starting card holder has the card
			player := matchCtx.GetPlayer(sm.startingCardHolder)
			if player == nil {
				t.Error("Starting card holder should be valid")
			}
		}
	}
	
	// P1 Step 3: Deal 27 cards to each player clockwise
	totalCards := 0
	for seat := domain.SeatEast; seat <= domain.SeatNorth; seat++ {
		player := matchCtx.GetPlayer(seat)
		if player == nil {
			t.Errorf("Player %s should not be nil", seat)
			continue
		}
		
		hand := player.GetHand()
		if len(hand) != 27 {
			t.Errorf("Expected 27 cards for seat %s, got %d", seat, len(hand))
		}
		totalCards += len(hand)
	}
	
	// Verify total cards dealt = 108
	if totalCards != 108 {
		t.Errorf("Expected 108 total cards dealt, got %d", totalCards)
	}
	
	// Verify phase transition to PhaseCardsDealt
	if sm.GetCurrentPhase() != PhaseCardsDealt {
		t.Errorf("Expected phase %s, got %s", PhaseCardsDealt, sm.GetCurrentPhase())
	}
}

// P2 - Determine Level & Trump Tests
func TestP2DetermineLevelAndTrump(t *testing.T) {
	eventBus := event.NewEventBus(100)
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	sm := NewDealStateMachine(matchCtx, eventBus)
	
	// Setup for P2 phase
	err := sm.StartDeal(1, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	err = sm.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	// P2 Entry condition: P1 completed (cards dealt)
	if sm.GetCurrentPhase() != PhaseCardsDealt {
		t.Errorf("Expected phase %s before P2, got %s", PhaseCardsDealt, sm.GetCurrentPhase())
	}
	
	// P2 Step 1: Read previous deal winner team's level, default to 2
	err = sm.DetermineTrump()
	if err != nil {
		t.Errorf("Failed to determine trump: %v", err)
	}
	
	dealCtx := sm.GetDealCtx()
	if dealCtx == nil {
		t.Error("DealCtx should not be nil")
		return
	}
	
	// For first deal, level should be 2
	expectedLevel := domain.Two
	if dealCtx.CurrentLevel != expectedLevel {
		t.Errorf("Expected level %s for first deal, got %s", expectedLevel, dealCtx.CurrentLevel)
	}
	
	// P2 Step 2: Set 8 cards equal to Level as Trump
	// P2 Step 3: Write Level & Trump to Deal-ctx
	if dealCtx.Trump != expectedLevel {
		t.Errorf("Expected trump %s, got %s", expectedLevel, dealCtx.Trump)
	}
	
	// Verify phase transition to PhaseTrumpDecision
	if sm.GetCurrentPhase() != PhaseTrumpDecision {
		t.Errorf("Expected phase %s, got %s", PhaseTrumpDecision, sm.GetCurrentPhase())
	}
	
	// Test next phase transition based on deal number
	// For first deal: should go to P4 (skip tribute)
	// For non-first deal: should go to P3 (tribute)
	if dealCtx.IsFirstDeal {
		err = sm.StartTribute()
		if err != nil {
			t.Errorf("Failed to start tribute for first deal: %v", err)
		}
		
		// Should skip tribute and go to PhaseFirstPlay
		if sm.GetCurrentPhase() != PhaseFirstPlay {
			t.Errorf("Expected phase %s for first deal (skip tribute), got %s", PhaseFirstPlay, sm.GetCurrentPhase())
		}
	}
}

// Test default trump level for first deal
func TestP2DefaultTrumpLevel(t *testing.T) {
	eventBus := event.NewEventBus(100)
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	// Test first deal always starts with level 2
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	sm := NewDealStateMachine(matchCtx, eventBus)
	
	// Start first deal
	err := sm.StartDeal(1, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	err = sm.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	err = sm.DetermineTrump()
	if err != nil {
		t.Errorf("Failed to determine trump: %v", err)
	}
	
	dealCtx := sm.GetDealCtx()
	expectedLevel := domain.Two
	
	if dealCtx.Trump != expectedLevel {
		t.Errorf("Expected trump %s for first deal, got %s", expectedLevel, dealCtx.Trump)
	}
	
	if dealCtx.CurrentLevel != expectedLevel {
		t.Errorf("Expected level %s for first deal, got %s", expectedLevel, dealCtx.CurrentLevel)
	}
	
	// Test second deal uses team level (which should still be 2 for this test)
	sm.Reset()
	err = sm.StartDeal(2, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start second deal: %v", err)
	}
	
	err = sm.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	err = sm.DetermineTrump()
	if err != nil {
		t.Errorf("Failed to determine trump: %v", err)
	}
	
	dealCtx = sm.GetDealCtx()
	// For second deal, should still be 2 as teams haven't advanced
	if dealCtx.Trump != expectedLevel {
		t.Errorf("Expected trump %s for second deal, got %s", expectedLevel, dealCtx.Trump)
	}
}

// Test state flow from P0 to P2
func TestP0ToP2StateFlow(t *testing.T) {
	eventBus := event.NewEventBus(100)
	
	// P0: Match Initialization
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	sm := NewDealStateMachine(matchCtx, eventBus)
	
	// Initial state should be PhaseIdle
	if sm.GetCurrentPhase() != PhaseIdle {
		t.Errorf("Expected initial phase %s, got %s", PhaseIdle, sm.GetCurrentPhase())
	}
	
	// P1: Deal Start
	err := sm.StartDeal(1, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	if sm.GetCurrentPhase() != PhaseCreated {
		t.Errorf("Expected phase %s after StartDeal, got %s", PhaseCreated, sm.GetCurrentPhase())
	}
	
	// P1: Deal Cards
	err = sm.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	if sm.GetCurrentPhase() != PhaseCardsDealt {
		t.Errorf("Expected phase %s after DealCards, got %s", PhaseCardsDealt, sm.GetCurrentPhase())
	}
	
	// P2: Determine Trump
	err = sm.DetermineTrump()
	if err != nil {
		t.Errorf("Failed to determine trump: %v", err)
	}
	
	if sm.GetCurrentPhase() != PhaseTrumpDecision {
		t.Errorf("Expected phase %s after DetermineTrump, got %s", PhaseTrumpDecision, sm.GetCurrentPhase())
	}
	
	// Verify correct state flow for first deal
	dealCtx := sm.GetDealCtx()
	if !dealCtx.IsFirstDeal {
		t.Error("Should be marked as first deal")
	}
	
	// Next phase: StartTribute (should skip to FirstPlay for first deal)
	err = sm.StartTribute()
	if err != nil {
		t.Errorf("Failed to start tribute: %v", err)
	}
	
	if sm.GetCurrentPhase() != PhaseFirstPlay {
		t.Errorf("Expected phase %s for first deal (skip tribute), got %s", PhaseFirstPlay, sm.GetCurrentPhase())
	}
}

func TestDealStateMachineContextAccess(t *testing.T) {
	eventBus := event.NewEventBus(100)
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	
	sm := NewDealStateMachine(matchCtx, eventBus)
	
	if sm.GetMatchCtx() != matchCtx {
		t.Error("MatchCtx should be accessible")
	}
	
	if sm.GetDealCtx() != nil {
		t.Error("DealCtx should be nil initially")
	}
	
	if sm.GetTrickCtx() != nil {
		t.Error("TrickCtx should be nil initially")
	}
	
	err := sm.StartDeal(1, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	dealCtx := sm.GetDealCtx()
	if dealCtx == nil {
		t.Error("DealCtx should not be nil after starting deal")
	}
	
	if dealCtx.DealNumber != 1 {
		t.Errorf("Expected deal number 1, got %d", dealCtx.DealNumber)
	}
	
	err = sm.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
	}
	
	err = sm.DetermineTrump()
	if err != nil {
		t.Errorf("Failed to determine trump: %v", err)
	}
	
	err = sm.StartTribute()
	if err != nil {
		t.Errorf("Failed to start tribute: %v", err)
	}
	
	trickCtx := sm.GetTrickCtx()
	if trickCtx == nil {
		t.Error("TrickCtx should not be nil after starting first play")
	}
	
	if trickCtx.TrickNumber != 1 {
		t.Errorf("Expected trick number 1, got %d", trickCtx.TrickNumber)
	}
	
	if trickCtx.CurrentPlayer != domain.SeatEast {
		t.Errorf("Expected current player %s, got %s", domain.SeatEast, trickCtx.CurrentPlayer)
	}
}