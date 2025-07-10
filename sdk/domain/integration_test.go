package domain

import (
	"math/rand"
	"testing"
	"time"
)

// Integration tests that verify components work together correctly

func TestFullGameSimulation(t *testing.T) {
	// Simulate a complete game flow
	players := []*Player{
		NewPlayer("p1", "Alice", SeatEast),
		NewPlayer("p2", "Bob", SeatSouth),
		NewPlayer("p3", "Charlie", SeatWest),
		NewPlayer("p4", "David", SeatNorth),
	}

	matchCtx := NewMatchCtx("test-match", players, 12345)
	
	// Start first deal
	dealCtx := NewDealCtx(1, Two, SeatEast)
	
	// Deal cards
	deck := NewDeckWithSeed(matchCtx.Seed)
	deck.Shuffle()
	hands := deck.DealToHands(4)
	
	// Distribute cards to players
	for i, hand := range hands {
		seat := SeatID(i)
		player := matchCtx.GetPlayer(seat)
		if player != nil {
			player.ClearHand()
			player.AddCards(hand)
		}
	}
	
	// Verify all players have cards
	for seat := SeatEast; seat <= SeatNorth; seat++ {
		player := matchCtx.GetPlayer(seat)
		if player == nil {
			t.Errorf("Player at seat %v not found", seat)
			continue
		}
		
		if player.HandSize() == 0 {
			t.Errorf("Player %s has no cards", player.Name)
		}
		
		// Verify player has exactly 27 cards (108/4)
		if player.HandSize() != 27 {
			t.Errorf("Player %s should have 27 cards, got %d", player.Name, player.HandSize())
		}
	}
	
	// Test trick simulation
	trickCtx := NewTrickCtx(1, dealCtx.FirstPlayer)
	
	// Simulate players playing cards in order
	playersWhoPlayed := 0
	for round := 0; round < 4; round++ {
		currentPlayer := matchCtx.GetPlayer(trickCtx.CurrentPlayer)
		if currentPlayer == nil {
			t.Fatalf("Current player not found: %v", trickCtx.CurrentPlayer)
		}
		
		// Player plays a single card (first card in hand)
		if currentPlayer.HandSize() > 0 {
			cardToPlay := currentPlayer.GetHand()[0]
			cards := []Card{cardToPlay}
			
			// Verify player has the card
			if !currentPlayer.HasCards(cards) {
				t.Errorf("Player %s should have card %v", currentPlayer.Name, cardToPlay)
			}
			
			// Create card group and verify it's valid
			cardGroup := NewCardGroup(cards)
			if !cardGroup.IsValid() {
				t.Errorf("Single card should be valid: %v", cardToPlay)
			}
			
			// For first play or if can follow, player plays
			if trickCtx.LastPlay == nil || CanFollow(cardGroup, trickCtx.LastPlay, dealCtx.Trump) {
				// Remove card from player's hand
				if !currentPlayer.RemoveCards(cards) {
					t.Errorf("Failed to remove card from player %s", currentPlayer.Name)
				}
				
				// Update trick context
				trickCtx = trickCtx.WithLastPlay(cardGroup, trickCtx.CurrentPlayer)
				trickCtx = trickCtx.WithCurrentPlayer(trickCtx.CurrentPlayer.Next())
				
				// Add to play history
				play := TrickPlay{
					Player:    currentPlayer.SeatID,
					Cards:     cards,
					CardGroup: cardGroup,
					Timestamp: time.Now(),
				}
				trickCtx = trickCtx.WithPlayHistory(play)
				playersWhoPlayed++
			} else {
				// Player passes
				trickCtx = trickCtx.WithPlayerPassed(trickCtx.CurrentPlayer)
				trickCtx = trickCtx.WithCurrentPlayer(trickCtx.CurrentPlayer.Next())
			}
		}
	}
	
	// Verify trick history
	if len(trickCtx.PlayHistory) == 0 {
		t.Error("Trick should have play history")
	}
	
	// Verify all players have one less card
	for seat := SeatEast; seat <= SeatNorth; seat++ {
		player := matchCtx.GetPlayer(seat)
		if player != nil && player.HandSize() != 26 {
			t.Errorf("Player %s should have 26 cards after playing one, got %d", player.Name, player.HandSize())
		}
	}
}

