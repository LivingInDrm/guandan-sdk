package engine

import (
	"fmt"
	"time"
	"guandan/sdk/domain"
	"guandan/sdk/event"
)

type DealPhase int

const (
	PhaseIdle DealPhase = iota
	PhaseCreated
	PhaseCardsDealt
	PhaseTribute
	PhaseFirstPlay
	PhaseInProgress
	PhaseRankList
	PhaseFinished
)

func (p DealPhase) String() string {
	switch p {
	case PhaseIdle:
		return "Idle"
	case PhaseCreated:
		return "Created"
	case PhaseCardsDealt:
		return "CardsDealt"
	case PhaseTribute:
		return "Tribute"
	case PhaseFirstPlay:
		return "FirstPlay"
	case PhaseInProgress:
		return "InProgress"
	case PhaseRankList:
		return "RankList"
	case PhaseFinished:
		return "Finished"
	default:
		return "Unknown"
	}
}

type DealStateMachine struct {
	currentPhase DealPhase
	matchCtx     *domain.MatchCtx
	dealCtx      *domain.DealCtx
	trickCtx     *domain.TrickCtx
	eventBus     *event.EventBus
	deck         *domain.Deck
}

func NewDealStateMachine(matchCtx *domain.MatchCtx, eventBus *event.EventBus) *DealStateMachine {
	return &DealStateMachine{
		currentPhase: PhaseIdle,
		matchCtx:     matchCtx,
		eventBus:     eventBus,
	}
}

func (sm *DealStateMachine) GetCurrentPhase() DealPhase {
	return sm.currentPhase
}

func (sm *DealStateMachine) GetMatchCtx() *domain.MatchCtx {
	return sm.matchCtx
}

func (sm *DealStateMachine) GetDealCtx() *domain.DealCtx {
	return sm.dealCtx
}

func (sm *DealStateMachine) GetTrickCtx() *domain.TrickCtx {
	return sm.trickCtx
}

func (sm *DealStateMachine) StartDeal(dealNumber int, trump domain.Rank, firstPlayer domain.SeatID) error {
	if sm.currentPhase != PhaseIdle {
		return fmt.Errorf("cannot start deal from phase %s", sm.currentPhase.String())
	}
	
	sm.dealCtx = domain.NewDealCtx(dealNumber, trump, firstPlayer)
	sm.currentPhase = PhaseCreated
	
	sm.eventBus.Publish(event.NewDealStartedEvent(
		sm.matchCtx.ID,
		dealNumber,
		trump,
		firstPlayer,
	))
	
	return nil
}

func (sm *DealStateMachine) DealCards() error {
	if sm.currentPhase != PhaseCreated {
		return fmt.Errorf("cannot deal cards from phase %s", sm.currentPhase.String())
	}
	
	sm.deck = domain.NewDeckWithSeed(sm.matchCtx.Seed)
	sm.deck.Shuffle()
	
	hands := sm.deck.DealToHands(4)
	handMap := make(map[domain.SeatID][]domain.Card)
	
	for i, hand := range hands {
		seat := domain.SeatID(i)
		player := sm.matchCtx.GetPlayer(seat)
		if player != nil {
			player.ClearHand()
			player.AddCards(hand)
			handMap[seat] = hand
		}
	}
	
	sm.dealCtx = sm.dealCtx.WithState(domain.DealStateDealt)
	sm.currentPhase = PhaseCardsDealt
	
	sm.eventBus.Publish(event.NewCardsDealtEvent(
		sm.matchCtx.ID,
		handMap,
	))
	
	return nil
}

func (sm *DealStateMachine) StartTribute() error {
	if sm.currentPhase != PhaseCardsDealt {
		return fmt.Errorf("cannot start tribute from phase %s", sm.currentPhase.String())
	}
	
	if sm.dealCtx.IsFirstDeal {
		return sm.skipTribute()
	}
	
	tributeRequirements := sm.calculateTributeRequirements()
	
	sm.dealCtx = sm.dealCtx.WithState(domain.DealStateTribute)
	sm.currentPhase = PhaseTribute
	
	sm.eventBus.Publish(event.NewTributeRequestedEvent(
		sm.matchCtx.ID,
		tributeRequirements,
	))
	
	return nil
}

