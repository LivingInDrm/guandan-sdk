package service

import (
	"encoding/json"
	"fmt"
	"time"
	"guandan/sdk/domain"
	"guandan/sdk/event"
)

type MatchSnapshot struct {
	Version     int                            `json:"version"`
	MatchID     domain.MatchID                 `json:"match_id"`
	MatchCtx    domain.MatchCtx                `json:"match_ctx"`
	DealCtx     domain.DealCtx                 `json:"deal_ctx"`
	TrickCtx    domain.TrickCtx                `json:"trick_ctx"`
	Hands       map[domain.SeatID][]domain.Card `json:"hands"`
	History     []event.DomainEvent            `json:"history"`
	CreatedAt   time.Time                      `json:"created_at"`
	UpdatedAt   time.Time                      `json:"updated_at"`
}

func (s *MatchSnapshot) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

func (s *MatchSnapshot) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *MatchSnapshot) Validate() error {
	if s.Version <= 0 {
		return fmt.Errorf("invalid snapshot version: %d", s.Version)
	}
	
	if s.MatchID == "" {
		return fmt.Errorf("missing match ID")
	}
	
	if len(s.Hands) != 4 {
		return fmt.Errorf("invalid hands count: expected 4, got %d", len(s.Hands))
	}
	
	for seat := domain.SeatEast; seat <= domain.SeatNorth; seat++ {
		if _, exists := s.Hands[seat]; !exists {
			return fmt.Errorf("missing hand for seat %s", seat.String())
		}
	}
	
	return nil
}

func (s *MatchSnapshot) GetPlayerHand(seat domain.SeatID) []domain.Card {
	if hand, exists := s.Hands[seat]; exists {
		result := make([]domain.Card, len(hand))
		copy(result, hand)
		return result
	}
	return nil
}

func (s *MatchSnapshot) GetAllHands() map[domain.SeatID][]domain.Card {
	hands := make(map[domain.SeatID][]domain.Card)
	for seat, hand := range s.Hands {
		hands[seat] = make([]domain.Card, len(hand))
		copy(hands[seat], hand)
	}
	return hands
}

func (s *MatchSnapshot) GetTotalCardCount() int {
	total := 0
	for _, hand := range s.Hands {
		total += len(hand)
	}
	return total
}

func (s *MatchSnapshot) IsValid() bool {
	return s.Validate() == nil
}

func (s *MatchSnapshot) GetGameProgress() float64 {
	if s.DealCtx.DealNumber == 0 {
		return 0.0
	}
	
	totalCards := s.GetTotalCardCount()
	if totalCards == 0 {
		return 1.0
	}
	
	return float64(108-totalCards) / 108.0
}

