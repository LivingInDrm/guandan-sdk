package engine

import (
	"fmt"
	"sync"
	"guandan/sdk/domain"
	"guandan/sdk/event"
)

type GameEngine struct {
	mu              sync.RWMutex
	stateMachine    *DealStateMachine
	eventBus        *event.EventBus
	isInitialized   bool
	allowedActions  map[domain.SeatID][]string
}

func NewGameEngine(eventBus *event.EventBus) *GameEngine {
	return &GameEngine{
		eventBus:       eventBus,
		allowedActions: make(map[domain.SeatID][]string),
	}
}

func (ge *GameEngine) Initialize(matchCtx *domain.MatchCtx) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	
	if ge.isInitialized {
		return fmt.Errorf("engine already initialized")
	}
	
	ge.stateMachine = NewDealStateMachine(matchCtx, ge.eventBus)
	ge.isInitialized = true
	
	return nil
}

func (ge *GameEngine) IsInitialized() bool {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	return ge.isInitialized
}

func (ge *GameEngine) GetCurrentPhase() DealPhase {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	
	if !ge.isInitialized {
		return PhaseIdle
	}
	
	return ge.stateMachine.GetCurrentPhase()
}

func (ge *GameEngine) GetMatchCtx() *domain.MatchCtx {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	
	if !ge.isInitialized {
		return nil
	}
	
	return ge.stateMachine.GetMatchCtx()
}

func (ge *GameEngine) GetDealCtx() *domain.DealCtx {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	
	if !ge.isInitialized {
		return nil
	}
	
	return ge.stateMachine.GetDealCtx()
}

func (ge *GameEngine) GetTrickCtx() *domain.TrickCtx {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	
	if !ge.isInitialized {
		return nil
	}
	
	return ge.stateMachine.GetTrickCtx()
}

func (ge *GameEngine) StartDeal(dealNumber int, trump domain.Rank, firstPlayer domain.SeatID) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	
	if !ge.isInitialized {
		return fmt.Errorf("engine not initialized")
	}
	
	return ge.stateMachine.StartDeal(dealNumber, trump, firstPlayer)
}

func (ge *GameEngine) DealCards() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	
	if !ge.isInitialized {
		return fmt.Errorf("engine not initialized")
	}
	
	return ge.stateMachine.DealCards()
}

func (ge *GameEngine) StartTribute() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	
	if !ge.isInitialized {
		return fmt.Errorf("engine not initialized")
	}
	
	return ge.stateMachine.StartTribute()
}

func (ge *GameEngine) GiveTribute(from, to domain.SeatID, cards []domain.Card) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	
	if !ge.isInitialized {
		return fmt.Errorf("engine not initialized")
	}
	
	if !ge.isActionAllowed(from, "tribute") {
		return fmt.Errorf("tribute action not allowed for player %s", from.String())
	}
	
	return ge.stateMachine.GiveTribute(from, to, cards)
}

func (ge *GameEngine) PlayCards(seat domain.SeatID, cards []domain.Card) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	
	if !ge.isInitialized {
		return fmt.Errorf("engine not initialized")
	}
	
	if !ge.isActionAllowed(seat, "play") {
		return fmt.Errorf("play action not allowed for player %s", seat.String())
	}
	
	if !ge.isValidPlay(seat, cards) {
		return fmt.Errorf("invalid play for player %s", seat.String())
	}
	
	return ge.stateMachine.PlayCards(seat, cards)
}

func (ge *GameEngine) Pass(seat domain.SeatID) error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	
	if !ge.isInitialized {
		return fmt.Errorf("engine not initialized")
	}
	
	if !ge.isActionAllowed(seat, "pass") {
		return fmt.Errorf("pass action not allowed for player %s", seat.String())
	}
	
	return ge.stateMachine.Pass(seat)
}

func (ge *GameEngine) StartFirstPlay() error {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	
	if !ge.isInitialized {
		return fmt.Errorf("engine not initialized")
	}
	
	return ge.stateMachine.StartFirstPlay()
}

func (ge *GameEngine) GetValidPlays(seat domain.SeatID) [][]domain.Card {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	
	if !ge.isInitialized {
		return nil
	}
	
	matchCtx := ge.stateMachine.GetMatchCtx()
	dealCtx := ge.stateMachine.GetDealCtx()
	trickCtx := ge.stateMachine.GetTrickCtx()
	
	if matchCtx == nil || dealCtx == nil {
		return nil
	}
	
	player := matchCtx.GetPlayer(seat)
	if player == nil {
		return nil
	}
	
	var tablePlay *domain.CardGroup
	if trickCtx != nil {
		tablePlay = trickCtx.LastPlay
	}
	
	return domain.GetPlayableCards(player.GetHand(), tablePlay, dealCtx.Trump)
}

