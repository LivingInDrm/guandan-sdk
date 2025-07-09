package service

import (
	"encoding/json"
	"testing"
	"time"
	"guandan/sdk/domain"
)

func TestMatchSnapshotSerialization(t *testing.T) {
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	hands := make(map[domain.SeatID][]domain.Card)
	for _, player := range players {
		hands[player.SeatID] = player.GetHand()
	}
	
	snapshot := CreateSnapshotFromGameState("test-match", matchCtx, nil, nil, hands, nil)
	
	// Test JSON serialization
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Errorf("Failed to marshal snapshot: %v", err)
	}
	
	if len(data) == 0 {
		t.Error("Serialized data should not be empty")
	}
	
	// Test JSON deserialization
	var deserializedSnapshot MatchSnapshot
	err = json.Unmarshal(data, &deserializedSnapshot)
	if err != nil {
		t.Errorf("Failed to unmarshal snapshot: %v", err)
	}
	
	if deserializedSnapshot.MatchCtx.ID != snapshot.MatchCtx.ID {
		t.Errorf("Expected match ID %s, got %s", snapshot.MatchCtx.ID, deserializedSnapshot.MatchCtx.ID)
	}
	
	if deserializedSnapshot.MatchCtx.Seed != snapshot.MatchCtx.Seed {
		t.Errorf("Expected seed %d, got %d", snapshot.MatchCtx.Seed, deserializedSnapshot.MatchCtx.Seed)
	}
}

func TestMatchSnapshotWithDeal(t *testing.T) {
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	dealCtx := domain.NewDealCtx(1, domain.Two, domain.SeatEast)
	hands := make(map[domain.SeatID][]domain.Card)
	for _, player := range players {
		hands[player.SeatID] = player.GetHand()
	}
	
	snapshot := CreateSnapshotFromGameState("test-match", matchCtx, dealCtx, nil, hands, nil)
	
	if snapshot.DealCtx.DealNumber != 1 {
		t.Errorf("Expected deal number 1, got %d", snapshot.DealCtx.DealNumber)
	}
	
	if snapshot.DealCtx.Trump != domain.Two {
		t.Errorf("Expected trump %s, got %s", domain.Two, snapshot.DealCtx.Trump)
	}
	
	if snapshot.DealCtx.FirstPlayer != domain.SeatEast {
		t.Errorf("Expected first player %s, got %s", domain.SeatEast, snapshot.DealCtx.FirstPlayer)
	}
	
	// Test serialization with deal context
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Errorf("Failed to marshal snapshot with deal: %v", err)
	}
	
	var deserializedSnapshot MatchSnapshot
	err = json.Unmarshal(data, &deserializedSnapshot)
	if err != nil {
		t.Errorf("Failed to unmarshal snapshot with deal: %v", err)
	}
	
	if deserializedSnapshot.DealCtx.DealNumber != snapshot.DealCtx.DealNumber {
		t.Errorf("Expected deal number %d, got %d", snapshot.DealCtx.DealNumber, deserializedSnapshot.DealCtx.DealNumber)
	}
}

