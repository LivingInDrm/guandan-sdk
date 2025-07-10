package domain

import (
	"testing"
)

func TestSeatIDString(t *testing.T) {
	testCases := []struct {
		seat     SeatID
		expected string
	}{
		{SeatEast, "East"},
		{SeatSouth, "South"},
		{SeatWest, "West"},
		{SeatNorth, "North"},
		{SeatID(999), "Unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.seat.String()
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestSeatIDNext(t *testing.T) {
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
			result := tc.current.Next()
			if result != tc.next {
				t.Errorf("Expected next seat %v, got %v", tc.next, result)
			}
		})
	}
}

func TestSeatIDPrevious(t *testing.T) {
	testCases := []struct {
		current  SeatID
		previous SeatID
	}{
		{SeatEast, SeatNorth},
		{SeatSouth, SeatEast},
		{SeatWest, SeatSouth},
		{SeatNorth, SeatWest},
	}

	for _, tc := range testCases {
		t.Run(tc.current.String(), func(t *testing.T) {
			result := tc.current.Previous()
			if result != tc.previous {
				t.Errorf("Expected previous seat %v, got %v", tc.previous, result)
			}
		})
	}
}

func TestSeatIDOpposite(t *testing.T) {
	testCases := []struct {
		current  SeatID
		opposite SeatID
	}{
		{SeatEast, SeatWest},
		{SeatSouth, SeatNorth},
		{SeatWest, SeatEast},
		{SeatNorth, SeatSouth},
	}

	for _, tc := range testCases {
		t.Run(tc.current.String(), func(t *testing.T) {
			result := tc.current.Opposite()
			if result != tc.opposite {
				t.Errorf("Expected opposite seat %v, got %v", tc.opposite, result)
			}
		})
	}
}

func TestSeatIDIsValid(t *testing.T) {
	testCases := []struct {
		seat     SeatID
		expected bool
	}{
		{SeatEast, true},
		{SeatSouth, true},
		{SeatWest, true},
		{SeatNorth, true},
		{SeatID(-1), false},
		{SeatID(4), false},
		{SeatID(999), false},
	}

	for _, tc := range testCases {
		t.Run("Seat_"+tc.seat.String(), func(t *testing.T) {
			result := tc.seat.IsValid()
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for seat %v", tc.expected, result, tc.seat)
			}
		})
	}
}

