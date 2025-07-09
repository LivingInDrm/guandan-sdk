package event

import (
	"sync"
	"time"
	"guandan/sdk/domain"
)

type DomainEvent interface {
	EventType() string
	Timestamp() time.Time
	MatchID() domain.MatchID
}

type BaseEvent struct {
	EventTypeName string
	EventTime     time.Time
	MatchIDValue  domain.MatchID
}

func (e BaseEvent) EventType() string {
	return e.EventTypeName
}

func (e BaseEvent) Timestamp() time.Time {
	return e.EventTime
}

func (e BaseEvent) MatchID() domain.MatchID {
	return e.MatchIDValue
}

type MatchCreatedEvent struct {
	BaseEvent
	Players []domain.Player
	Teams   [2]domain.Team
	Seed    int64
}

func NewMatchCreatedEvent(matchID domain.MatchID, players []domain.Player, teams [2]domain.Team, seed int64) *MatchCreatedEvent {
	return &MatchCreatedEvent{
		BaseEvent: BaseEvent{
			EventTypeName: "MatchCreated",
			EventTime:     time.Now(),
			MatchIDValue:  matchID,
		},
		Players: players,
		Teams:   teams,
		Seed:    seed,
	}
}

type DealStartedEvent struct {
	BaseEvent
	DealNumber  int
	Trump       domain.Rank
	FirstPlayer domain.SeatID
}

func NewDealStartedEvent(matchID domain.MatchID, dealNumber int, trump domain.Rank, firstPlayer domain.SeatID) *DealStartedEvent {
	return &DealStartedEvent{
		BaseEvent: BaseEvent{
			EventTypeName: "DealStarted",
			EventTime:     time.Now(),
			MatchIDValue:  matchID,
		},
		DealNumber:  dealNumber,
		Trump:       trump,
		FirstPlayer: firstPlayer,
	}
}

type CardsDealtEvent struct {
	BaseEvent
	Hands map[domain.SeatID][]domain.Card
}

func NewCardsDealtEvent(matchID domain.MatchID, hands map[domain.SeatID][]domain.Card) *CardsDealtEvent {
	return &CardsDealtEvent{
		BaseEvent: BaseEvent{
			EventTypeName: "CardsDealt",
			EventTime:     time.Now(),
			MatchIDValue:  matchID,
		},
		Hands: hands,
	}
}

type TributeRequestedEvent struct {
	BaseEvent
	RequiredTributes map[domain.SeatID]int
}

func NewTributeRequestedEvent(matchID domain.MatchID, requiredTributes map[domain.SeatID]int) *TributeRequestedEvent {
	return &TributeRequestedEvent{
		BaseEvent: BaseEvent{
			EventTypeName: "TributeRequested",
			EventTime:     time.Now(),
			MatchIDValue:  matchID,
		},
		RequiredTributes: requiredTributes,
	}
}

type TributeGivenEvent struct {
	BaseEvent
	From  domain.SeatID
	To    domain.SeatID
	Cards []domain.Card
}

func NewTributeGivenEvent(matchID domain.MatchID, from, to domain.SeatID, cards []domain.Card) *TributeGivenEvent {
	return &TributeGivenEvent{
		BaseEvent: BaseEvent{
			EventTypeName: "TributeGiven",
			EventTime:     time.Now(),
			MatchIDValue:  matchID,
		},
		From:  from,
		To:    to,
		Cards: cards,
	}
}

type CardsPlayedEvent struct {
	BaseEvent
	Player    domain.SeatID
	Cards     []domain.Card
	CardGroup *domain.CardGroup
}

func NewCardsPlayedEvent(matchID domain.MatchID, player domain.SeatID, cards []domain.Card, cardGroup *domain.CardGroup) *CardsPlayedEvent {
	return &CardsPlayedEvent{
		BaseEvent: BaseEvent{
			EventTypeName: "CardsPlayed",
			EventTime:     time.Now(),
			MatchIDValue:  matchID,
		},
		Player:    player,
		Cards:     cards,
		CardGroup: cardGroup,
	}
}

type PlayerPassedEvent struct {
	BaseEvent
	Player domain.SeatID
}

func NewPlayerPassedEvent(matchID domain.MatchID, player domain.SeatID) *PlayerPassedEvent {
	return &PlayerPassedEvent{
		BaseEvent: BaseEvent{
			EventTypeName: "PlayerPassed",
			EventTime:     time.Now(),
			MatchIDValue:  matchID,
		},
		Player: player,
	}
}

type TrickWonEvent struct {
	BaseEvent
	Winner      domain.SeatID
	TrickNumber int
}

func NewTrickWonEvent(matchID domain.MatchID, winner domain.SeatID, trickNumber int) *TrickWonEvent {
	return &TrickWonEvent{
		BaseEvent: BaseEvent{
			EventTypeName: "TrickWon",
			EventTime:     time.Now(),
			MatchIDValue:  matchID,
		},
		Winner:      winner,
		TrickNumber: trickNumber,
	}
}

