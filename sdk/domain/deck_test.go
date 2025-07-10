package domain

import (
	"testing"
)

func TestDeckInitialization(t *testing.T) {
	deck := NewDeck()
	
	if deck.Size() != 108 {
		t.Errorf("Expected deck size 108, got %d", deck.Size())
	}
	
	if deck.IsEmpty() {
		t.Error("New deck should not be empty")
	}
	
	if deck.Remaining() != 108 {
		t.Errorf("Expected 108 remaining cards, got %d", deck.Remaining())
	}
}

func TestDeckWithSeed(t *testing.T) {
	seed := int64(12345)
	deck1 := NewDeckWithSeed(seed)
	deck2 := NewDeckWithSeed(seed)
	
	deck1.Shuffle()
	deck2.Shuffle()
	
	for i := 0; i < 10; i++ {
		card1 := deck1.GetCard(i)
		card2 := deck2.GetCard(i)
		
		if card1.Suit != card2.Suit || card1.Rank != card2.Rank {
			t.Errorf("Cards at position %d should be identical with same seed", i)
		}
	}
}

func TestDeckCardCounts(t *testing.T) {
	deck := NewDeck()
	
	suitCounts := make(map[Suit]int)
	rankCounts := make(map[Rank]int)
	
	for i := 0; i < deck.Size(); i++ {
		card := deck.GetCard(i)
		suitCounts[card.Suit]++
		rankCounts[card.Rank]++
	}
	
	for suit := Hearts; suit <= Spades; suit++ {
		if suitCounts[suit] != 26 {
			t.Errorf("Expected 26 cards of suit %v, got %d", suit, suitCounts[suit])
		}
	}
	
	if suitCounts[Joker] != 4 {
		t.Errorf("Expected 4 jokers, got %d", suitCounts[Joker])
	}
	
	for rank := Ace; rank <= King; rank++ {
		if rankCounts[rank] != 8 {
			t.Errorf("Expected 8 cards of rank %v, got %d", rank, rankCounts[rank])
		}
	}
	
	if rankCounts[SmallJoker] != 2 {
		t.Errorf("Expected 2 small jokers, got %d", rankCounts[SmallJoker])
	}
	
	if rankCounts[BigJoker] != 2 {
		t.Errorf("Expected 2 big jokers, got %d", rankCounts[BigJoker])
	}
}

func TestDeckDeal(t *testing.T) {
	deck := NewDeck()
	originalSize := deck.Size()
	
	dealt := deck.Deal(10)
	
	if len(dealt) != 10 {
		t.Errorf("Expected 10 dealt cards, got %d", len(dealt))
	}
	
	if deck.Size() != originalSize-10 {
		t.Errorf("Expected deck size %d, got %d", originalSize-10, deck.Size())
	}
	
	if deck.Remaining() != originalSize-10 {
		t.Errorf("Expected %d remaining cards, got %d", originalSize-10, deck.Remaining())
	}
}

func TestDeckDealToHands(t *testing.T) {
	deck := NewDeck()
	
	hands := deck.DealToHands(4)
	
	if len(hands) != 4 {
		t.Errorf("Expected 4 hands, got %d", len(hands))
	}
	
	expectedCardsPerHand := 108 / 4
	for i, hand := range hands {
		if len(hand) != expectedCardsPerHand {
			t.Errorf("Expected %d cards in hand %d, got %d", expectedCardsPerHand, i, len(hand))
		}
	}
	
	if deck.Remaining() != 0 {
		t.Errorf("Expected no remaining cards, got %d", deck.Remaining())
	}
	
	if !deck.IsEmpty() {
		t.Error("Deck should be empty after dealing all cards")
	}
}

func TestDeckDealToHandsInvalidPlayerCount(t *testing.T) {
	deck := NewDeck()
	
	testCases := []int{0, -1, 5, 10}
	
	for _, playerCount := range testCases {
		hands := deck.DealToHands(playerCount)
		if hands != nil {
			t.Errorf("Expected nil for invalid player count %d, got %v", playerCount, hands)
		}
	}
}

func TestDeckReset(t *testing.T) {
	deck := NewDeck()
	
	deck.Deal(50)
	
	if deck.Size() == 108 {
		t.Error("Deck size should have changed after dealing")
	}
	
	deck.Reset()
	
	if deck.Size() != 108 {
		t.Errorf("Expected deck size 108 after reset, got %d", deck.Size())
	}
	
	if deck.IsEmpty() {
		t.Error("Deck should not be empty after reset")
	}
}

func TestDeckGetCardBounds(t *testing.T) {
	deck := NewDeck()
	
	validCard := deck.GetCard(0)
	if validCard.Suit == 0 && validCard.Rank == 0 {
		t.Error("Valid index should return a real card")
	}
	
	invalidCard := deck.GetCard(-1)
	if invalidCard.Suit != 0 || invalidCard.Rank != 0 {
		t.Error("Invalid negative index should return zero card")
	}
	
	invalidCard = deck.GetCard(200)
	if invalidCard.Suit != 0 || invalidCard.Rank != 0 {
		t.Error("Invalid large index should return zero card")
	}
}

