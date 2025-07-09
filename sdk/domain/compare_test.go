package domain

import (
	"testing"
)

func TestCardCompare(t *testing.T) {
	testCases := []struct {
		name     string
		cardA    Card
		cardB    Card
		trump    Rank
		expected CmpResult
	}{
		{
			name:     "Same cards are equal",
			cardA:    NewCard(Hearts, Ace),
			cardB:    NewCard(Hearts, Ace),
			trump:    Two,
			expected: CmpEqual,
		},
		{
			name:     "Ace beats King",
			cardA:    NewCard(Hearts, Ace),
			cardB:    NewCard(Hearts, King),
			trump:    Two,
			expected: CmpGreater,
		},
		{
			name:     "King loses to Ace",
			cardA:    NewCard(Hearts, King),
			cardB:    NewCard(Hearts, Ace),
			trump:    Two,
			expected: CmpLess,
		},
		{
			name:     "Trump beats non-trump",
			cardA:    NewCard(Hearts, Two),
			cardB:    NewCard(Hearts, Ace),
			trump:    Two,
			expected: CmpGreater,
		},
		{
			name:     "Non-trump loses to trump",
			cardA:    NewCard(Hearts, Ace),
			cardB:    NewCard(Hearts, Two),
			trump:    Two,
			expected: CmpLess,
		},
		{
			name:     "Big joker beats small joker",
			cardA:    NewJoker(BigJoker),
			cardB:    NewJoker(SmallJoker),
			trump:    Two,
			expected: CmpGreater,
		},
		{
			name:     "Small joker loses to big joker",
			cardA:    NewJoker(SmallJoker),
			cardB:    NewJoker(BigJoker),
			trump:    Two,
			expected: CmpLess,
		},
		{
			name:     "Joker beats trump",
			cardA:    NewJoker(SmallJoker),
			cardB:    NewCard(Hearts, Two),
			trump:    Two,
			expected: CmpGreater,
		},
		{
			name:     "Trump suits matter for same rank",
			cardA:    NewCard(Hearts, Two),
			cardB:    NewCard(Spades, Two),
			trump:    Two,
			expected: CmpLess,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CompareCards(tc.cardA, tc.cardB, tc.trump)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for %v vs %v with trump %v", 
					tc.expected, result, tc.cardA, tc.cardB, tc.trump)
			}
		})
	}
}

func TestCardGroupCompare(t *testing.T) {
	testCases := []struct {
		name     string
		groupA   *CardGroup
		groupB   *CardGroup
		trump    Rank
		expected CmpResult
	}{
		{
			name:     "Single card comparison",
			groupA:   NewCardGroup([]Card{NewCard(Hearts, Ace)}),
			groupB:   NewCardGroup([]Card{NewCard(Hearts, King)}),
			trump:    Two,
			expected: CmpGreater,
		},
		{
			name:     "Pair comparison",
			groupA:   NewCardGroup([]Card{NewCard(Hearts, Ace), NewCard(Spades, Ace)}),
			groupB:   NewCardGroup([]Card{NewCard(Hearts, King), NewCard(Spades, King)}),
			trump:    Two,
			expected: CmpGreater,
		},
		{
			name:     "Bomb beats normal play",
			groupA:   NewCardGroup([]Card{NewCard(Hearts, King), NewCard(Spades, King), NewCard(Clubs, King), NewCard(Diamonds, King)}),
			groupB:   NewCardGroup([]Card{NewCard(Hearts, Ace)}),
			trump:    Two,
			expected: CmpGreater,
		},
		{
			name:     "Joker bomb beats normal bomb",
			groupA:   NewCardGroup([]Card{NewJoker(SmallJoker), NewJoker(BigJoker)}),
			groupB:   NewCardGroup([]Card{NewCard(Hearts, King), NewCard(Spades, King), NewCard(Clubs, King), NewCard(Diamonds, King)}),
			trump:    Two,
			expected: CmpGreater,
		},
		{
			name:     "Larger joker bomb beats smaller joker bomb",
			groupA:   NewCardGroup([]Card{NewJoker(SmallJoker), NewJoker(BigJoker), NewJoker(SmallJoker)}),
			groupB:   NewCardGroup([]Card{NewJoker(SmallJoker), NewJoker(BigJoker)}),
			trump:    Two,
			expected: CmpGreater,
		},
		{
			name:     "Different categories cannot be compared",
			groupA:   NewCardGroup([]Card{NewCard(Hearts, Ace)}),
			groupB:   NewCardGroup([]Card{NewCard(Hearts, King), NewCard(Spades, King)}),
			trump:    Two,
			expected: CmpEqual,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CompareCardGroups(tc.groupA, tc.groupB, tc.trump)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for %v vs %v with trump %v", 
					tc.expected, result, tc.groupA, tc.groupB, tc.trump)
			}
		})
	}
}