func TestMatchSnapshotWithTrick(t *testing.T) {
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	dealCtx := domain.NewDealCtx(1, domain.Two, domain.SeatEast)
	trickCtx := domain.NewTrickCtx(1, domain.SeatEast)
	
	// Add a play to the trick
	cards := []domain.Card{domain.NewCard(domain.Hearts, domain.Ace)}
	cardGroup := domain.NewCardGroup(cards)
	trickCtx = trickCtx.WithLastPlay(cardGroup, domain.SeatEast)
	
	hands := make(map[domain.SeatID][]domain.Card)
	for _, player := range players {
		hands[player.SeatID] = player.GetHand()
	}
	
	snapshot := CreateSnapshotFromGameState("test-match", matchCtx, dealCtx, trickCtx, hands, nil)
	
	if snapshot.TrickCtx.TrickNumber != 1 {
		t.Errorf("Expected trick number 1, got %d", snapshot.TrickCtx.TrickNumber)
	}
	
	if snapshot.TrickCtx.StartPlayer != domain.SeatEast {
		t.Errorf("Expected start player %s, got %s", domain.SeatEast, snapshot.TrickCtx.StartPlayer)
	}
	
	if snapshot.TrickCtx.LastPlay == nil {
		t.Error("LastPlay should not be nil")
	}
	
	if snapshot.TrickCtx.LastPlayer != domain.SeatEast {
		t.Errorf("Expected last player %s, got %s", domain.SeatEast, snapshot.TrickCtx.LastPlayer)
	}
	
	// Test serialization with trick context
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Errorf("Failed to marshal snapshot with trick: %v", err)
	}
	
	var deserializedSnapshot MatchSnapshot
	err = json.Unmarshal(data, &deserializedSnapshot)
	if err != nil {
		t.Errorf("Failed to unmarshal snapshot with trick: %v", err)
	}
	
	if deserializedSnapshot.TrickCtx.TrickNumber != snapshot.TrickCtx.TrickNumber {
		t.Errorf("Expected trick number %d, got %d", snapshot.TrickCtx.TrickNumber, deserializedSnapshot.TrickCtx.TrickNumber)
	}
	
	if deserializedSnapshot.TrickCtx.LastPlay == nil {
		t.Error("Deserialized LastPlay should not be nil")
	}
}

func TestMatchSnapshotPlayerHands(t *testing.T) {
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	// Add some cards to player hands
	testCards := []domain.Card{
		domain.NewCard(domain.Hearts, domain.Ace),
		domain.NewCard(domain.Spades, domain.King),
		domain.NewCard(domain.Clubs, domain.Queen),
	}
	
	for _, player := range players {
		player.AddCards(testCards)
	}
	
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	hands := make(map[domain.SeatID][]domain.Card)
	for _, player := range players {
		hands[player.SeatID] = player.GetHand()
	}
	
	snapshot := CreateSnapshotFromGameState("test-match", matchCtx, nil, nil, hands, nil)
	
	// Test that player hands are preserved in snapshot
	for seat := domain.SeatEast; seat <= domain.SeatNorth; seat++ {
		hand, exists := snapshot.Hands[seat]
		if !exists {
			t.Errorf("Hand for seat %s should exist in snapshot", seat)
			continue
		}
		
		if len(hand) != len(testCards) {
			t.Errorf("Expected %d cards in hand for %s, got %d", len(testCards), seat, len(hand))
		}
		
		for i, card := range hand {
			if i < len(testCards) && card != testCards[i] {
				t.Errorf("Expected card %v at position %d for %s, got %v", testCards[i], i, seat, card)
			}
		}
	}
	
	// Test serialization preserves player hands
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Errorf("Failed to marshal snapshot: %v", err)
	}
	
	var deserializedSnapshot MatchSnapshot
	err = json.Unmarshal(data, &deserializedSnapshot)
	if err != nil {
		t.Errorf("Failed to unmarshal snapshot: %v", err)
	}
	
	// Verify deserialized player hands
	for seat := domain.SeatEast; seat <= domain.SeatNorth; seat++ {
		hand, exists := deserializedSnapshot.Hands[seat]
		if !exists {
			t.Errorf("Deserialized hand for seat %s should exist", seat)
			continue
		}
		
		if len(hand) != len(testCards) {
			t.Errorf("Expected %d cards in deserialized hand for %s, got %d", len(testCards), seat, len(hand))
		}
	}
}