func (sm *DealStateMachine) skipTribute() error {
	sm.dealCtx = sm.dealCtx.WithTributeGiven(true)
	return sm.StartFirstPlay()
}

func (sm *DealStateMachine) GiveTribute(from, to domain.SeatID, cards []domain.Card) error {
	if sm.currentPhase != PhaseTribute {
		return fmt.Errorf("cannot give tribute from phase %s", sm.currentPhase.String())
	}
	
	fromPlayer := sm.matchCtx.GetPlayer(from)
	toPlayer := sm.matchCtx.GetPlayer(to)
	
	if fromPlayer == nil || toPlayer == nil {
		return fmt.Errorf("invalid player seats")
	}
	
	if !fromPlayer.HasCards(cards) {
		return fmt.Errorf("player does not have required cards")
	}
	
	fromPlayer.RemoveCards(cards)
	toPlayer.AddCards(cards)
	
	sm.dealCtx.TributeCards[from] = cards
	
	sm.eventBus.Publish(event.NewTributeGivenEvent(
		sm.matchCtx.ID,
		from,
		to,
		cards,
	))
	
	if sm.allTributesGiven() {
		sm.dealCtx = sm.dealCtx.WithTributeGiven(true)
		return sm.StartFirstPlay()
	}
	
	return nil
}

func (sm *DealStateMachine) StartFirstPlay() error {
	if sm.currentPhase != PhaseTribute && sm.currentPhase != PhaseCardsDealt {
		return fmt.Errorf("cannot start first play from phase %s", sm.currentPhase.String())
	}
	
	sm.dealCtx = sm.dealCtx.WithState(domain.DealStateFirstPlay)
	sm.currentPhase = PhaseFirstPlay
	
	sm.trickCtx = domain.NewTrickCtx(1, sm.dealCtx.FirstPlayer)
	
	return nil
}

func (sm *DealStateMachine) TransitionToInProgress() error {
	if sm.currentPhase != PhaseFirstPlay {
		return fmt.Errorf("cannot transition to in progress from phase %s", sm.currentPhase.String())
	}
	
	sm.dealCtx = sm.dealCtx.WithState(domain.DealStateInProgress)
	sm.currentPhase = PhaseInProgress
	
	return nil
}

func (sm *DealStateMachine) StartNewTrick(startPlayer domain.SeatID) error {
	if sm.currentPhase != PhaseInProgress {
		return fmt.Errorf("cannot start new trick from phase %s", sm.currentPhase.String())
	}
	
	sm.dealCtx = sm.dealCtx.WithTrickCount(sm.dealCtx.TrickCount + 1)
	sm.trickCtx = domain.NewTrickCtx(sm.dealCtx.TrickCount, startPlayer)
	
	return nil
}

func (sm *DealStateMachine) PlayCards(seat domain.SeatID, cards []domain.Card) error {
	if sm.currentPhase != PhaseFirstPlay && sm.currentPhase != PhaseInProgress {
		return fmt.Errorf("cannot play cards from phase %s", sm.currentPhase.String())
	}
	
	if sm.trickCtx.CurrentPlayer != seat {
		return fmt.Errorf("not player's turn")
	}
	
	player := sm.matchCtx.GetPlayer(seat)
	if player == nil {
		return fmt.Errorf("player not found")
	}
	
	if !player.HasCards(cards) {
		return fmt.Errorf("player does not have required cards")
	}
	
	cardGroup := domain.NewCardGroup(cards)
	if !cardGroup.IsValid() {
		return fmt.Errorf("invalid card combination")
	}
	
	if !domain.CanFollow(cardGroup, sm.trickCtx.LastPlay, sm.dealCtx.Trump) {
		return fmt.Errorf("cannot beat current play")
	}
	
	player.RemoveCards(cards)
	
	sm.trickCtx = sm.trickCtx.WithLastPlay(cardGroup, seat)
	sm.trickCtx = sm.trickCtx.WithCurrentPlayer(seat.Next())
	
	trickPlay := domain.TrickPlay{
		Player:    seat,
		Cards:     cards,
		CardGroup: cardGroup,
		Timestamp: time.Now(),
	}
	sm.trickCtx = sm.trickCtx.WithPlayHistory(trickPlay)
	
	sm.eventBus.Publish(event.NewCardsPlayedEvent(
		sm.matchCtx.ID,
		seat,
		cards,
		cardGroup,
	))
	
	if sm.currentPhase == PhaseFirstPlay {
		sm.TransitionToInProgress()
	}
	
	if player.IsHandEmpty() {
		return sm.handlePlayerFinished(seat)
	}
	
	return nil
}

