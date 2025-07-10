package domain

import (
	"testing"
)

func TestCardGroupAnalysis(t *testing.T) {
	testCases := []struct {
		name             string
		cards            []Card
		expectedCategory CardCategory
		expectedRank     Rank
		expectedValid    bool
	}{
		{
			name:             "Single card",
			cards:            []Card{NewCard(Hearts, Ace)},
			expectedCategory: Single,
			expectedRank:     Ace,
			expectedValid:    true,
		},
		{
			name:             "Valid pair",
			cards:            []Card{NewCard(Hearts, Ace), NewCard(Spades, Ace)},
			expectedCategory: Pair,
			expectedRank:     Ace,
			expectedValid:    true,
		},
		{
			name:             "Invalid pair (different ranks)",
			cards:            []Card{NewCard(Hearts, Ace), NewCard(Spades, King)},
			expectedCategory: InvalidCategory,
			expectedRank:     0,
			expectedValid:    false,
		},
		{
			name:             "Valid triple",
			cards:            []Card{NewCard(Hearts, King), NewCard(Spades, King), NewCard(Clubs, King)},
			expectedCategory: Triple,
			expectedRank:     King,
			expectedValid:    true,
		},
		{
			name:             "Valid bomb (four of a kind)",
			cards:            []Card{NewCard(Hearts, King), NewCard(Spades, King), NewCard(Clubs, King), NewCard(Diamonds, King)},
			expectedCategory: Bomb,
			expectedRank:     King,
			expectedValid:    true,
		},
		{
			name:             "Valid joker bomb",
			cards:            []Card{NewJoker(SmallJoker), NewJoker(BigJoker)},
			expectedCategory: JokerBomb,
			expectedRank:     0,
			expectedValid:    true,
		},
		{
			name:             "Valid straight",
			cards:            []Card{NewCard(Hearts, Three), NewCard(Spades, Four), NewCard(Clubs, Five), NewCard(Diamonds, Six), NewCard(Hearts, Seven)},
			expectedCategory: Straight,
			expectedRank:     Three,
			expectedValid:    true,
		},
		{
			name:             "Invalid straight (too short)",
			cards:            []Card{NewCard(Hearts, Three), NewCard(Spades, Four), NewCard(Clubs, Five), NewCard(Diamonds, Six)},
			expectedCategory: InvalidCategory,
			expectedRank:     0,
			expectedValid:    false,
		},
		{
			name:             "Valid pair straight",
			cards: []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three),
				NewCard(Clubs, Four), NewCard(Diamonds, Four),
				NewCard(Hearts, Five), NewCard(Spades, Five),
			},
			expectedCategory: PairStraight,
			expectedRank:     Three,
			expectedValid:    true,
		},
		{
			name:             "Valid triple straight",
			cards: []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three), NewCard(Clubs, Three),
				NewCard(Diamonds, Four), NewCard(Hearts, Four), NewCard(Spades, Four),
			},
			expectedCategory: TripleStraight,
			expectedRank:     Three,
			expectedValid:    true,
		},
		{
			name:             "Empty cards",
			cards:            []Card{},
			expectedCategory: InvalidCategory,
			expectedRank:     0,
			expectedValid:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			group := NewCardGroup(tc.cards)
			
			if group.Category != tc.expectedCategory {
				t.Errorf("Expected category %v, got %v", tc.expectedCategory, group.Category)
			}
			
			if group.Rank != tc.expectedRank {
				t.Errorf("Expected rank %v, got %v", tc.expectedRank, group.Rank)
			}
			
			if group.IsValid() != tc.expectedValid {
				t.Errorf("Expected valid %v, got %v", tc.expectedValid, group.IsValid())
			}
		})
	}
}

func TestCardGroupIsBomb(t *testing.T) {
	testCases := []struct {
		name     string
		cards    []Card
		expected bool
	}{
		{
			name:     "Normal bomb",
			cards:    []Card{NewCard(Hearts, King), NewCard(Spades, King), NewCard(Clubs, King), NewCard(Diamonds, King)},
			expected: true,
		},
		{
			name:     "Joker bomb",
			cards:    []Card{NewJoker(SmallJoker), NewJoker(BigJoker)},
			expected: true,
		},
		{
			name:     "Not a bomb",
			cards:    []Card{NewCard(Hearts, Ace)},
			expected: false,
		},
		{
			name:     "Pair is not a bomb",
			cards:    []Card{NewCard(Hearts, King), NewCard(Spades, King)},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			group := NewCardGroup(tc.cards)
			result := group.IsBomb()
			
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for cards %v", tc.expected, result, tc.cards)
			}
		})
	}
}

