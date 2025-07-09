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
	
	err := sm.StartDeal(1, domain.Two, domain.SeatEast)
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
	
	err = sm.StartDeal(1, domain.Two, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	err = sm.StartDeal(2, domain.Three, domain.SeatSouth)
	if err == nil {
		t.Error("Should not be able to start deal twice")
	}
	
	err = sm.StartTribute()
	if err == nil {
		t.Error("Should not be able to start tribute before dealing cards")
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
	
	err := sm.StartDeal(1, domain.Two, domain.SeatEast)
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
	
	if dealCtx.Trump != domain.Two {
		t.Errorf("Expected trump %s, got %s", domain.Two, dealCtx.Trump)
	}
	
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
	
	err := sm.StartDeal(1, domain.Two, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	err = sm.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
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
	
	err := sm.StartDeal(1, domain.Two, domain.SeatEast)
	if err != nil {
		t.Errorf("Failed to start deal: %v", err)
	}
	
	err = sm.DealCards()
	if err != nil {
		t.Errorf("Failed to deal cards: %v", err)
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
	
	err := sm.StartDeal(1, domain.Two, domain.SeatEast)
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
		if dealStartedEvent.Trump != domain.Two {
			t.Errorf("Expected trump %s, got %s", domain.Two, dealStartedEvent.Trump)
		}
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
	
	err := sm.StartDeal(1, domain.Two, domain.SeatEast)
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
	
	err := sm.StartDeal(1, domain.Two, domain.SeatEast)
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