func (s *MatchSnapshot) Clone() *MatchSnapshot {
	clone := &MatchSnapshot{
		Version:   s.Version,
		MatchID:   s.MatchID,
		MatchCtx:  s.MatchCtx,
		DealCtx:   s.DealCtx,
		TrickCtx:  s.TrickCtx,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
	
	clone.Hands = make(map[domain.SeatID][]domain.Card)
	for seat, hand := range s.Hands {
		clone.Hands[seat] = make([]domain.Card, len(hand))
		copy(clone.Hands[seat], hand)
	}
	
	clone.History = make([]event.DomainEvent, len(s.History))
	copy(clone.History, s.History)
	
	return clone
}

type SnapshotManager struct {
	snapshots map[domain.MatchID]*MatchSnapshot
}

func NewSnapshotManager() *SnapshotManager {
	return &SnapshotManager{
		snapshots: make(map[domain.MatchID]*MatchSnapshot),
	}
}

func (sm *SnapshotManager) SaveSnapshot(snapshot *MatchSnapshot) error {
	if err := snapshot.Validate(); err != nil {
		return fmt.Errorf("invalid snapshot: %w", err)
	}
	
	sm.snapshots[snapshot.MatchID] = snapshot.Clone()
	return nil
}

func (sm *SnapshotManager) LoadSnapshot(matchID domain.MatchID) (*MatchSnapshot, error) {
	snapshot, exists := sm.snapshots[matchID]
	if !exists {
		return nil, fmt.Errorf("snapshot not found for match: %s", matchID)
	}
	
	return snapshot.Clone(), nil
}

func (sm *SnapshotManager) DeleteSnapshot(matchID domain.MatchID) error {
	if _, exists := sm.snapshots[matchID]; !exists {
		return fmt.Errorf("snapshot not found for match: %s", matchID)
	}
	
	delete(sm.snapshots, matchID)
	return nil
}

func (sm *SnapshotManager) HasSnapshot(matchID domain.MatchID) bool {
	_, exists := sm.snapshots[matchID]
	return exists
}

func (sm *SnapshotManager) GetSnapshotCount() int {
	return len(sm.snapshots)
}

func (sm *SnapshotManager) GetAllMatchIDs() []domain.MatchID {
	matchIDs := make([]domain.MatchID, 0, len(sm.snapshots))
	for matchID := range sm.snapshots {
		matchIDs = append(matchIDs, matchID)
	}
	return matchIDs
}

func (sm *SnapshotManager) Clear() {
	sm.snapshots = make(map[domain.MatchID]*MatchSnapshot)
}

type ReplayManager struct {
	snapshotManager *SnapshotManager
}

func NewReplayManager() *ReplayManager {
	return &ReplayManager{
		snapshotManager: NewSnapshotManager(),
	}
}

func (rm *ReplayManager) RecordSnapshot(snapshot *MatchSnapshot) error {
	return rm.snapshotManager.SaveSnapshot(snapshot)
}

func (rm *ReplayManager) GetReplayData(matchID domain.MatchID) (*ReplayData, error) {
	snapshot, err := rm.snapshotManager.LoadSnapshot(matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to load snapshot: %w", err)
	}
	
	return &ReplayData{
		MatchID:   matchID,
		Snapshot:  snapshot,
		Events:    snapshot.History,
		CreatedAt: snapshot.CreatedAt,
		UpdatedAt: snapshot.UpdatedAt,
	}, nil
}

func (rm *ReplayManager) ReplayFromSnapshot(snapshot *MatchSnapshot) (*ReplayResult, error) {
	if err := snapshot.Validate(); err != nil {
		return nil, fmt.Errorf("invalid snapshot for replay: %w", err)
	}
	
	result := &ReplayResult{
		MatchID:    snapshot.MatchID,
		StartTime:  time.Now(),
		Events:     make([]event.DomainEvent, 0),
		IsComplete: false,
	}
	
	for _, domainEvent := range snapshot.History {
		result.Events = append(result.Events, domainEvent)
	}
	
	result.EndTime = time.Now()
	result.IsComplete = true
	result.Duration = result.EndTime.Sub(result.StartTime)
	
	return result, nil
}

func (rm *ReplayManager) ValidateReplay(matchID domain.MatchID) error {
	snapshot, err := rm.snapshotManager.LoadSnapshot(matchID)
	if err != nil {
		return fmt.Errorf("failed to load snapshot: %w", err)
	}
	
	if !snapshot.IsValid() {
		return fmt.Errorf("invalid snapshot for replay")
	}
	
	expectedCardCount := 108
	actualCardCount := snapshot.GetTotalCardCount()
	
	if actualCardCount > expectedCardCount {
		return fmt.Errorf("invalid card count: expected <= %d, got %d", expectedCardCount, actualCardCount)
	}
	
	return nil
}

type ReplayData struct {
	MatchID   domain.MatchID      `json:"match_id"`
	Snapshot  *MatchSnapshot      `json:"snapshot"`
	Events    []event.DomainEvent `json:"events"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

func (rd *ReplayData) ToJSON() ([]byte, error) {
	return json.MarshalIndent(rd, "", "  ")
}

func (rd *ReplayData) FromJSON(data []byte) error {
	return json.Unmarshal(data, rd)
}

type ReplayResult struct {
	MatchID    domain.MatchID      `json:"match_id"`
	StartTime  time.Time           `json:"start_time"`
	EndTime    time.Time           `json:"end_time"`
	Duration   time.Duration       `json:"duration"`
	Events     []event.DomainEvent `json:"events"`
	IsComplete bool                `json:"is_complete"`
	Error      string              `json:"error,omitempty"`
}

func (rr *ReplayResult) ToJSON() ([]byte, error) {
	return json.MarshalIndent(rr, "", "  ")
}

func (rr *ReplayResult) FromJSON(data []byte) error {
	return json.Unmarshal(data, rr)
}

func (rr *ReplayResult) GetEventCount() int {
	return len(rr.Events)
}

func (rr *ReplayResult) GetEventsPerSecond() float64 {
	if rr.Duration.Seconds() == 0 {
		return 0
	}
	return float64(len(rr.Events)) / rr.Duration.Seconds()
}

func CreateSnapshotFromGameState(
	matchID domain.MatchID,
	matchCtx *domain.MatchCtx,
	dealCtx *domain.DealCtx,
	trickCtx *domain.TrickCtx,
	hands map[domain.SeatID][]domain.Card,
	history []event.DomainEvent,
) *MatchSnapshot {
	snapshot := &MatchSnapshot{
		Version:   1,
		MatchID:   matchID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if matchCtx != nil {
		snapshot.MatchCtx = *matchCtx
	}
	
	if dealCtx != nil {
		snapshot.DealCtx = *dealCtx
	}
	
	if trickCtx != nil {
		snapshot.TrickCtx = *trickCtx
	}
	
	snapshot.Hands = make(map[domain.SeatID][]domain.Card)
	for seat, hand := range hands {
		snapshot.Hands[seat] = make([]domain.Card, len(hand))
		copy(snapshot.Hands[seat], hand)
	}
	
	snapshot.History = make([]event.DomainEvent, len(history))
	copy(snapshot.History, history)
	
	return snapshot
}