func TestComplexCardGroupInteractions(t *testing.T) {
	// Test complex interactions between card groups and comparison logic
	testCases := []struct {
		name        string
		cards1      []Card
		cards2      []Card
		trump       Rank
		expectWin1  bool
		expectEqual bool
	}{
		{
			name:        "Bomb vs straight",
			cards1:      []Card{NewCard(Hearts, King), NewCard(Spades, King), NewCard(Clubs, King), NewCard(Diamonds, King)},
			cards2:      []Card{NewCard(Hearts, Three), NewCard(Spades, Four), NewCard(Clubs, Five), NewCard(Diamonds, Six), NewCard(Hearts, Seven)},
			trump:       Two,
			expectWin1:  true,
			expectEqual: false,
		},
		{
			name:        "Trump bomb vs normal bomb",
			cards1:      []Card{NewCard(Hearts, Two), NewCard(Spades, Two), NewCard(Clubs, Two), NewCard(Diamonds, Two)},
			cards2:      []Card{NewCard(Hearts, Ace), NewCard(Spades, Ace), NewCard(Clubs, Ace), NewCard(Diamonds, Ace)},
			trump:       Two,
			expectWin1:  true,
			expectEqual: false,
		},
		{
			name:        "Joker bomb vs trump bomb",
			cards1:      []Card{NewJoker(SmallJoker), NewJoker(BigJoker), NewJoker(SmallJoker)},
			cards2:      []Card{NewCard(Hearts, Two), NewCard(Spades, Two), NewCard(Clubs, Two), NewCard(Diamonds, Two)},
			trump:       Two,
			expectWin1:  true,
			expectEqual: false,
		},
		{
			name:        "Same category different ranks",
			cards1:      []Card{NewCard(Hearts, Ace), NewCard(Spades, Ace)},
			cards2:      []Card{NewCard(Hearts, King), NewCard(Spades, King)},
			trump:       Two,
			expectWin1:  true,
			expectEqual: false,
		},
		{
			name:        "Incomparable groups",
			cards1:      []Card{NewCard(Hearts, Ace)},
			cards2:      []Card{NewCard(Hearts, King), NewCard(Spades, King)},
			trump:       Two,
			expectWin1:  false,
			expectEqual: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			group1 := NewCardGroup(tc.cards1)
			group2 := NewCardGroup(tc.cards2)
			
			if !group1.IsValid() {
				t.Errorf("Group 1 should be valid: %v", tc.cards1)
			}
			if !group2.IsValid() {
				t.Errorf("Group 2 should be valid: %v", tc.cards2)
			}
			
			result := CompareCardGroups(group1, group2, tc.trump)
			
			if tc.expectWin1 && result != CmpGreater {
				t.Errorf("Group 1 should win, got %v", result)
			} else if tc.expectEqual && result != CmpEqual {
				t.Errorf("Groups should be equal/incomparable, got %v", result)
			} else if !tc.expectWin1 && !tc.expectEqual && result != CmpLess {
				t.Errorf("Group 2 should win, got %v", result)
			}
		})
	}
}

func TestPlayerTeamIntegration(t *testing.T) {
	// Test player-team integration across all components
	players := []*Player{
		NewPlayer("east", "East Player", SeatEast),
		NewPlayer("south", "South Player", SeatSouth),
		NewPlayer("west", "West Player", SeatWest),
		NewPlayer("north", "North Player", SeatNorth),
	}
	
	matchCtx := NewMatchCtx("team-test", players, 54321)
	
	// Verify team assignments
	eastWestTeam := matchCtx.GetTeam(TeamEastWest)
	southNorthTeam := matchCtx.GetTeam(TeamSouthNorth)
	
	if eastWestTeam == nil || southNorthTeam == nil {
		t.Fatal("Teams not properly initialized")
	}
	
	eastWestPlayers := eastWestTeam.GetPlayers()
	southNorthPlayers := southNorthTeam.GetPlayers()
	
	if len(eastWestPlayers) != 2 {
		t.Errorf("East-West team should have 2 players, got %d", len(eastWestPlayers))
	}
	if len(southNorthPlayers) != 2 {
		t.Errorf("South-North team should have 2 players, got %d", len(southNorthPlayers))
	}
	
	// Verify player-partner relationships
	eastPlayer := matchCtx.GetPlayer(SeatEast)
	westPlayer := matchCtx.GetPlayer(SeatWest)
	
	if eastPlayer.GetPartnerSeat() != SeatWest {
		t.Error("East player's partner should be West")
	}
	if westPlayer.GetPartnerSeat() != SeatEast {
		t.Error("West player's partner should be East")
	}
	
	// Verify team lookup by seat
	if eastWestTeam.GetPlayerBySeat(SeatEast) != eastPlayer {
		t.Error("Team should find correct player by seat")
	}
	if eastWestTeam.GetPlayerBySeat(SeatSouth) != nil {
		t.Error("Team should not find player from different team")
	}
	
	// Test opposing teams
	if eastWestTeam.ID.OpposingTeam() != TeamSouthNorth {
		t.Error("East-West opposing team should be South-North")
	}
	
	// Deal cards and verify team-level operations
	deck := NewDeckWithSeed(matchCtx.Seed)
	deck.Shuffle()
	hands := deck.DealToHands(4)
	
	for i, hand := range hands {
		seat := SeatID(i)
		player := matchCtx.GetPlayer(seat)
		player.AddCards(hand)
	}
	
	// Count total cards per team
	eastWestCardCount := 0
	southNorthCardCount := 0
	
	for _, player := range eastWestPlayers {
		eastWestCardCount += player.HandSize()
	}
	for _, player := range southNorthPlayers {
		southNorthCardCount += player.HandSize()
	}
	
	if eastWestCardCount != 54 { // 108/2 = 54 cards per team
		t.Errorf("East-West team should have 54 cards total, got %d", eastWestCardCount)
	}
	if southNorthCardCount != 54 {
		t.Errorf("South-North team should have 54 cards total, got %d", southNorthCardCount)
	}
}