func TestCardGroupComparisonKey(t *testing.T) {
	group := NewCardGroup([]Card{NewCard(Hearts, Ace), NewCard(Spades, Ace)})
	key := group.ComparisonKey()
	
	if key.Category != Pair {
		t.Errorf("Expected category Pair, got %v", key.Category)
	}
	
	if key.Size != 2 {
		t.Errorf("Expected size 2, got %v", key.Size)
	}
	
	if key.Rank != Ace {
		t.Errorf("Expected rank Ace, got %v", key.Rank)
	}
}

func TestStraightValidation(t *testing.T) {
	testCases := []struct {
		name     string
		cards    []Card
		expected bool
	}{
		{
			name: "Valid 5-card straight",
			cards: []Card{
				NewCard(Hearts, Three),
				NewCard(Spades, Four),
				NewCard(Clubs, Five),
				NewCard(Diamonds, Six),
				NewCard(Hearts, Seven),
			},
			expected: true,
		},
		{
			name: "Valid 6-card straight",
			cards: []Card{
				NewCard(Hearts, Three),
				NewCard(Spades, Four),
				NewCard(Clubs, Five),
				NewCard(Diamonds, Six),
				NewCard(Hearts, Seven),
				NewCard(Spades, Eight),
			},
			expected: true,
		},
		{
			name: "Invalid straight with gap",
			cards: []Card{
				NewCard(Hearts, Three),
				NewCard(Spades, Four),
				NewCard(Clubs, Six),
				NewCard(Diamonds, Seven),
				NewCard(Hearts, Eight),
			},
			expected: false,
		},
		{
			name: "Invalid straight with duplicate",
			cards: []Card{
				NewCard(Hearts, Three),
				NewCard(Spades, Three),
				NewCard(Clubs, Four),
				NewCard(Diamonds, Five),
				NewCard(Hearts, Six),
			},
			expected: false,
		},
		{
			name: "Too short for straight",
			cards: []Card{
				NewCard(Hearts, Three),
				NewCard(Spades, Four),
				NewCard(Clubs, Five),
				NewCard(Diamonds, Six),
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			group := NewCardGroup(tc.cards)
			result := (group.Category == Straight)
			
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for cards %v", tc.expected, result, tc.cards)
			}
		})
	}
}

func TestPairStraightValidation(t *testing.T) {
	testCases := []struct {
		name     string
		cards    []Card
		expected bool
	}{
		{
			name: "Valid 3-pair straight",
			cards: []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three),
				NewCard(Clubs, Four), NewCard(Diamonds, Four),
				NewCard(Hearts, Five), NewCard(Spades, Five),
			},
			expected: true,
		},
		{
			name: "Valid 4-pair straight",
			cards: []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three),
				NewCard(Clubs, Four), NewCard(Diamonds, Four),
				NewCard(Hearts, Five), NewCard(Spades, Five),
				NewCard(Clubs, Six), NewCard(Diamonds, Six),
			},
			expected: true,
		},
		{
			name: "Invalid pair straight with gap",
			cards: []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three),
				NewCard(Clubs, Five), NewCard(Diamonds, Five),
				NewCard(Hearts, Six), NewCard(Spades, Six),
			},
			expected: false,
		},
		{
			name: "Invalid pair straight with odd number of cards",
			cards: []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three),
				NewCard(Clubs, Four), NewCard(Diamonds, Four),
				NewCard(Hearts, Five),
			},
			expected: false,
		},
		{
			name: "Too short for pair straight",
			cards: []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three),
				NewCard(Clubs, Four), NewCard(Diamonds, Four),
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			group := NewCardGroup(tc.cards)
			result := (group.Category == PairStraight)
			
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for cards %v", tc.expected, result, tc.cards)
			}
		})
	}
}

func TestTripleStraightValidation(t *testing.T) {
	testCases := []struct {
		name     string
		cards    []Card
		expected bool
	}{
		{
			name: "Valid 2-triple straight",
			cards: []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three), NewCard(Clubs, Three),
				NewCard(Diamonds, Four), NewCard(Hearts, Four), NewCard(Spades, Four),
			},
			expected: true,
		},
		{
			name: "Invalid triple straight with gap",
			cards: []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three), NewCard(Clubs, Three),
				NewCard(Diamonds, Five), NewCard(Hearts, Five), NewCard(Spades, Five),
			},
			expected: false,
		},
		{
			name: "Invalid triple straight with wrong count",
			cards: []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three),
				NewCard(Diamonds, Four), NewCard(Hearts, Four), NewCard(Spades, Four),
			},
			expected: false,
		},
		{
			name: "Too short for triple straight",
			cards: []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three), NewCard(Clubs, Three),
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			group := NewCardGroup(tc.cards)
			result := (group.Category == TripleStraight)
			
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for cards %v", tc.expected, result, tc.cards)
			}
		})
	}
}

