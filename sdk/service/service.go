package service

import (
	"fmt"
	"sync"
	"time"
	"guandan/sdk/domain"
	"guandan/sdk/engine"
	"guandan/sdk/event"
)

type MatchOptions struct {
	DealLimit int
	Seed      int64
}

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

type MatchState struct {
	MatchID     domain.MatchID
	Phase       engine.DealPhase
	CurrentDeal int
	Trump       domain.Rank
	Players     []*domain.Player
	Teams       [2]*domain.Team
	IsFinished  bool
	Winner      *domain.TeamID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GameServiceImpl struct {
	mu       sync.RWMutex
	matches  map[domain.MatchID]*MatchInstance
	eventBus *event.EventBus
	idSeed   int64
}

type MatchInstance struct {
	MatchCtx      *domain.MatchCtx
	Engine        *engine.GameEngine
	EventBus      *event.EventBus
	CreatedAt     time.Time
	UpdatedAt     time.Time
	IsActive      bool
	Subscribers   map[string]func(event.DomainEvent)
	SubscribersMu sync.RWMutex
}

func NewGameService() GameService {
	eventBus := event.NewEventBus(1000)
	eventBus.Start()
	
	return &GameServiceImpl{
		matches:  make(map[domain.MatchID]*MatchInstance),
		eventBus: eventBus,
		idSeed:   time.Now().UnixNano(),
	}
}

func (gs *GameServiceImpl) CreateMatch(players []*domain.Player, opt *MatchOptions) (domain.MatchID, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	
	if len(players) != 4 {
		return "", fmt.Errorf("exactly 4 players required, got %d", len(players))
	}
	
	if opt == nil {
		opt = &MatchOptions{
			DealLimit: 0,
			Seed:      time.Now().UnixNano(),
		}
	}
	
	matchID := gs.generateMatchID()
	
	matchCtx := domain.NewMatchCtx(matchID, players, opt.Seed)
	matchCtx = matchCtx.WithState(domain.MatchStateCreated)
	
	gameEngine := engine.NewGameEngine(gs.eventBus)
	if err := gameEngine.Initialize(matchCtx); err != nil {
		return "", fmt.Errorf("failed to initialize game engine: %w", err)
	}
	
	matchInstance := &MatchInstance{
		MatchCtx:    matchCtx,
		Engine:      gameEngine,
		EventBus:    gs.eventBus,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
		Subscribers: make(map[string]func(event.DomainEvent)),
	}
	
	gs.matches[matchID] = matchInstance
	
	matchCtx = matchCtx.WithState(domain.MatchStateInProgress)
	matchInstance.MatchCtx = matchCtx
	
	playersArray := make([]domain.Player, len(players))
	for i, player := range players {
		playersArray[i] = *player
	}
	
	teamsArray := [2]domain.Team{
		*matchCtx.GetTeam(domain.TeamEastWest),
		*matchCtx.GetTeam(domain.TeamSouthNorth),
	}
	
	gs.eventBus.Publish(event.NewMatchCreatedEvent(
		matchID,
		playersArray,
		teamsArray,
		opt.Seed,
	))
	
	return matchID, nil
}

func (gs *GameServiceImpl) StartNextDeal(matchID domain.MatchID) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	
	matchInstance, exists := gs.matches[matchID]
	if !exists {
		return fmt.Errorf("match not found: %s", matchID)
	}
	
	if !matchInstance.IsActive {
		return fmt.Errorf("match is not active: %s", matchID)
	}
	
	dealNumber := matchInstance.MatchCtx.CurrentDeal + 1
	trump := gs.calculateTrump(dealNumber)
	firstPlayer := gs.calculateFirstPlayer(dealNumber)
	
	if err := matchInstance.Engine.StartDeal(dealNumber, trump, firstPlayer); err != nil {
		return fmt.Errorf("failed to start deal: %w", err)
	}
	
	if err := matchInstance.Engine.DealCards(); err != nil {
		return fmt.Errorf("failed to deal cards: %w", err)
	}
	
	if err := matchInstance.Engine.StartTribute(); err != nil {
		return fmt.Errorf("failed to start tribute: %w", err)
	}
	
	matchInstance.MatchCtx = matchInstance.MatchCtx.WithCurrentDeal(dealNumber)
	matchInstance.UpdatedAt = time.Now()
	
	return nil
}