func TestSnapshotValidation(t *testing.T) {
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	hands := make(map[domain.SeatID][]domain.Card)
	for _, player := range players {
		hands[player.SeatID] = player.GetHand()
	}
	
	snapshot := CreateSnapshotFromGameState("test-match", matchCtx, nil, nil, hands, nil)
	
	if !snapshot.IsValid() {
		t.Error("Valid snapshot should return true for IsValid()")
	}
	
	// Test with invalid snapshot (empty hands)
	invalidSnapshot := &MatchSnapshot{
		Version: 1,
		MatchID: "test",
		Hands:   make(map[domain.SeatID][]domain.Card),
	}
	if invalidSnapshot.IsValid() {
		t.Error("Invalid snapshot should return false for IsValid()")
	}
	
	// Test version
	if snapshot.Version != 1 {
		t.Errorf("Expected version 1, got %d", snapshot.Version)
	}
	
	// Test timestamp
	if snapshot.CreatedAt.IsZero() {
		t.Error("CreatedAt timestamp should not be zero")
	}
	
	if snapshot.UpdatedAt.IsZero() {
		t.Error("UpdatedAt timestamp should not be zero")
	}
}

func TestSnapshotComparison(t *testing.T) {
	players1 := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	players2 := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	matchCtx1 := domain.NewMatchCtx("test-match-1", players1, 12345)
	matchCtx2 := domain.NewMatchCtx("test-match-2", players2, 54321)
	
	hands1 := make(map[domain.SeatID][]domain.Card)
	for _, player := range players1 {
		hands1[player.SeatID] = player.GetHand()
	}
	
	hands2 := make(map[domain.SeatID][]domain.Card)
	for _, player := range players2 {
		hands2[player.SeatID] = player.GetHand()
	}
	
	snapshot1 := CreateSnapshotFromGameState("test-match-1", matchCtx1, nil, nil, hands1, nil)
	snapshot2 := CreateSnapshotFromGameState("test-match-2", matchCtx2, nil, nil, hands2, nil)
	
	// Different matches should have different snapshots
	if snapshot1.MatchCtx.ID == snapshot2.MatchCtx.ID {
		t.Error("Different matches should have different IDs")
	}
	
	if snapshot1.MatchCtx.Seed == snapshot2.MatchCtx.Seed {
		t.Error("Different matches should have different seeds")
	}
	
	// Same match should produce equivalent snapshots
	snapshot1Copy := CreateSnapshotFromGameState("test-match-1", matchCtx1, nil, nil, hands1, nil)
	if snapshot1.MatchCtx.ID != snapshot1Copy.MatchCtx.ID {
		t.Error("Same match should produce same ID")
	}
	
	if snapshot1.MatchCtx.Seed != snapshot1Copy.MatchCtx.Seed {
		t.Error("Same match should produce same seed")
	}
}

func TestSnapshotLargeData(t *testing.T) {
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	// Create a large hand with many cards
	var largeHand []domain.Card
	for suit := domain.Hearts; suit <= domain.Spades; suit++ {
		for rank := domain.Two; rank <= domain.Ace; rank++ {
			largeHand = append(largeHand, domain.NewCard(suit, rank))
		}
	}
	
	// Add jokers
	largeHand = append(largeHand, domain.NewJoker(domain.SmallJoker))
	largeHand = append(largeHand, domain.NewJoker(domain.BigJoker))
	
	for _, player := range players {
		player.AddCards(largeHand)
	}
	
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	dealCtx := domain.NewDealCtx(1, domain.Two, domain.SeatEast)
	trickCtx := domain.NewTrickCtx(1, domain.SeatEast)
	
	// Add multiple plays to trick history
	for i := 0; i < 10; i++ {
		cards := []domain.Card{largeHand[i]}
		cardGroup := domain.NewCardGroup(cards)
		play := domain.TrickPlay{
			Player:    domain.SeatEast,
			Cards:     cards,
			CardGroup: cardGroup,
			Timestamp: time.Now(),
		}
		trickCtx = trickCtx.WithPlayHistory(play)
	}
	
	hands := make(map[domain.SeatID][]domain.Card)
	for _, player := range players {
		hands[player.SeatID] = player.GetHand()
	}
	
	snapshot := CreateSnapshotFromGameState("test-match", matchCtx, dealCtx, trickCtx, hands, nil)
	
	// Test serialization of large data
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Errorf("Failed to marshal large snapshot: %v", err)
	}
	
	if len(data) == 0 {
		t.Error("Serialized large data should not be empty")
	}
	
	// Test deserialization
	var deserializedSnapshot MatchSnapshot
	err = json.Unmarshal(data, &deserializedSnapshot)
	if err != nil {
		t.Errorf("Failed to unmarshal large snapshot: %v", err)
	}
	
	// Verify data integrity
	if !deserializedSnapshot.IsValid() {
		t.Error("Deserialized large snapshot should be valid")
	}
	
	// Verify play history preserved
	if len(deserializedSnapshot.TrickCtx.PlayHistory) != 10 {
		t.Errorf("Expected 10 plays in history, got %d", len(deserializedSnapshot.TrickCtx.PlayHistory))
	}
}