func TestContextStateMachineSimulation(t *testing.T) {
	// Test state transitions across all context types
	players := []*Player{
		NewPlayer("p1", "Player1", SeatEast),
		NewPlayer("p2", "Player2", SeatSouth),
		NewPlayer("p3", "Player3", SeatWest),
		NewPlayer("p4", "Player4", SeatNorth),
	}
	
	// Match lifecycle
	matchCtx := NewMatchCtx("state-test", players, 98765)
	if matchCtx.State != MatchStateCreated {
		t.Error("New match should be in Created state")
	}
	
	// Start match
	matchCtx = matchCtx.WithState(MatchStateInProgress)
	if matchCtx.State != MatchStateInProgress {
		t.Error("Match should be in InProgress state")
	}
	
	// Deal lifecycle
	dealCtx := NewDealCtx(1, Three, SeatSouth)
	if dealCtx.State != DealStateCreated {
		t.Error("New deal should be in Created state")
	}
	
	// Progress through deal states
	dealCtx = dealCtx.WithState(DealStateDealt)
	dealCtx = dealCtx.WithState(DealStateTribute)
	dealCtx = dealCtx.WithState(DealStateFirstPlay)
	dealCtx = dealCtx.WithState(DealStateInProgress)
	
	// Simulate players finishing
	dealCtx = dealCtx.AddToRankList(SeatNorth)  // First to finish
	dealCtx = dealCtx.AddToRankList(SeatEast)   // Second
	dealCtx = dealCtx.AddToRankList(SeatSouth)  // Third
	
	if len(dealCtx.RankList) != 3 {
		t.Errorf("Expected 3 players in rank list, got %d", len(dealCtx.RankList))
	}
	
	if dealCtx.GetRankPosition(SeatNorth) != 1 {
		t.Error("North should be in first position")
	}
	if dealCtx.GetRankPosition(SeatWest) != 0 {
		t.Error("West should not be in rank list yet")
	}
	
	// Finish deal
	dealCtx = dealCtx.WithState(DealStateFinished)
	dealCtx = dealCtx.WithEndTime(time.Now())
	
	if !dealCtx.IsFinished() {
		t.Error("Deal should be finished")
	}
	
	// Trick lifecycle simulation
	trickCtx := NewTrickCtx(1, SeatEast)
	if trickCtx.State != TrickStateActive {
		t.Error("New trick should be active")
	}
	
	// Simulate turn progression
	trickCtx = trickCtx.WithCurrentPlayer(SeatSouth)
	trickCtx = trickCtx.WithCurrentPlayer(SeatWest)
	trickCtx = trickCtx.WithCurrentPlayer(SeatNorth)
	
	// Players pass
	trickCtx = trickCtx.WithPlayerPassed(SeatSouth)
	trickCtx = trickCtx.WithPlayerPassed(SeatWest)
	trickCtx = trickCtx.WithPlayerPassed(SeatNorth)
	
	if trickCtx.GetActivePlayerCount() != 1 {
		t.Errorf("Expected 1 active player, got %d", trickCtx.GetActivePlayerCount())
	}
	
	if !trickCtx.ShouldFinish() {
		t.Error("Trick should be ready to finish")
	}
	
	// Finish trick
	trickCtx = trickCtx.WithWinner(SeatEast)
	if !trickCtx.IsFinished() {
		t.Error("Trick should be finished")
	}
	
	// Complete match
	matchCtx = matchCtx.WithWinner(TeamEastWest)
	if !matchCtx.IsFinished() {
		t.Error("Match should be finished")
	}
	if matchCtx.Winner == nil || *matchCtx.Winner != TeamEastWest {
		t.Error("Match winner should be East-West team")
	}
}

// Stress tests for performance and memory usage

func TestStressCardGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}
	
	// Generate many cards and verify no memory leaks or panics
	const iterations = 10000
	
	for i := 0; i < iterations; i++ {
		// Create various card combinations
		suits := []Suit{Hearts, Diamonds, Clubs, Spades, Joker}
		ranks := []Rank{Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King, Ace, SmallJoker, BigJoker}
		
		var cards []Card
		for j := 0; j < 50; j++ { // 50 random cards per iteration
			suit := suits[rand.Intn(len(suits))]
			rank := ranks[rand.Intn(len(ranks))]
			
			// Skip invalid combinations
			if suit == Joker && rank != SmallJoker && rank != BigJoker {
				continue
			}
			if suit != Joker && (rank == SmallJoker || rank == BigJoker) {
				continue
			}
			
			var card Card
			if suit == Joker {
				card = NewJoker(rank)
			} else {
				card = NewCard(suit, rank)
			}
			
			cards = append(cards, card)
		}
		
		// Test card groups with these cards
		if len(cards) > 0 {
			group := NewCardGroup(cards[:min(len(cards), 10)])
			_ = group.IsValid()
			_ = group.IsBomb()
		}
	}
}

func TestStressDeckOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}
	
	const iterations = 1000
	
	for i := 0; i < iterations; i++ {
		deck := NewDeckWithSeed(int64(i))
		deck.Shuffle()
		
		// Deal cards in various ways
		_ = deck.Deal(10)
		_ = deck.Deal(20)
		_ = deck.Deal(30)
		
		// Reset and try different patterns
		deck.Reset()
		hands := deck.DealToHands(4)
		if len(hands) != 4 {
			t.Errorf("Iteration %d: Expected 4 hands, got %d", i, len(hands))
		}
		
		// Verify total cards
		totalCards := 0
		for _, hand := range hands {
			totalCards += len(hand)
		}
		if totalCards != 108 {
			t.Errorf("Iteration %d: Expected 108 total cards, got %d", i, totalCards)
		}
	}
}

func TestStressPlayerOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}
	
	const iterations = 1000
	
	for i := 0; i < iterations; i++ {
		players := []*Player{
			NewPlayer("p1", "Player1", SeatEast),
			NewPlayer("p2", "Player2", SeatSouth),
			NewPlayer("p3", "Player3", SeatWest),
			NewPlayer("p4", "Player4", SeatNorth),
		}
		
		// Generate random cards for each player
		for _, player := range players {
			var cards []Card
			for j := 0; j < 27; j++ { // Standard hand size
				suit := Suit(rand.Intn(4)) // Hearts to Spades
				rank := Rank(rand.Intn(13) + 1) // Two to Ace
				cards = append(cards, NewCard(suit, rank))
			}
			
			player.AddCards(cards)
			
			// Test various operations
			if player.HandSize() != 27 {
				t.Errorf("Iteration %d: Player should have 27 cards, got %d", i, player.HandSize())
			}
			
			// Remove some cards
			toRemove := cards[:5]
			if !player.RemoveCards(toRemove) {
				t.Errorf("Iteration %d: Should be able to remove cards", i)
			}
			
			if player.HandSize() != 22 {
				t.Errorf("Iteration %d: Player should have 22 cards after removal, got %d", i, player.HandSize())
			}
			
			// Test card checking
			for _, card := range toRemove {
				if player.HasCard(card) {
					t.Errorf("Iteration %d: Player should not have removed card", i)
				}
			}
		}
	}
}

func TestStressComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}
	
	const iterations = 5000
	
	for i := 0; i < iterations; i++ {
		// Generate random cards
		trump := Rank(rand.Intn(13) + 1) // Two to Ace
		
		card1 := NewCard(Suit(rand.Intn(4)), Rank(rand.Intn(13)+1))
		card2 := NewCard(Suit(rand.Intn(4)), Rank(rand.Intn(13)+1))
		
		// Test card comparison
		result1 := CompareCards(card1, card2, trump)
		result2 := CompareCards(card2, card1, trump)
		
		// Verify anti-symmetry
		if result1 == CmpGreater && result2 != CmpLess {
			t.Errorf("Iteration %d: Comparison should be anti-symmetric", i)
		}
		if result1 == CmpLess && result2 != CmpGreater {
			t.Errorf("Iteration %d: Comparison should be anti-symmetric", i)
		}
		if result1 == CmpEqual && result2 != CmpEqual {
			t.Errorf("Iteration %d: Equal comparison should be symmetric", i)
		}
		
		// Test reflexivity
		selfResult := CompareCards(card1, card1, trump)
		if selfResult != CmpEqual {
			t.Errorf("Iteration %d: Card should be equal to itself", i)
		}
	}
}

// Helper function for min (Go 1.21+ has this in slices package)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}