func (gs *GameServiceImpl) PlayCards(matchID domain.MatchID, seat domain.SeatID, cards []domain.Card) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	
	matchInstance, exists := gs.matches[matchID]
	if !exists {
		return fmt.Errorf("match not found: %s", matchID)
	}
	
	if !matchInstance.IsActive {
		return fmt.Errorf("match is not active: %s", matchID)
	}
	
	if err := matchInstance.Engine.PlayCards(seat, cards); err != nil {
		return fmt.Errorf("failed to play cards: %w", err)
	}
	
	matchInstance.UpdatedAt = time.Now()
	
	return nil
}

func (gs *GameServiceImpl) Pass(matchID domain.MatchID, seat domain.SeatID) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	
	matchInstance, exists := gs.matches[matchID]
	if !exists {
		return fmt.Errorf("match not found: %s", matchID)
	}
	
	if !matchInstance.IsActive {
		return fmt.Errorf("match is not active: %s", matchID)
	}
	
	if err := matchInstance.Engine.Pass(seat); err != nil {
		return fmt.Errorf("failed to pass: %w", err)
	}
	
	matchInstance.UpdatedAt = time.Now()
	
	return nil
}

func (gs *GameServiceImpl) GetSnapshot(matchID domain.MatchID) (*MatchSnapshot, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	
	matchInstance, exists := gs.matches[matchID]
	if !exists {
		return nil, fmt.Errorf("match not found: %s", matchID)
	}
	
	return gs.createSnapshot(matchInstance), nil
}

func (gs *GameServiceImpl) Subscribe(matchID domain.MatchID, callback func(event.DomainEvent)) (func(), error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	
	matchInstance, exists := gs.matches[matchID]
	if !exists {
		return nil, fmt.Errorf("match not found: %s", matchID)
	}
	
	subscriberID := fmt.Sprintf("%d", time.Now().UnixNano())
	
	matchInstance.SubscribersMu.Lock()
	matchInstance.Subscribers[subscriberID] = callback
	matchInstance.SubscribersMu.Unlock()
	
	unsubscribe := gs.eventBus.SubscribeWithCallback(matchID, callback)
	
	return func() {
		matchInstance.SubscribersMu.Lock()
		delete(matchInstance.Subscribers, subscriberID)
		matchInstance.SubscribersMu.Unlock()
		
		unsubscribe()
	}, nil
}

func (gs *GameServiceImpl) GetValidPlays(matchID domain.MatchID, seat domain.SeatID) ([][]domain.Card, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	
	matchInstance, exists := gs.matches[matchID]
	if !exists {
		return nil, fmt.Errorf("match not found: %s", matchID)
	}
	
	return matchInstance.Engine.GetValidPlays(seat), nil
}

func (gs *GameServiceImpl) GetCurrentPlayer(matchID domain.MatchID) (domain.SeatID, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	
	matchInstance, exists := gs.matches[matchID]
	if !exists {
		return domain.SeatEast, fmt.Errorf("match not found: %s", matchID)
	}
	
	return matchInstance.Engine.GetCurrentPlayer(), nil
}

func (gs *GameServiceImpl) IsPlayerTurn(matchID domain.MatchID, seat domain.SeatID) (bool, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	
	matchInstance, exists := gs.matches[matchID]
	if !exists {
		return false, fmt.Errorf("match not found: %s", matchID)
	}
	
	return matchInstance.Engine.IsPlayerTurn(seat), nil
}

