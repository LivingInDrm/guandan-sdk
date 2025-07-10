package domain

import (
	"testing"
	"time"
)

func TestMatchStateString(t *testing.T) {
	testCases := []struct {
		state    MatchState
		expected string
	}{
		{MatchStateCreated, "Created"},
		{MatchStateInProgress, "InProgress"},
		{MatchStateFinished, "Finished"},
		{MatchState(999), "Unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.state.String()
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestNewMatchCtx(t *testing.T) {
	matchID := MatchID("test-match-123")
	players := []*Player{
		NewPlayer("p1", "Alice", SeatEast),
		NewPlayer("p2", "Bob", SeatSouth),
		NewPlayer("p3", "Charlie", SeatWest),
		NewPlayer("p4", "David", SeatNorth),
	}
	seed := int64(12345)

	matchCtx := NewMatchCtx(matchID, players, seed)

	if matchCtx.ID != matchID {
		t.Errorf("Expected match ID %s, got %s", matchID, matchCtx.ID)
	}

	if matchCtx.State != MatchStateCreated {
		t.Errorf("Expected state %v, got %v", MatchStateCreated, matchCtx.State)
	}

	if matchCtx.Seed != seed {
		t.Errorf("Expected seed %d, got %d", seed, matchCtx.Seed)
	}

	if matchCtx.CurrentDeal != 0 {
		t.Errorf("Expected current deal 0, got %d", matchCtx.CurrentDeal)
	}

	if matchCtx.MaxDeals != 0 {
		t.Errorf("Expected max deals 0, got %d", matchCtx.MaxDeals)
	}

	if matchCtx.Winner != nil {
		t.Errorf("Expected no winner initially, got %v", matchCtx.Winner)
	}

	if time.Since(matchCtx.StartTime) > time.Second {
		t.Error("Start time should be recent")
	}

	if matchCtx.EndTime != nil {
		t.Error("End time should be nil for new match")
	}

	// Check players are assigned correctly
	for _, player := range players {
		retrievedPlayer := matchCtx.GetPlayer(player.SeatID)
		if retrievedPlayer == nil {
			t.Errorf("Player at seat %v not found", player.SeatID)
		}
		if retrievedPlayer.ID != player.ID {
			t.Errorf("Expected player ID %s, got %s", player.ID, retrievedPlayer.ID)
		}
	}

	// Check teams are set up correctly
	eastWestTeam := matchCtx.GetTeam(TeamEastWest)
	if eastWestTeam == nil {
		t.Error("East-West team not found")
	}

	southNorthTeam := matchCtx.GetTeam(TeamSouthNorth)
	if southNorthTeam == nil {
		t.Error("South-North team not found")
	}
}

func TestMatchCtxImmutableUpdates(t *testing.T) {
	players := []*Player{
		NewPlayer("p1", "Alice", SeatEast),
		NewPlayer("p2", "Bob", SeatSouth),
		NewPlayer("p3", "Charlie", SeatWest),
		NewPlayer("p4", "David", SeatNorth),
	}
	original := NewMatchCtx("test-match", players, 12345)

	// Test WithState
	updated := original.WithState(MatchStateInProgress)
	if original.State == MatchStateInProgress {
		t.Error("Original match context should not be modified")
	}
	if updated.State != MatchStateInProgress {
		t.Error("Updated match context should have new state")
	}

	// Test WithCurrentDeal
	updated = original.WithCurrentDeal(5)
	if original.CurrentDeal == 5 {
		t.Error("Original match context should not be modified")
	}
	if updated.CurrentDeal != 5 {
		t.Error("Updated match context should have new deal number")
	}

	// Test WithWinner
	winner := TeamEastWest
	updated = original.WithWinner(winner)
	if original.Winner != nil {
		t.Error("Original match context should not have winner")
	}
	if updated.Winner == nil || *updated.Winner != winner {
		t.Error("Updated match context should have winner")
	}
	if updated.State != MatchStateFinished {
		t.Error("Updated match context should be finished")
	}
	if updated.EndTime == nil {
		t.Error("Updated match context should have end time")
	}
}

func TestMatchCtxIsFinished(t *testing.T) {
	players := []*Player{
		NewPlayer("p1", "Alice", SeatEast),
		NewPlayer("p2", "Bob", SeatSouth),
		NewPlayer("p3", "Charlie", SeatWest),
		NewPlayer("p4", "David", SeatNorth),
	}
	matchCtx := NewMatchCtx("test-match", players, 12345)

	if matchCtx.IsFinished() {
		t.Error("New match should not be finished")
	}

	finishedMatch := matchCtx.WithState(MatchStateFinished)
	if !finishedMatch.IsFinished() {
		t.Error("Match with finished state should be finished")
	}
}

func TestDealStateString(t *testing.T) {
	testCases := []struct {
		state    DealState
		expected string
	}{
		{DealStateCreated, "Created"},
		{DealStateDealt, "Dealt"},
		{DealStateTribute, "Tribute"},
		{DealStateFirstPlay, "FirstPlay"},
		{DealStateInProgress, "InProgress"},
		{DealStateFinished, "Finished"},
		{DealState(999), "Unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.state.String()
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestNewDealCtx(t *testing.T) {
	dealNumber := 3
	trump := King
	firstPlayer := SeatWest

	dealCtx := NewDealCtx(dealNumber, trump, firstPlayer)

	if dealCtx.DealNumber != dealNumber {
		t.Errorf("Expected deal number %d, got %d", dealNumber, dealCtx.DealNumber)
	}

	if dealCtx.State != DealStateCreated {
		t.Errorf("Expected state %v, got %v", DealStateCreated, dealCtx.State)
	}

	if dealCtx.Trump != trump {
		t.Errorf("Expected trump %v, got %v", trump, dealCtx.Trump)
	}

	if dealCtx.CurrentLevel != trump {
		t.Errorf("Expected current level %v, got %v", trump, dealCtx.CurrentLevel)
	}

	if dealCtx.FirstPlayer != firstPlayer {
		t.Errorf("Expected first player %v, got %v", firstPlayer, dealCtx.FirstPlayer)
	}

	if dealCtx.TrickCount != 0 {
		t.Errorf("Expected trick count 0, got %d", dealCtx.TrickCount)
	}

	if dealCtx.IsFirstDeal != (dealNumber == 1) {
		t.Errorf("Expected is first deal %v, got %v", dealNumber == 1, dealCtx.IsFirstDeal)
	}

	if dealCtx.TributeGiven {
		t.Error("Tribute should not be given initially")
	}

	if len(dealCtx.TributeCards) != 0 {
		t.Error("Tribute cards should be empty initially")
	}

	if len(dealCtx.RankList) != 0 {
		t.Error("Rank list should be empty initially")
	}

	if dealCtx.EndTime != nil {
		t.Error("End time should be nil initially")
	}

	if time.Since(dealCtx.StartTime) > time.Second {
		t.Error("Start time should be recent")
	}
}

func TestDealCtxImmutableUpdates(t *testing.T) {
	original := NewDealCtx(1, Two, SeatEast)

	// Test WithState
	updated := original.WithState(DealStateDealt)
	if original.State == DealStateDealt {
		t.Error("Original deal context should not be modified")
	}
	if updated.State != DealStateDealt {
		t.Error("Updated deal context should have new state")
	}

	// Test WithTrickCount
	updated = original.WithTrickCount(5)
	if original.TrickCount == 5 {
		t.Error("Original deal context should not be modified")
	}
	if updated.TrickCount != 5 {
		t.Error("Updated deal context should have new trick count")
	}

	// Test WithRankList
	rankList := []SeatID{SeatEast, SeatSouth}
	updated = original.WithRankList(rankList)
	if len(original.RankList) != 0 {
		t.Error("Original deal context should not be modified")
	}
	if len(updated.RankList) != 2 {
		t.Error("Updated deal context should have new rank list")
	}
	if updated.RankList[0] != SeatEast || updated.RankList[1] != SeatSouth {
		t.Error("Updated deal context should have correct rank list")
	}

	// Test WithTributeGiven
	updated = original.WithTributeGiven(true)
	if original.TributeGiven {
		t.Error("Original deal context should not be modified")
	}
	if !updated.TributeGiven {
		t.Error("Updated deal context should have tribute given")
	}

	// Test WithEndTime
	endTime := time.Now()
	updated = original.WithEndTime(endTime)
	if original.EndTime != nil {
		t.Error("Original deal context should not have end time")
	}
	if updated.EndTime == nil {
		t.Error("Updated deal context should have end time")
	}
	if updated.State != DealStateFinished {
		t.Error("Updated deal context should be finished")
	}
}

func TestDealCtxAddToRankList(t *testing.T) {
	dealCtx := NewDealCtx(1, Two, SeatEast)

	// Add first player
	updated := dealCtx.AddToRankList(SeatSouth)
	if len(dealCtx.RankList) != 0 {
		t.Error("Original deal context should not be modified")
	}
	if len(updated.RankList) != 1 {
		t.Error("Updated deal context should have one player in rank list")
	}
	if updated.RankList[0] != SeatSouth {
		t.Error("First player in rank list should be SeatSouth")
	}

	// Add second player
	updated2 := updated.AddToRankList(SeatWest)
	if len(updated.RankList) != 1 {
		t.Error("Previous deal context should not be modified")
	}
	if len(updated2.RankList) != 2 {
		t.Error("Updated deal context should have two players in rank list")
	}
	if updated2.RankList[1] != SeatWest {
		t.Error("Second player in rank list should be SeatWest")
	}
}

func TestDealCtxIsFinished(t *testing.T) {
	dealCtx := NewDealCtx(1, Two, SeatEast)

	if dealCtx.IsFinished() {
		t.Error("New deal should not be finished")
	}

	finished := dealCtx.WithState(DealStateFinished)
	if !finished.IsFinished() {
		t.Error("Deal with finished state should be finished")
	}
}

func TestDealCtxGetRankPosition(t *testing.T) {
	dealCtx := NewDealCtx(1, Two, SeatEast)
	rankList := []SeatID{SeatSouth, SeatWest, SeatNorth}
	updated := dealCtx.WithRankList(rankList)

	testCases := []struct {
		seat     SeatID
		expected int
	}{
		{SeatSouth, 1},
		{SeatWest, 2},
		{SeatNorth, 3},
		{SeatEast, 0}, // Not in list
	}

	for _, tc := range testCases {
		t.Run(tc.seat.String(), func(t *testing.T) {
			position := updated.GetRankPosition(tc.seat)
			if position != tc.expected {
				t.Errorf("Expected position %d for seat %v, got %d", tc.expected, tc.seat, position)
			}
		})
	}
}

func TestTrickStateString(t *testing.T) {
	testCases := []struct {
		state    TrickState
		expected string
	}{
		{TrickStateActive, "Active"},
		{TrickStateFinished, "Finished"},
		{TrickState(999), "Unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.state.String()
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestNewTrickCtx(t *testing.T) {
	trickNumber := 5
	startPlayer := SeatWest

	trickCtx := NewTrickCtx(trickNumber, startPlayer)

	if trickCtx.TrickNumber != trickNumber {
		t.Errorf("Expected trick number %d, got %d", trickNumber, trickCtx.TrickNumber)
	}

	if trickCtx.State != TrickStateActive {
		t.Errorf("Expected state %v, got %v", TrickStateActive, trickCtx.State)
	}

	if trickCtx.StartPlayer != startPlayer {
		t.Errorf("Expected start player %v, got %v", startPlayer, trickCtx.StartPlayer)
	}

	if trickCtx.CurrentPlayer != startPlayer {
		t.Errorf("Expected current player %v, got %v", startPlayer, trickCtx.CurrentPlayer)
	}

	if trickCtx.LastPlay != nil {
		t.Error("Last play should be nil initially")
	}

	if trickCtx.LastPlayer != SeatID(0) {
		t.Error("Last player should be zero value initially")
	}

	if len(trickCtx.PassedPlayers) != 0 {
		t.Error("Passed players should be empty initially")
	}

	if len(trickCtx.PlayHistory) != 0 {
		t.Error("Play history should be empty initially")
	}

	if trickCtx.Winner != SeatID(0) {
		t.Error("Winner should be zero value initially")
	}
}

func TestTrickCtxImmutableUpdates(t *testing.T) {
	original := NewTrickCtx(1, SeatEast)

	// Test WithCurrentPlayer
	updated := original.WithCurrentPlayer(SeatSouth)
	if original.CurrentPlayer != SeatEast {
		t.Error("Original trick context should not be modified")
	}
	if updated.CurrentPlayer != SeatSouth {
		t.Error("Updated trick context should have new current player")
	}

	// Test WithLastPlay
	cardGroup := NewCardGroup([]Card{NewCard(Hearts, Ace)})
	updated = original.WithLastPlay(cardGroup, SeatWest)
	if original.LastPlay != nil {
		t.Error("Original trick context should not be modified")
	}
	if updated.LastPlay != cardGroup {
		t.Error("Updated trick context should have new last play")
	}
	if updated.LastPlayer != SeatWest {
		t.Error("Updated trick context should have new last player")
	}

	// Test WithPlayerPassed
	updated = original.WithPlayerPassed(SeatNorth)
	if len(original.PassedPlayers) != 0 {
		t.Error("Original trick context should not be modified")
	}
	if !updated.HasPlayerPassed(SeatNorth) {
		t.Error("Updated trick context should have player passed")
	}

	// Test WithWinner
	updated = original.WithWinner(SeatEast)
	if original.Winner != SeatID(0) { // Check against zero value, not SeatEast
		t.Error("Original trick context should not be modified")
	}
	if updated.Winner != SeatEast {
		t.Error("Updated trick context should have winner")
	}
	if updated.State != TrickStateFinished {
		t.Error("Updated trick context should be finished")
	}
}

func TestTrickCtxPlayHistory(t *testing.T) {
	trickCtx := NewTrickCtx(1, SeatEast)

	play1 := TrickPlay{
		Player:    SeatEast,
		Cards:     []Card{NewCard(Hearts, Ace)},
		CardGroup: NewCardGroup([]Card{NewCard(Hearts, Ace)}),
		Timestamp: time.Now(),
	}

	updated := trickCtx.WithPlayHistory(play1)
	if len(trickCtx.PlayHistory) != 0 {
		t.Error("Original trick context should not be modified")
	}
	if len(updated.PlayHistory) != 1 {
		t.Error("Updated trick context should have one play in history")
	}
	if updated.PlayHistory[0].Player != SeatEast {
		t.Error("Play history should contain correct player")
	}

	play2 := TrickPlay{
		Player:    SeatSouth,
		Cards:     []Card{NewCard(Spades, King)},
		CardGroup: NewCardGroup([]Card{NewCard(Spades, King)}),
		Timestamp: time.Now(),
	}

	updated2 := updated.WithPlayHistory(play2)
	if len(updated.PlayHistory) != 1 {
		t.Error("Previous trick context should not be modified")
	}
	if len(updated2.PlayHistory) != 2 {
		t.Error("Updated trick context should have two plays in history")
	}
	if updated2.PlayHistory[1].Player != SeatSouth {
		t.Error("Second play in history should be from SeatSouth")
	}
}

func TestTrickCtxPlayerManagement(t *testing.T) {
	trickCtx := NewTrickCtx(1, SeatEast)

	// Initially no players have passed
	if trickCtx.HasPlayerPassed(SeatEast) {
		t.Error("No players should have passed initially")
	}

	if trickCtx.GetActivePlayerCount() != 4 {
		t.Error("All 4 players should be active initially")
	}

	// Pass one player
	updated := trickCtx.WithPlayerPassed(SeatSouth)
	if !updated.HasPlayerPassed(SeatSouth) {
		t.Error("SeatSouth should have passed")
	}
	if updated.HasPlayerPassed(SeatEast) {
		t.Error("SeatEast should not have passed")
	}
	if updated.GetActivePlayerCount() != 3 {
		t.Error("Should have 3 active players after one passes")
	}

	// Pass another player
	updated2 := updated.WithPlayerPassed(SeatWest)
	if updated2.GetActivePlayerCount() != 2 {
		t.Error("Should have 2 active players after two pass")
	}

	// Test ShouldFinish
	if updated2.ShouldFinish() {
		t.Error("Trick should not finish with 2 active players")
	}

	// Pass third player
	updated3 := updated2.WithPlayerPassed(SeatNorth)
	if updated3.GetActivePlayerCount() != 1 {
		t.Error("Should have 1 active player after three pass")
	}
	if !updated3.ShouldFinish() {
		t.Error("Trick should finish with only 1 active player")
	}
}

func TestTrickCtxIsFinished(t *testing.T) {
	trickCtx := NewTrickCtx(1, SeatEast)

	if trickCtx.IsFinished() {
		t.Error("New trick should not be finished")
	}

	finished := trickCtx.WithWinner(SeatEast)
	if !finished.IsFinished() {
		t.Error("Trick with winner should be finished")
	}
}

func TestTrickCtxGetNextPlayer(t *testing.T) {
	testCases := []struct {
		current SeatID
		next    SeatID
	}{
		{SeatEast, SeatSouth},
		{SeatSouth, SeatWest},
		{SeatWest, SeatNorth},
		{SeatNorth, SeatEast},
	}

	for _, tc := range testCases {
		t.Run(tc.current.String(), func(t *testing.T) {
			trickCtx := NewTrickCtx(1, tc.current)
			next := trickCtx.GetNextPlayer()
			if next != tc.next {
				t.Errorf("Expected next player %v, got %v", tc.next, next)
			}
		})
	}
}

func TestContextsIndependence(t *testing.T) {
	// Test that contexts don't interfere with each other
	players := []*Player{
		NewPlayer("p1", "Alice", SeatEast),
		NewPlayer("p2", "Bob", SeatSouth),
		NewPlayer("p3", "Charlie", SeatWest),
		NewPlayer("p4", "David", SeatNorth),
	}

	matchCtx1 := NewMatchCtx("match1", players, 111)
	matchCtx2 := NewMatchCtx("match2", players, 222)

	if matchCtx1.ID == matchCtx2.ID {
		t.Error("Different match contexts should have different IDs")
	}

	if matchCtx1.Seed == matchCtx2.Seed {
		t.Error("Different match contexts should have different seeds")
	}

	dealCtx1 := NewDealCtx(1, Two, SeatEast)
	dealCtx2 := NewDealCtx(2, Three, SeatWest)

	if dealCtx1.DealNumber == dealCtx2.DealNumber {
		t.Error("Different deal contexts should have different deal numbers")
	}

	if dealCtx1.Trump == dealCtx2.Trump {
		t.Error("Different deal contexts should have different trumps")
	}

	trickCtx1 := NewTrickCtx(1, SeatEast)
	trickCtx2 := NewTrickCtx(2, SeatWest)

	if trickCtx1.TrickNumber == trickCtx2.TrickNumber {
		t.Error("Different trick contexts should have different trick numbers")
	}

	if trickCtx1.StartPlayer == trickCtx2.StartPlayer {
		t.Error("Different trick contexts should have different start players")
	}
}

func BenchmarkNewMatchCtx(b *testing.B) {
	players := []*Player{
		NewPlayer("p1", "Alice", SeatEast),
		NewPlayer("p2", "Bob", SeatSouth),
		NewPlayer("p3", "Charlie", SeatWest),
		NewPlayer("p4", "David", SeatNorth),
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewMatchCtx("test-match", players, int64(i))
	}
}

func BenchmarkDealCtxWithState(b *testing.B) {
	dealCtx := NewDealCtx(1, Two, SeatEast)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = dealCtx.WithState(DealStateInProgress)
	}
}

func BenchmarkTrickCtxWithPlayerPassed(b *testing.B) {
	trickCtx := NewTrickCtx(1, SeatEast)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = trickCtx.WithPlayerPassed(SeatSouth)
	}
}