// Additional edge case tests for CardGroup

func TestCardGroupEdgeCases(t *testing.T) {
	testCases := []struct {
		name             string
		cards            []Card
		expectedCategory CardCategory
		expectedValid    bool
	}{
		{
			name:             "Single joker",
			cards:            []Card{NewJoker(SmallJoker)},
			expectedCategory: Single,
			expectedValid:    true,
		},
		{
			name:             "Big joker single",
			cards:            []Card{NewJoker(BigJoker)},
			expectedCategory: Single,
			expectedValid:    true,
		},
		{
			name:             "Mixed joker bomb",
			cards:            []Card{NewJoker(SmallJoker), NewJoker(BigJoker), NewJoker(SmallJoker)},
			expectedCategory: JokerBomb,
			expectedValid:    true,
		},
		{
			name:             "Four jokers bomb",
			cards:            []Card{NewJoker(SmallJoker), NewJoker(BigJoker), NewJoker(SmallJoker), NewJoker(BigJoker)},
			expectedCategory: JokerBomb,
			expectedValid:    true,
		},
		{
			name:             "Invalid mixed cards with jokers",
			cards:            []Card{NewJoker(SmallJoker), NewCard(Hearts, Ace)},
			expectedCategory: InvalidCategory,
			expectedValid:    false,
		},
		{
			name:             "Longest possible straight (A-K)",
			cards:            []Card{
				NewCard(Hearts, Ace), NewCard(Spades, Two), NewCard(Clubs, Three), 
				NewCard(Diamonds, Four), NewCard(Hearts, Five), NewCard(Spades, Six),
				NewCard(Clubs, Seven), NewCard(Diamonds, Eight), NewCard(Hearts, Nine),
				NewCard(Spades, Ten), NewCard(Clubs, Jack), NewCard(Diamonds, Queen),
				NewCard(Hearts, King),
			},
			expectedCategory: Straight,
			expectedValid:    true,
		},
		{
			name:             "Longest possible pair straight",
			cards:            []Card{
				NewCard(Hearts, Two), NewCard(Spades, Two),
				NewCard(Clubs, Three), NewCard(Diamonds, Three),
				NewCard(Hearts, Four), NewCard(Spades, Four),
				NewCard(Clubs, Five), NewCard(Diamonds, Five),
				NewCard(Hearts, Six), NewCard(Spades, Six),
				NewCard(Clubs, Seven), NewCard(Diamonds, Seven),
			},
			expectedCategory: PairStraight,
			expectedValid:    true,
		},
		{
			name:             "Unsorted cards valid straight",
			cards:            []Card{
				NewCard(Hearts, Seven), NewCard(Spades, Five), NewCard(Clubs, Six), 
				NewCard(Diamonds, Four), NewCard(Hearts, Three),
			},
			expectedCategory: Straight,
			expectedValid:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			group := NewCardGroup(tc.cards)
			
			if group.Category != tc.expectedCategory {
				t.Errorf("Expected category %v, got %v", tc.expectedCategory, group.Category)
			}
			
			if group.IsValid() != tc.expectedValid {
				t.Errorf("Expected valid %v, got %v", tc.expectedValid, group.IsValid())
			}
		})
	}
}

func TestCardGroupSize(t *testing.T) {
	testCases := []struct {
		name         string
		cards        []Card
		expectedSize int
	}{
		{"Empty", []Card{}, 0},
		{"Single", []Card{NewCard(Hearts, Ace)}, 1},
		{"Pair", []Card{NewCard(Hearts, Ace), NewCard(Spades, Ace)}, 2},
		{"Large group", []Card{
			NewCard(Hearts, Three), NewCard(Spades, Four), NewCard(Clubs, Five), 
			NewCard(Diamonds, Six), NewCard(Hearts, Seven), NewCard(Spades, Eight),
		}, 6},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			group := NewCardGroup(tc.cards)
			if group.Size != tc.expectedSize {
				t.Errorf("Expected size %d, got %d", tc.expectedSize, group.Size)
			}
		})
	}
}

func TestCardGroupConsistency(t *testing.T) {
	// Test that the same cards always produce the same analysis
	cards := []Card{
		NewCard(Hearts, Ace), NewCard(Spades, Ace), NewCard(Clubs, Ace),
	}

	group1 := NewCardGroup(cards)
	group2 := NewCardGroup(cards)

	if group1.Category != group2.Category {
		t.Error("Same cards should produce same category")
	}
	if group1.Size != group2.Size {
		t.Error("Same cards should produce same size")
	}
	if group1.Rank != group2.Rank {
		t.Error("Same cards should produce same rank")
	}
}