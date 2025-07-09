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