func TestDeckShuffle(t *testing.T) {
	deck1 := NewDeck()
	deck2 := NewDeck()
	
	originalOrder1 := make([]Card, deck1.Size())
	for i := 0; i < deck1.Size(); i++ {
		originalOrder1[i] = deck1.GetCard(i)
	}
	
	originalOrder2 := make([]Card, deck2.Size())
	for i := 0; i < deck2.Size(); i++ {
		originalOrder2[i] = deck2.GetCard(i)
	}
	
	deck1.Shuffle()
	
	differentPositions := 0
	for i := 0; i < deck1.Size(); i++ {
		card1 := deck1.GetCard(i)
		card2 := originalOrder1[i]
		if card1.Suit != card2.Suit || card1.Rank != card2.Rank {
			differentPositions++
		}
	}
	
	if differentPositions < 50 {
		t.Errorf("Shuffle should change many positions, only %d changed", differentPositions)
	}
}

func TestDeckShuffleWithSeed(t *testing.T) {
	deck := NewDeck()
	seed := int64(98765)
	
	originalOrder := make([]Card, deck.Size())
	for i := 0; i < deck.Size(); i++ {
		originalOrder[i] = deck.GetCard(i)
	}
	
	deck.ShuffleWithSeed(seed)
	
	firstShuffleOrder := make([]Card, deck.Size())
	for i := 0; i < deck.Size(); i++ {
		firstShuffleOrder[i] = deck.GetCard(i)
	}
	
	deck.Reset()
	deck.ShuffleWithSeed(seed)
	
	for i := 0; i < deck.Size(); i++ {
		card1 := firstShuffleOrder[i]
		card2 := deck.GetCard(i)
		if card1.Suit != card2.Suit || card1.Rank != card2.Rank {
			t.Errorf("Cards at position %d should be identical when shuffled with same seed", i)
		}
	}
}

// Additional edge case tests for Deck

func TestDeckEdgeCases(t *testing.T) {
	t.Run("Deal more cards than available", func(t *testing.T) {
		deck := NewDeck()
		originalSize := deck.Size()
		
		// Try to deal more cards than available
		dealt := deck.Deal(originalSize + 10)
		
		if len(dealt) != originalSize {
			t.Errorf("Expected to deal %d cards (all available), got %d", originalSize, len(dealt))
		}
		
		if !deck.IsEmpty() {
			t.Error("Deck should be empty after dealing all cards")
		}
	})
	
	t.Run("Deal from empty deck", func(t *testing.T) {
		deck := NewDeck()
		deck.Deal(deck.Size()) // Empty the deck
		
		dealt := deck.Deal(5)
		if len(dealt) != 0 {
			t.Errorf("Expected 0 cards from empty deck, got %d", len(dealt))
		}
	})
	
	t.Run("Deal zero cards", func(t *testing.T) {
		deck := NewDeck()
		originalSize := deck.Size()
		
		dealt := deck.Deal(0)
		if len(dealt) != 0 {
			t.Errorf("Expected 0 cards when dealing 0, got %d", len(dealt))
		}
		
		if deck.Size() != originalSize {
			t.Error("Deck size should not change when dealing 0 cards")
		}
	})
	
	t.Run("Deal negative cards", func(t *testing.T) {
		deck := NewDeck()
		originalSize := deck.Size()
		
		dealt := deck.Deal(-5)
		if len(dealt) != 0 {
			t.Errorf("Expected 0 cards when dealing negative amount, got %d", len(dealt))
		}
		
		if deck.Size() != originalSize {
			t.Error("Deck size should not change when dealing negative amount")
		}
	})
}

func TestDeckMultipleResets(t *testing.T) {
	deck := NewDeck()
	
	for i := 0; i < 5; i++ {
		// Deal some cards
		deck.Deal(20)
		
		// Reset and verify
		deck.Reset()
		
		if deck.Size() != 108 {
			t.Errorf("Reset %d: Expected size 108, got %d", i+1, deck.Size())
		}
		
		if deck.IsEmpty() {
			t.Errorf("Reset %d: Deck should not be empty after reset", i+1)
		}
	}
}