func TestSnapshotManager(t *testing.T) {
	manager := NewSnapshotManager()
	
	if manager.GetSnapshotCount() != 0 {
		t.Errorf("Expected 0 snapshots initially, got %d", manager.GetSnapshotCount())
	}
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	hands := make(map[domain.SeatID][]domain.Card)
	for _, player := range players {
		hands[player.SeatID] = player.GetHand()
	}
	
	snapshot := CreateSnapshotFromGameState("test-match", matchCtx, nil, nil, hands, nil)
	
	// Test save
	err := manager.SaveSnapshot(snapshot)
	if err != nil {
		t.Errorf("Failed to save snapshot: %v", err)
	}
	
	if manager.GetSnapshotCount() != 1 {
		t.Errorf("Expected 1 snapshot after save, got %d", manager.GetSnapshotCount())
	}
	
	// Test load
	loadedSnapshot, err := manager.LoadSnapshot("test-match")
	if err != nil {
		t.Errorf("Failed to load snapshot: %v", err)
	}
	
	if loadedSnapshot.MatchID != snapshot.MatchID {
		t.Errorf("Expected loaded snapshot match ID %s, got %s", snapshot.MatchID, loadedSnapshot.MatchID)
	}
	
	// Test has
	if !manager.HasSnapshot("test-match") {
		t.Error("Should have snapshot for test-match")
	}
	
	if manager.HasSnapshot("nonexistent") {
		t.Error("Should not have snapshot for nonexistent match")
	}
	
	// Test delete
	err = manager.DeleteSnapshot("test-match")
	if err != nil {
		t.Errorf("Failed to delete snapshot: %v", err)
	}
	
	if manager.GetSnapshotCount() != 0 {
		t.Errorf("Expected 0 snapshots after delete, got %d", manager.GetSnapshotCount())
	}
}

func TestReplayManager(t *testing.T) {
	replayManager := NewReplayManager()
	
	players := []*domain.Player{
		domain.NewPlayer("p1", "Player1", domain.SeatEast),
		domain.NewPlayer("p2", "Player2", domain.SeatSouth),
		domain.NewPlayer("p3", "Player3", domain.SeatWest),
		domain.NewPlayer("p4", "Player4", domain.SeatNorth),
	}
	
	matchCtx := domain.NewMatchCtx("test-match", players, 12345)
	hands := make(map[domain.SeatID][]domain.Card)
	for _, player := range players {
		hands[player.SeatID] = player.GetHand()
	}
	
	snapshot := CreateSnapshotFromGameState("test-match", matchCtx, nil, nil, hands, nil)
	
	// Test record
	err := replayManager.RecordSnapshot(snapshot)
	if err != nil {
		t.Errorf("Failed to record snapshot: %v", err)
	}
	
	// Test get replay data
	replayData, err := replayManager.GetReplayData("test-match")
	if err != nil {
		t.Errorf("Failed to get replay data: %v", err)
	}
	
	if replayData.MatchID != "test-match" {
		t.Errorf("Expected replay data match ID test-match, got %s", replayData.MatchID)
	}
	
	// Test validate replay
	err = replayManager.ValidateReplay("test-match")
	if err != nil {
		t.Errorf("Failed to validate replay: %v", err)
	}
}