func (sm *DealStateMachine) Pass(seat domain.SeatID) error {
	if sm.currentPhase != PhaseInProgress {
		return fmt.Errorf("cannot pass from phase %s", sm.currentPhase.String())
	}
	
	if sm.trickCtx.CurrentPlayer != seat {
		return fmt.Errorf("not player's turn")
	}
	
	sm.trickCtx = sm.trickCtx.WithPlayerPassed(seat)
	sm.trickCtx = sm.trickCtx.WithCurrentPlayer(seat.Next())
	
	sm.eventBus.Publish(event.NewPlayerPassedEvent(
		sm.matchCtx.ID,
		seat,
	))
	
	if sm.trickCtx.ShouldFinish() {
		return sm.finishTrick()
	}
	
	return nil
}

func (sm *DealStateMachine) finishTrick() error {
	winner := sm.trickCtx.LastPlayer
	sm.trickCtx = sm.trickCtx.WithWinner(winner)
	
	sm.eventBus.Publish(event.NewTrickWonEvent(
		sm.matchCtx.ID,
		winner,
		sm.trickCtx.TrickNumber,
	))
	
	if sm.shouldFinishDeal() {
		return sm.finishDeal()
	}
	
	return sm.StartNewTrick(winner)
}

func (sm *DealStateMachine) handlePlayerFinished(seat domain.SeatID) error {
	position := len(sm.dealCtx.RankList) + 1
	sm.dealCtx = sm.dealCtx.WithRankList(append(sm.dealCtx.RankList, seat))
	
	sm.eventBus.Publish(event.NewPlayerFinishedEvent(
		sm.matchCtx.ID,
		seat,
		position,
	))
	
	if sm.shouldFinishDeal() {
		return sm.finishDeal()
	}
	
	return nil
}

func (sm *DealStateMachine) shouldFinishDeal() bool {
	return len(sm.dealCtx.RankList) >= 3
}

func (sm *DealStateMachine) finishDeal() error {
	sm.currentPhase = PhaseRankList
	
	winnerTeam := sm.determineWinnerTeam()
	
	sm.eventBus.Publish(event.NewDealEndedEvent(
		sm.matchCtx.ID,
		sm.dealCtx.DealNumber,
		sm.dealCtx.RankList,
		winnerTeam,
	))
	
	if sm.shouldFinishMatch() {
		return sm.finishMatch(winnerTeam)
	}
	
	sm.currentPhase = PhaseFinished
	return nil
}

func (sm *DealStateMachine) finishMatch(winnerTeam domain.TeamID) error {
	sm.currentPhase = PhaseFinished
	
	finalScore := make(map[domain.TeamID]int)
	finalScore[winnerTeam] = 1
	finalScore[winnerTeam.OpposingTeam()] = 0
	
	sm.eventBus.Publish(event.NewMatchEndedEvent(
		sm.matchCtx.ID,
		winnerTeam,
		finalScore,
	))
	
	sm.matchCtx = sm.matchCtx.WithWinner(winnerTeam)
	
	return nil
}

func (sm *DealStateMachine) determineWinnerTeam() domain.TeamID {
	if len(sm.dealCtx.RankList) >= 2 {
		first := sm.dealCtx.RankList[0]
		second := sm.dealCtx.RankList[1]
		
		if domain.GetTeamFromSeat(first) == domain.GetTeamFromSeat(second) {
			return domain.GetTeamFromSeat(first)
		}
	}
	
	return domain.GetTeamFromSeat(sm.dealCtx.RankList[0])
}

func (sm *DealStateMachine) shouldFinishMatch() bool {
	return sm.dealCtx.CurrentLevel >= domain.Ace
}

func (sm *DealStateMachine) calculateTributeRequirements() map[domain.SeatID]int {
	requirements := make(map[domain.SeatID]int)
	
	return requirements
}

func (sm *DealStateMachine) allTributesGiven() bool {
	return true
}

func (sm *DealStateMachine) Reset() {
	sm.currentPhase = PhaseIdle
	sm.dealCtx = nil
	sm.trickCtx = nil
	sm.deck = nil
}