func (gs *GameServiceImpl) GetMatchState(matchID domain.MatchID) (*MatchState, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	
	matchInstance, exists := gs.matches[matchID]
	if !exists {
		return nil, fmt.Errorf("match not found: %s", matchID)
	}
	
	dealCtx := matchInstance.Engine.GetDealCtx()
	trump := domain.Two
	if dealCtx != nil {
		trump = dealCtx.Trump
	}
	
	return &MatchState{
		MatchID:     matchID,
		Phase:       matchInstance.Engine.GetCurrentPhase(),
		CurrentDeal: matchInstance.MatchCtx.CurrentDeal,
		Trump:       trump,
		Players:     matchInstance.MatchCtx.Players.All(),
		Teams:       matchInstance.MatchCtx.Teams,
		IsFinished:  matchInstance.Engine.IsGameFinished(),
		Winner:      matchInstance.Engine.GetGameWinner(),
		CreatedAt:   matchInstance.CreatedAt,
		UpdatedAt:   matchInstance.UpdatedAt,
	}, nil
}

func (gs *GameServiceImpl) DeleteMatch(matchID domain.MatchID) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	
	matchInstance, exists := gs.matches[matchID]
	if !exists {
		return fmt.Errorf("match not found: %s", matchID)
	}
	
	matchInstance.IsActive = false
	
	matchInstance.SubscribersMu.Lock()
	for subscriberID := range matchInstance.Subscribers {
		delete(matchInstance.Subscribers, subscriberID)
	}
	matchInstance.SubscribersMu.Unlock()
	
	gs.eventBus.ClearSubscribers(matchID)
	
	delete(gs.matches, matchID)
	
	return nil
}

func (gs *GameServiceImpl) generateMatchID() domain.MatchID {
	gs.idSeed++
	return domain.MatchID(fmt.Sprintf("match_%d_%d", time.Now().Unix(), gs.idSeed))
}

func (gs *GameServiceImpl) calculateTrump(dealNumber int) domain.Rank {
	trumps := []domain.Rank{
		domain.Two, domain.Three, domain.Four, domain.Five, domain.Six,
		domain.Seven, domain.Eight, domain.Nine, domain.Ten, domain.Jack,
		domain.Queen, domain.King, domain.Ace,
	}
	
	index := (dealNumber - 1) % len(trumps)
	return trumps[index]
}

func (gs *GameServiceImpl) calculateFirstPlayer(dealNumber int) domain.SeatID {
	players := []domain.SeatID{
		domain.SeatEast, domain.SeatSouth, domain.SeatWest, domain.SeatNorth,
	}
	
	index := (dealNumber - 1) % len(players)
	return players[index]
}

func (gs *GameServiceImpl) createSnapshot(matchInstance *MatchInstance) *MatchSnapshot {
	snapshot := &MatchSnapshot{
		Version:   1,
		MatchID:   matchInstance.MatchCtx.ID,
		MatchCtx:  *matchInstance.MatchCtx,
		CreatedAt: matchInstance.CreatedAt,
		UpdatedAt: matchInstance.UpdatedAt,
	}
	
	if dealCtx := matchInstance.Engine.GetDealCtx(); dealCtx != nil {
		snapshot.DealCtx = *dealCtx
	}
	
	if trickCtx := matchInstance.Engine.GetTrickCtx(); trickCtx != nil {
		snapshot.TrickCtx = *trickCtx
	}
	
	snapshot.Hands = make(map[domain.SeatID][]domain.Card)
	for _, player := range matchInstance.MatchCtx.Players.All() {
		snapshot.Hands[player.SeatID] = player.GetHand()
	}
	
	return snapshot
}

func (gs *GameServiceImpl) GetMatchCount() int {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	
	return len(gs.matches)
}

func (gs *GameServiceImpl) GetActiveMatches() []domain.MatchID {
	gs.mu.RLock()
	defer gs.mu.RUnlock()
	
	var activeMatches []domain.MatchID
	for matchID, instance := range gs.matches {
		if instance.IsActive {
			activeMatches = append(activeMatches, matchID)
		}
	}
	
	return activeMatches
}

func (gs *GameServiceImpl) Shutdown() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	
	for matchID := range gs.matches {
		gs.DeleteMatch(matchID)
	}
	
	gs.eventBus.Stop()
}