type PlayerFinishedEvent struct {
	BaseEvent
	Player   domain.SeatID
	Position int
}

func NewPlayerFinishedEvent(matchID domain.MatchID, player domain.SeatID, position int) *PlayerFinishedEvent {
	return &PlayerFinishedEvent{
		BaseEvent: BaseEvent{
			EventTypeName: "PlayerFinished",
			EventTime:     time.Now(),
			MatchIDValue:  matchID,
		},
		Player:   player,
		Position: position,
	}
}

type DealEndedEvent struct {
	BaseEvent
	DealNumber int
	RankList   []domain.SeatID
	WinnerTeam domain.TeamID
}

func NewDealEndedEvent(matchID domain.MatchID, dealNumber int, rankList []domain.SeatID, winnerTeam domain.TeamID) *DealEndedEvent {
	return &DealEndedEvent{
		BaseEvent: BaseEvent{
			EventTypeName: "DealEnded",
			EventTime:     time.Now(),
			MatchIDValue:  matchID,
		},
		DealNumber: dealNumber,
		RankList:   rankList,
		WinnerTeam: winnerTeam,
	}
}

type MatchEndedEvent struct {
	BaseEvent
	WinnerTeam domain.TeamID
	FinalScore map[domain.TeamID]int
}

func NewMatchEndedEvent(matchID domain.MatchID, winnerTeam domain.TeamID, finalScore map[domain.TeamID]int) *MatchEndedEvent {
	return &MatchEndedEvent{
		BaseEvent: BaseEvent{
			EventTypeName: "MatchEnded",
			EventTime:     time.Now(),
			MatchIDValue:  matchID,
		},
		WinnerTeam: winnerTeam,
		FinalScore: finalScore,
	}
}

type EventBus struct {
	mu           sync.RWMutex
	subscribers  map[domain.MatchID][]chan<- DomainEvent
	globalChan   chan DomainEvent
	bufferSize   int
	isRunning    bool
	stopChan     chan struct{}
}

func NewEventBus(bufferSize int) *EventBus {
	return &EventBus{
		subscribers: make(map[domain.MatchID][]chan<- DomainEvent),
		globalChan:  make(chan DomainEvent, bufferSize),
		bufferSize:  bufferSize,
		stopChan:    make(chan struct{}),
	}
}

func (eb *EventBus) Start() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	if eb.isRunning {
		return
	}
	
	eb.isRunning = true
	go eb.eventLoop()
}

func (eb *EventBus) Stop() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	if !eb.isRunning {
		return
	}
	
	eb.isRunning = false
	close(eb.stopChan)
}

func (eb *EventBus) eventLoop() {
	for {
		select {
		case event := <-eb.globalChan:
			eb.distributeEvent(event)
		case <-eb.stopChan:
			return
		}
	}
}

func (eb *EventBus) distributeEvent(event DomainEvent) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	
	if subscribers, exists := eb.subscribers[event.MatchID()]; exists {
		for _, subscriber := range subscribers {
			select {
			case subscriber <- event:
			default:
			}
		}
	}
}

func (eb *EventBus) Publish(event DomainEvent) {
	select {
	case eb.globalChan <- event:
	default:
	}
}

func (eb *EventBus) Subscribe(matchID domain.MatchID) (<-chan DomainEvent, func()) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	eventChan := make(chan DomainEvent, eb.bufferSize)
	eb.subscribers[matchID] = append(eb.subscribers[matchID], eventChan)
	
	unsubscribe := func() {
		eb.mu.Lock()
		defer eb.mu.Unlock()
		
		if subscribers, exists := eb.subscribers[matchID]; exists {
			for i, subscriber := range subscribers {
				if subscriber == eventChan {
					eb.subscribers[matchID] = append(subscribers[:i], subscribers[i+1:]...)
					close(eventChan)
					break
				}
			}
			
			if len(eb.subscribers[matchID]) == 0 {
				delete(eb.subscribers, matchID)
			}
		}
	}
	
	return eventChan, unsubscribe
}

func (eb *EventBus) SubscribeWithCallback(matchID domain.MatchID, callback func(DomainEvent)) func() {
	eventChan, unsubscribe := eb.Subscribe(matchID)
	
	go func() {
		for event := range eventChan {
			callback(event)
		}
	}()
	
	return unsubscribe
}

func (eb *EventBus) GetSubscriberCount(matchID domain.MatchID) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	
	return len(eb.subscribers[matchID])
}

func (eb *EventBus) ClearSubscribers(matchID domain.MatchID) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	if subscribers, exists := eb.subscribers[matchID]; exists {
		for _, subscriber := range subscribers {
			close(subscriber)
		}
		delete(eb.subscribers, matchID)
	}
}