func TestTeamIDString(t *testing.T) {
	testCases := []struct {
		team     TeamID
		expected string
	}{
		{TeamEastWest, "East-West"},
		{TeamSouthNorth, "South-North"},
		{TeamID(999), "Unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.team.String()
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestTeamIDOpposingTeam(t *testing.T) {
	testCases := []struct {
		team     TeamID
		opposing TeamID
	}{
		{TeamEastWest, TeamSouthNorth},
		{TeamSouthNorth, TeamEastWest},
	}

	for _, tc := range testCases {
		t.Run(tc.team.String(), func(t *testing.T) {
			result := tc.team.OpposingTeam()
			if result != tc.opposing {
				t.Errorf("Expected opposing team %v, got %v", tc.opposing, result)
			}
		})
	}
}

func TestGetTeamFromSeat(t *testing.T) {
	testCases := []struct {
		seat SeatID
		team TeamID
	}{
		{SeatEast, TeamEastWest},
		{SeatWest, TeamEastWest},
		{SeatSouth, TeamSouthNorth},
		{SeatNorth, TeamSouthNorth},
	}

	for _, tc := range testCases {
		t.Run(tc.seat.String(), func(t *testing.T) {
			result := GetTeamFromSeat(tc.seat)
			if result != tc.team {
				t.Errorf("Expected team %v for seat %v, got %v", tc.team, tc.seat, result)
			}
		})
	}
}

func TestNewPlayer(t *testing.T) {
	id := "player123"
	name := "Alice"
	seat := SeatEast

	player := NewPlayer(id, name, seat)

	if player.ID != id {
		t.Errorf("Expected ID %s, got %s", id, player.ID)
	}

	if player.Name != name {
		t.Errorf("Expected name %s, got %s", name, player.Name)
	}

	if player.SeatID != seat {
		t.Errorf("Expected seat %v, got %v", seat, player.SeatID)
	}

	expectedTeam := GetTeamFromSeat(seat)
	if player.TeamID != expectedTeam {
		t.Errorf("Expected team %v, got %v", expectedTeam, player.TeamID)
	}

	if player.Level != Two {
		t.Errorf("Expected level %v, got %v", Two, player.Level)
	}

	if len(player.Hand) != 0 {
		t.Error("New player should have empty hand")
	}

	if !player.IsOnline {
		t.Error("New player should be online")
	}
}

func TestPlayerString(t *testing.T) {
	player := NewPlayer("p1", "Alice", SeatEast)
	expected := "Alice(East)"
	result := player.String()

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestPlayerGetPartnerSeat(t *testing.T) {
	testCases := []struct {
		seat    SeatID
		partner SeatID
	}{
		{SeatEast, SeatWest},
		{SeatSouth, SeatNorth},
		{SeatWest, SeatEast},
		{SeatNorth, SeatSouth},
	}

	for _, tc := range testCases {
		t.Run(tc.seat.String(), func(t *testing.T) {
			player := NewPlayer("p1", "Test", tc.seat)
			partner := player.GetPartnerSeat()
			if partner != tc.partner {
				t.Errorf("Expected partner seat %v, got %v", tc.partner, partner)
			}
		})
	}
}

func TestPlayerAddCards(t *testing.T) {
	player := NewPlayer("p1", "Alice", SeatEast)

	if player.HandSize() != 0 {
		t.Error("New player should have empty hand")
	}

	cards := []Card{
		NewCard(Hearts, Ace),
		NewCard(Spades, King),
		NewCard(Diamonds, Queen),
	}

	player.AddCards(cards)

	if player.HandSize() != 3 {
		t.Errorf("Expected hand size 3, got %d", player.HandSize())
	}

	hand := player.GetHand()
	if len(hand) != 3 {
		t.Errorf("Expected hand length 3, got %d", len(hand))
	}

	// Verify specific cards
	for i, expectedCard := range cards {
		if hand[i].Suit != expectedCard.Suit || hand[i].Rank != expectedCard.Rank {
			t.Errorf("Expected card %v at position %d, got %v", expectedCard, i, hand[i])
		}
	}

	// Add more cards
	moreCards := []Card{NewCard(Clubs, Jack)}
	player.AddCards(moreCards)

	if player.HandSize() != 4 {
		t.Errorf("Expected hand size 4 after adding more cards, got %d", player.HandSize())
	}
}

func TestPlayerRemoveCards(t *testing.T) {
	player := NewPlayer("p1", "Alice", SeatEast)
	
	cards := []Card{
		NewCard(Hearts, Ace),
		NewCard(Spades, King),
		NewCard(Diamonds, Queen),
		NewCard(Clubs, Jack),
	}
	player.AddCards(cards)

	// Remove existing cards
	toRemove := []Card{NewCard(Hearts, Ace), NewCard(Clubs, Jack)}
	success := player.RemoveCards(toRemove)

	if !success {
		t.Error("Should successfully remove existing cards")
	}

	if player.HandSize() != 2 {
		t.Errorf("Expected hand size 2 after removal, got %d", player.HandSize())
	}

	// Verify remaining cards
	if player.HasCard(NewCard(Hearts, Ace)) {
		t.Error("Should not have removed card")
	}

	if player.HasCard(NewCard(Clubs, Jack)) {
		t.Error("Should not have removed card")
	}

	if !player.HasCard(NewCard(Spades, King)) {
		t.Error("Should still have non-removed card")
	}

	if !player.HasCard(NewCard(Diamonds, Queen)) {
		t.Error("Should still have non-removed card")
	}

	// Try to remove non-existing card
	nonExisting := []Card{NewCard(Hearts, Two)}
	success = player.RemoveCards(nonExisting)

	if success {
		t.Error("Should fail to remove non-existing card")
	}

	if player.HandSize() != 2 {
		t.Error("Hand size should not change when removal fails")
	}
}

func TestPlayerHasCard(t *testing.T) {
	player := NewPlayer("p1", "Alice", SeatEast)
	
	card := NewCard(Hearts, Ace)
	player.AddCards([]Card{card})

	if !player.HasCard(card) {
		t.Error("Player should have the added card")
	}

	differentCard := NewCard(Spades, King)
	if player.HasCard(differentCard) {
		t.Error("Player should not have card that wasn't added")
	}
}

func TestPlayerHasCards(t *testing.T) {
	player := NewPlayer("p1", "Alice", SeatEast)
	
	cards := []Card{
		NewCard(Hearts, Ace),
		NewCard(Spades, King),
		NewCard(Diamonds, Queen),
	}
	player.AddCards(cards)

	// Test with subset of cards
	subset := []Card{NewCard(Hearts, Ace), NewCard(Spades, King)}
	if !player.HasCards(subset) {
		t.Error("Player should have all cards in subset")
	}

	// Test with one missing card
	withMissing := []Card{
		NewCard(Hearts, Ace),
		NewCard(Clubs, Jack), // This one is missing
	}
	if player.HasCards(withMissing) {
		t.Error("Player should not have all cards when one is missing")
	}

	// Test with empty slice
	if !player.HasCards([]Card{}) {
		t.Error("Player should have all cards in empty slice")
	}
}

func TestPlayerHandManagement(t *testing.T) {
	player := NewPlayer("p1", "Alice", SeatEast)

	if !player.IsHandEmpty() {
		t.Error("New player should have empty hand")
	}

	cards := []Card{
		NewCard(Hearts, Ace),
		NewCard(Spades, King),
	}
	player.AddCards(cards)

	if player.IsHandEmpty() {
		t.Error("Player with cards should not have empty hand")
	}

	if player.HandSize() != 2 {
		t.Errorf("Expected hand size 2, got %d", player.HandSize())
	}

	player.ClearHand()

	if !player.IsHandEmpty() {
		t.Error("Player should have empty hand after clearing")
	}

	if player.HandSize() != 0 {
		t.Errorf("Expected hand size 0 after clearing, got %d", player.HandSize())
	}
}

func TestPlayerGetHandImmutable(t *testing.T) {
	player := NewPlayer("p1", "Alice", SeatEast)
	
	originalCards := []Card{
		NewCard(Hearts, Ace),
		NewCard(Spades, King),
	}
	player.AddCards(originalCards)

	// Get hand copy
	handCopy := player.GetHand()

	// Modify the copy
	handCopy[0] = NewCard(Diamonds, Queen)

	// Original hand should be unchanged
	actualHand := player.GetHand()
	if actualHand[0].Suit != Hearts || actualHand[0].Rank != Ace {
		t.Error("Original hand should not be affected by modifications to copy")
	}
}

func TestNewTeam(t *testing.T) {
	teamID := TeamEastWest
	team := NewTeam(teamID)

	if team.ID != teamID {
		t.Errorf("Expected team ID %v, got %v", teamID, team.ID)
	}

	if team.Level != Two {
		t.Errorf("Expected level %v, got %v", Two, team.Level)
	}

	if team.Players[0] != nil || team.Players[1] != nil {
		t.Error("New team should have no players")
	}
}

func TestTeamAddPlayer(t *testing.T) {
	team := NewTeam(TeamEastWest)

	player1 := NewPlayer("p1", "Alice", SeatEast)
	player2 := NewPlayer("p2", "Bob", SeatWest)
	player3 := NewPlayer("p3", "Charlie", SeatSouth) // Wrong team

	// Add first player
	success := team.AddPlayer(player1)
	if !success {
		t.Error("Should successfully add first player")
	}

	players := team.GetPlayers()
	if len(players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(players))
	}

	// Add second player
	success = team.AddPlayer(player2)
	if !success {
		t.Error("Should successfully add second player")
	}

	players = team.GetPlayers()
	if len(players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(players))
	}

	// Try to add third player (should fail)
	success = team.AddPlayer(player3)
	if success {
		t.Error("Should fail to add third player")
	}

	players = team.GetPlayers()
	if len(players) != 2 {
		t.Error("Player count should not change when addition fails")
	}
}

func TestTeamGetPlayerBySeat(t *testing.T) {
	team := NewTeam(TeamEastWest)

	player1 := NewPlayer("p1", "Alice", SeatEast)
	player2 := NewPlayer("p2", "Bob", SeatWest)

	team.AddPlayer(player1)
	team.AddPlayer(player2)

	// Find existing players
	foundPlayer := team.GetPlayerBySeat(SeatEast)
	if foundPlayer == nil {
		t.Error("Should find player at SeatEast")
	}
	if foundPlayer.ID != "p1" {
		t.Error("Should return correct player")
	}

	foundPlayer = team.GetPlayerBySeat(SeatWest)
	if foundPlayer == nil {
		t.Error("Should find player at SeatWest")
	}
	if foundPlayer.ID != "p2" {
		t.Error("Should return correct player")
	}

	// Try to find non-existing player
	foundPlayer = team.GetPlayerBySeat(SeatSouth)
	if foundPlayer != nil {
		t.Error("Should not find player at SeatSouth")
	}
}

func TestTeamString(t *testing.T) {
	team := NewTeam(TeamEastWest)
	expected := "East-West"
	result := team.String()

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestNewPlayerArray(t *testing.T) {
	playerArray := NewPlayerArray()

	for seat := SeatEast; seat <= SeatNorth; seat++ {
		if playerArray.Get(seat) != nil {
			t.Errorf("New player array should have nil at seat %v", seat)
		}
	}

	if playerArray.IsComplete() {
		t.Error("New player array should not be complete")
	}

	all := playerArray.All()
	if len(all) != 0 {
		t.Error("New player array should return empty slice for All()")
	}
}

func TestPlayerArraySetGet(t *testing.T) {
	playerArray := NewPlayerArray()

	player1 := NewPlayer("p1", "Alice", SeatEast)
	player2 := NewPlayer("p2", "Bob", SeatSouth)

	playerArray.Set(SeatEast, player1)
	playerArray.Set(SeatSouth, player2)

	// Test valid gets
	retrieved := playerArray.Get(SeatEast)
	if retrieved == nil {
		t.Error("Should retrieve player at SeatEast")
	}
	if retrieved.ID != "p1" {
		t.Error("Should retrieve correct player")
	}

	retrieved = playerArray.Get(SeatSouth)
	if retrieved == nil {
		t.Error("Should retrieve player at SeatSouth")
	}
	if retrieved.ID != "p2" {
		t.Error("Should retrieve correct player")
	}

	// Test empty seats
	retrieved = playerArray.Get(SeatWest)
	if retrieved != nil {
		t.Error("Should return nil for empty seat")
	}

	// Test invalid seat
	retrieved = playerArray.Get(SeatID(999))
	if retrieved != nil {
		t.Error("Should return nil for invalid seat")
	}
}

func TestPlayerArrayInvalidSeat(t *testing.T) {
	playerArray := NewPlayerArray()
	player := NewPlayer("p1", "Alice", SeatEast)

	// Try to set invalid seat (should not crash)
	playerArray.Set(SeatID(999), player)

	// Verify it wasn't actually set
	all := playerArray.All()
	if len(all) != 0 {
		t.Error("Invalid seat should not be stored")
	}
}

func TestPlayerArrayIsComplete(t *testing.T) {
	playerArray := NewPlayerArray()

	if playerArray.IsComplete() {
		t.Error("Empty player array should not be complete")
	}

	// Add players one by one
	seats := []SeatID{SeatEast, SeatSouth, SeatWest, SeatNorth}
	for i, seat := range seats {
		player := NewPlayer("p"+string(rune('1'+i)), "Player"+string(rune('1'+i)), seat)
		playerArray.Set(seat, player)

		if i < 3 {
			if playerArray.IsComplete() {
				t.Errorf("Player array should not be complete with only %d players", i+1)
			}
		} else {
			if !playerArray.IsComplete() {
				t.Error("Player array should be complete with all 4 players")
			}
		}
	}
}

func TestPlayerArrayAll(t *testing.T) {
	playerArray := NewPlayerArray()

	// Add some players
	player1 := NewPlayer("p1", "Alice", SeatEast)
	player2 := NewPlayer("p2", "Bob", SeatSouth)
	player3 := NewPlayer("p3", "Charlie", SeatWest)

	playerArray.Set(SeatEast, player1)
	playerArray.Set(SeatSouth, player2)
	playerArray.Set(SeatWest, player3)

	all := playerArray.All()
	if len(all) != 3 {
		t.Errorf("Expected 3 players, got %d", len(all))
	}

	// Check that all players are present
	playerIDs := make(map[string]bool)
	for _, player := range all {
		playerIDs[player.ID] = true
	}

	expectedIDs := []string{"p1", "p2", "p3"}
	for _, expectedID := range expectedIDs {
		if !playerIDs[expectedID] {
			t.Errorf("Player %s not found in All() result", expectedID)
		}
	}
}

func TestPlayerTeamConsistency(t *testing.T) {
	// Test that players are assigned to correct teams
	eastPlayer := NewPlayer("east", "East Player", SeatEast)
	westPlayer := NewPlayer("west", "West Player", SeatWest)
	southPlayer := NewPlayer("south", "South Player", SeatSouth)
	northPlayer := NewPlayer("north", "North Player", SeatNorth)

	if eastPlayer.TeamID != TeamEastWest {
		t.Error("East player should be on East-West team")
	}
	if westPlayer.TeamID != TeamEastWest {
		t.Error("West player should be on East-West team")
	}
	if southPlayer.TeamID != TeamSouthNorth {
		t.Error("South player should be on South-North team")
	}
	if northPlayer.TeamID != TeamSouthNorth {
		t.Error("North player should be on South-North team")
	}

	// Test partner relationships
	if eastPlayer.GetPartnerSeat() != SeatWest {
		t.Error("East player's partner should be West")
	}
	if southPlayer.GetPartnerSeat() != SeatNorth {
		t.Error("South player's partner should be North")
	}
}

func TestPlayerCardOperationsEdgeCases(t *testing.T) {
	player := NewPlayer("p1", "Alice", SeatEast)

	// Test removing from empty hand
	success := player.RemoveCards([]Card{NewCard(Hearts, Ace)})
	if success {
		t.Error("Should fail to remove card from empty hand")
	}

	// Test removing empty slice
	player.AddCards([]Card{NewCard(Hearts, Ace)})
	success = player.RemoveCards([]Card{})
	if !success {
		t.Error("Should succeed to remove empty card slice")
	}
	if player.HandSize() != 1 {
		t.Error("Hand size should not change when removing empty slice")
	}

	// Test adding empty slice
	originalSize := player.HandSize()
	player.AddCards([]Card{})
	if player.HandSize() != originalSize {
		t.Error("Hand size should not change when adding empty slice")
	}

	// Test duplicate cards
	duplicateCard := NewCard(Hearts, Ace)
	player.AddCards([]Card{duplicateCard})
	if player.HandSize() != 2 {
		t.Error("Should be able to add duplicate cards")
	}

	// Remove one instance of duplicate
	success = player.RemoveCards([]Card{duplicateCard})
	if !success {
		t.Error("Should successfully remove one instance of duplicate card")
	}
	if player.HandSize() != 1 {
		t.Error("Should have one instance left after removing duplicate")
	}
	if !player.HasCard(duplicateCard) {
		t.Error("Should still have one instance of the card")
	}
}

func BenchmarkPlayerAddCards(b *testing.B) {
	player := NewPlayer("p1", "Test", SeatEast)
	cards := []Card{
		NewCard(Hearts, Ace),
		NewCard(Spades, King),
		NewCard(Diamonds, Queen),
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		player.ClearHand()
		player.AddCards(cards)
	}
}

func BenchmarkPlayerRemoveCards(b *testing.B) {
	player := NewPlayer("p1", "Test", SeatEast)
	cards := []Card{
		NewCard(Hearts, Ace),
		NewCard(Spades, King),
		NewCard(Diamonds, Queen),
	}
	toRemove := []Card{NewCard(Hearts, Ace)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		player.ClearHand()
		player.AddCards(cards)
		player.RemoveCards(toRemove)
	}
}

func BenchmarkPlayerHasCards(b *testing.B) {
	player := NewPlayer("p1", "Test", SeatEast)
	cards := []Card{
		NewCard(Hearts, Ace),
		NewCard(Spades, King),
		NewCard(Diamonds, Queen),
		NewCard(Clubs, Jack),
		NewCard(Hearts, Ten),
	}
	player.AddCards(cards)

	checkCards := []Card{NewCard(Hearts, Ace), NewCard(Spades, King)}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = player.HasCards(checkCards)
	}
}

func BenchmarkSeatIDNext(b *testing.B) {
	seat := SeatEast
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = seat.Next()
	}
}