func (ge *GameEngine) CanPlayCards(seat domain.SeatID, cards []domain.Card) bool {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	
	if !ge.isInitialized {
		return false
	}
	
	return ge.isValidPlay(seat, cards)
}

func (ge *GameEngine) GetCurrentPlayer() domain.SeatID {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	
	if !ge.isInitialized {
		return domain.SeatEast
	}
	
	trickCtx := ge.stateMachine.GetTrickCtx()
	if trickCtx == nil {
		return domain.SeatEast
	}
	
	return trickCtx.CurrentPlayer
}

func (ge *GameEngine) IsPlayerTurn(seat domain.SeatID) bool {
	return ge.GetCurrentPlayer() == seat
}

func (ge *GameEngine) GetPlayerHand(seat domain.SeatID) []domain.Card {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	
	if !ge.isInitialized {
		return nil
	}
	
	matchCtx := ge.stateMachine.GetMatchCtx()
	if matchCtx == nil {
		return nil
	}
	
	player := matchCtx.GetPlayer(seat)
	if player == nil {
		return nil
	}
	
	return player.GetHand()
}

func (ge *GameEngine) GetLastPlay() *domain.CardGroup {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	
	if !ge.isInitialized {
		return nil
	}
	
	trickCtx := ge.stateMachine.GetTrickCtx()
	if trickCtx == nil {
		return nil
	}
	
	return trickCtx.LastPlay
}

func (ge *GameEngine) GetPassedPlayers() []domain.SeatID {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	
	if !ge.isInitialized {
		return nil
	}
	
	trickCtx := ge.stateMachine.GetTrickCtx()
	if trickCtx == nil {
		return nil
	}
	
	var passedPlayers []domain.SeatID
	for seat := domain.SeatEast; seat <= domain.SeatNorth; seat++ {
		if trickCtx.HasPlayerPassed(seat) {
			passedPlayers = append(passedPlayers, seat)
		}
	}
	
	return passedPlayers
}

func (ge *GameEngine) isActionAllowed(seat domain.SeatID, action string) bool {
	allowedActions, exists := ge.allowedActions[seat]
	if !exists {
		return true
	}
	
	for _, allowedAction := range allowedActions {
		if allowedAction == action {
			return true
		}
	}
	
	return false
}

func (ge *GameEngine) isValidPlay(seat domain.SeatID, cards []domain.Card) bool {
	matchCtx := ge.stateMachine.GetMatchCtx()
	dealCtx := ge.stateMachine.GetDealCtx()
	trickCtx := ge.stateMachine.GetTrickCtx()
	
	if matchCtx == nil || dealCtx == nil {
		return false
	}
	
	player := matchCtx.GetPlayer(seat)
	if player == nil {
		return false
	}
	
	if !player.HasCards(cards) {
		return false
	}
	
	cardGroup := domain.NewCardGroup(cards)
	if !cardGroup.IsValid() {
		return false
	}
	
	var tablePlay *domain.CardGroup
	if trickCtx != nil {
		tablePlay = trickCtx.LastPlay
	}
	
	return domain.CanFollow(cardGroup, tablePlay, dealCtx.Trump)
}

func (ge *GameEngine) SetAllowedActions(seat domain.SeatID, actions []string) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	
	ge.allowedActions[seat] = actions
}

func (ge *GameEngine) ClearAllowedActions(seat domain.SeatID) {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	
	delete(ge.allowedActions, seat)
}

func (ge *GameEngine) IsGameFinished() bool {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	
	if !ge.isInitialized {
		return false
	}
	
	matchCtx := ge.stateMachine.GetMatchCtx()
	if matchCtx == nil {
		return false
	}
	
	return matchCtx.IsFinished()
}

func (ge *GameEngine) GetGameWinner() *domain.TeamID {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	
	if !ge.isInitialized {
		return nil
	}
	
	matchCtx := ge.stateMachine.GetMatchCtx()
	if matchCtx == nil {
		return nil
	}
	
	return matchCtx.Winner
}

func (ge *GameEngine) Reset() {
	ge.mu.Lock()
	defer ge.mu.Unlock()
	
	if ge.stateMachine != nil {
		ge.stateMachine.Reset()
	}
	
	ge.isInitialized = false
	ge.allowedActions = make(map[domain.SeatID][]string)
}

func (ge *GameEngine) GetStateMachine() *DealStateMachine {
	ge.mu.RLock()
	defer ge.mu.RUnlock()
	
	return ge.stateMachine
}