func TestDeckCardDistribution(t *testing.T) {
	deck := NewDeck()
	
	// Verify we have exactly the right distribution
	suitCounts := make(map[Suit]int)
	rankCounts := make(map[Rank]int)
	
	for i := 0; i < deck.Size(); i++ {
		card := deck.GetCard(i)
		suitCounts[card.Suit]++
		rankCounts[card.Rank]++
	}
	
	// Each normal suit should have 26 cards (2 decks * 13 ranks)
	for suit := Hearts; suit <= Spades; suit++ {
		if suitCounts[suit] != 26 {
			t.Errorf("Suit %v should have 26 cards, got %d", suit, suitCounts[suit])
		}
	}
	
	// Joker suit should have 4 cards (2 small + 2 big)
	if suitCounts[Joker] != 4 {
		t.Errorf("Joker suit should have 4 cards, got %d", suitCounts[Joker])
	}
	
	// Each normal rank should have 8 cards (2 decks * 4 suits)
	for rank := Two; rank <= Ace; rank++ {
		if rankCounts[rank] != 8 {
			t.Errorf("Rank %v should have 8 cards, got %d", rank, rankCounts[rank])
		}
	}
	
	// Each joker rank should have 2 cards
	if rankCounts[SmallJoker] != 2 {
		t.Errorf("SmallJoker should have 2 cards, got %d", rankCounts[SmallJoker])
	}
	if rankCounts[BigJoker] != 2 {
		t.Errorf("BigJoker should have 2 cards, got %d", rankCounts[BigJoker])
	}
}

func TestDeckDealToHandsEdgeCases(t *testing.T) {
	t.Run("Uneven distribution", func(t *testing.T) {
		deck := NewDeck()
		
		// 108 cards / 3 players = 36 cards each
		hands := deck.DealToHands(3)
		
		if len(hands) != 3 {
			t.Errorf("Expected 3 hands, got %d", len(hands))
		}
		
		for i, hand := range hands {
			if len(hand) != 36 {
				t.Errorf("Hand %d should have 36 cards, got %d", i, len(hand))
			}
		}
		
		if !deck.IsEmpty() {
			t.Error("Deck should be empty after dealing to 3 players")
		}
	})
	
	t.Run("More players than cards allow", func(t *testing.T) {
		deck := NewDeck()
		
		// 108 cards / 200 players = 0 cards each
		hands := deck.DealToHands(200)
		
		if hands != nil {
			t.Error("Should return nil for too many players")
		}
	})
	
	t.Run("Single player gets all cards", func(t *testing.T) {
		deck := NewDeck()
		originalSize := deck.Size()
		
		hands := deck.DealToHands(1)
		
		if len(hands) != 1 {
			t.Errorf("Expected 1 hand, got %d", len(hands))
		}
		
		if len(hands[0]) != originalSize {
			t.Errorf("Single player should get all %d cards, got %d", originalSize, len(hands[0]))
		}
		
		if !deck.IsEmpty() {
			t.Error("Deck should be empty after dealing all cards to one player")
		}
	})
}

func TestDeckShuffleConsistency(t *testing.T) {
	// Verify that different seeds produce different orders
	seed1 := int64(12345)
	seed2 := int64(54321)
	
	deck1 := NewDeckWithSeed(seed1)
	deck2 := NewDeckWithSeed(seed2)
	
	deck1.Shuffle()
	deck2.Shuffle()
	
	differentPositions := 0
	for i := 0; i < 10; i++ { // Check first 10 cards
		card1 := deck1.GetCard(i)
		card2 := deck2.GetCard(i)
		if card1.Suit != card2.Suit || card1.Rank != card2.Rank {
			differentPositions++
		}
	}
	
	if differentPositions == 0 {
		t.Error("Different seeds should produce different shuffles")
	}
}

func TestDeckMemoryIsolation(t *testing.T) {
	// Test that dealt cards are independent copies
	deck := NewDeck()
	
	hand1 := deck.Deal(5)
	_ = deck.Deal(5) // Second hand (we don't need to examine it)
	
	// Modify hand1 (this should not affect deck state)
	originalCard := hand1[0]
	hand1[0] = NewCard(Hearts, Ace) // Modify the slice
	
	// Subsequent deals should be unaffected
	hand2Second := deck.Deal(5)
	
	// The modification to hand1 should not affect subsequent deals
	for i, card := range hand2Second {
		if card.Suit == 0 && card.Rank == 0 {
			t.Errorf("Card %d in second deal appears to be zero value", i)
		}
	}
	
	// Restore for verification
	hand1[0] = originalCard
}

func TestDeckConcurrentAccess(t *testing.T) {
	// Basic test to ensure deck operations don't panic under concurrent access
	// Note: This doesn't test for race conditions, just basic safety
	deck := NewDeck()
	
	done := make(chan bool, 2)
	
	// Goroutine 1: Deal cards
	go func() {
		for i := 0; i < 10; i++ {
			deck.Deal(1)
		}
		done <- true
	}()
	
	// Goroutine 2: Check size
	go func() {
		for i := 0; i < 10; i++ {
			_ = deck.Size()
			_ = deck.Remaining()
			_ = deck.IsEmpty()
		}
		done <- true
	}()
	
	// Wait for both goroutines
	<-done
	<-done
}

// Benchmark tests
func BenchmarkDeckShuffle(b *testing.B) {
	deck := NewDeck()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		deck.Reset()
		deck.Shuffle()
	}
}

func BenchmarkDeckDeal(b *testing.B) {
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		deck := NewDeck()
		_ = deck.Deal(27) // Deal quarter of deck
	}
}

func BenchmarkDeckDealToHands(b *testing.B) {
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		deck := NewDeck()
		_ = deck.DealToHands(4)
	}
}