func TestCanBeat(t *testing.T) {
	testCases := []struct {
		name      string
		hand      *CardGroup
		tablePlay *CardGroup
		trump     Rank
		expected  bool
	}{
		{
			name:      "Can beat with higher single",
			hand:      NewCardGroup([]Card{NewCard(Hearts, Ace)}),
			tablePlay: NewCardGroup([]Card{NewCard(Hearts, King)}),
			trump:     Two,
			expected:  true,
		},
		{
			name:      "Cannot beat with lower single",
			hand:      NewCardGroup([]Card{NewCard(Hearts, King)}),
			tablePlay: NewCardGroup([]Card{NewCard(Hearts, Ace)}),
			trump:     Two,
			expected:  false,
		},
		{
			name:      "Can beat with bomb",
			hand:      NewCardGroup([]Card{NewCard(Hearts, King), NewCard(Spades, King), NewCard(Clubs, King), NewCard(Diamonds, King)}),
			tablePlay: NewCardGroup([]Card{NewCard(Hearts, Ace)}),
			trump:     Two,
			expected:  true,
		},
		{
			name:      "Can play on empty table",
			hand:      NewCardGroup([]Card{NewCard(Hearts, King)}),
			tablePlay: nil,
			trump:     Two,
			expected:  true,
		},
		{
			name:      "Cannot play invalid hand",
			hand:      nil,
			tablePlay: NewCardGroup([]Card{NewCard(Hearts, King)}),
			trump:     Two,
			expected:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CanBeat(tc.hand, tc.tablePlay, tc.trump)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for hand %v vs table %v with trump %v", 
					tc.expected, result, tc.hand, tc.tablePlay, tc.trump)
			}
		})
	}
}

func TestCanFollow(t *testing.T) {
	testCases := []struct {
		name      string
		hand      *CardGroup
		tablePlay *CardGroup
		trump     Rank
		expected  bool
	}{
		{
			name:      "Can follow with same category and higher rank",
			hand:      NewCardGroup([]Card{NewCard(Hearts, Ace)}),
			tablePlay: NewCardGroup([]Card{NewCard(Hearts, King)}),
			trump:     Two,
			expected:  true,
		},
		{
			name:      "Cannot follow with different category",
			hand:      NewCardGroup([]Card{NewCard(Hearts, Ace)}),
			tablePlay: NewCardGroup([]Card{NewCard(Hearts, King), NewCard(Spades, King)}),
			trump:     Two,
			expected:  false,
		},
		{
			name:      "Can follow with bomb on any play",
			hand:      NewCardGroup([]Card{NewCard(Hearts, King), NewCard(Spades, King), NewCard(Clubs, King), NewCard(Diamonds, King)}),
			tablePlay: NewCardGroup([]Card{NewCard(Hearts, Ace)}),
			trump:     Two,
			expected:  true,
		},
		{
			name:      "Can follow with bomb on pair",
			hand:      NewCardGroup([]Card{NewCard(Hearts, King), NewCard(Spades, King), NewCard(Clubs, King), NewCard(Diamonds, King)}),
			tablePlay: NewCardGroup([]Card{NewCard(Hearts, Ace), NewCard(Spades, Ace)}),
			trump:     Two,
			expected:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CanFollow(tc.hand, tc.tablePlay, tc.trump)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for hand %v vs table %v with trump %v", 
					tc.expected, result, tc.hand, tc.tablePlay, tc.trump)
			}
		})
	}
}

func TestIsTrump(t *testing.T) {
	testCases := []struct {
		name     string
		card     Card
		trump    Rank
		expected bool
	}{
		{
			name:     "Normal trump card",
			card:     NewCard(Hearts, Two),
			trump:    Two,
			expected: true,
		},
		{
			name:     "Non-trump card",
			card:     NewCard(Hearts, Ace),
			trump:    Two,
			expected: false,
		},
		{
			name:     "Small joker is always trump",
			card:     NewJoker(SmallJoker),
			trump:    Two,
			expected: true,
		},
		{
			name:     "Big joker is always trump",
			card:     NewJoker(BigJoker),
			trump:    Two,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsTrump(tc.card, tc.trump)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for card %v with trump %v", 
					tc.expected, result, tc.card, tc.trump)
			}
		})
	}
}

func TestCountTrumps(t *testing.T) {
	cards := []Card{
		NewCard(Hearts, Two),
		NewCard(Spades, Two),
		NewCard(Hearts, Ace),
		NewJoker(SmallJoker),
		NewJoker(BigJoker),
	}
	
	count := CountTrumps(cards, Two)
	expected := 4  // Two hearts, two spades, small joker, big joker
	
	if count != expected {
		t.Errorf("Expected %d trumps, got %d", expected, count)
	}
}

func TestFindBombs(t *testing.T) {
	cards := []Card{
		NewCard(Hearts, King),
		NewCard(Spades, King),
		NewCard(Clubs, King),
		NewCard(Diamonds, King),
		NewCard(Hearts, Ace),
		NewJoker(SmallJoker),
		NewJoker(BigJoker),
	}
	
	bombs := FindBombs(cards, Two)
	
	if len(bombs) < 2 {
		t.Errorf("Expected at least 2 bombs, got %d", len(bombs))
	}
	
	foundNormalBomb := false
	foundJokerBomb := false
	
	for _, bomb := range bombs {
		if bomb.Category == Bomb {
			foundNormalBomb = true
		}
		if bomb.Category == JokerBomb {
			foundJokerBomb = true
		}
	}
	
	if !foundNormalBomb {
		t.Error("Expected to find normal bomb")
	}
	
	if !foundJokerBomb {
		t.Error("Expected to find joker bomb")
	}
}