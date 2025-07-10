package domain

import (
	"fmt"
	"testing"
)

// 全面的牌组合法性测试，补充现有测试中的缺失场景

func TestAdvancedCardGroupValidity(t *testing.T) {
	testCases := []struct {
		name             string
		cards            []Card
		expectedCategory CardCategory
		expectedValid    bool
		description      string
	}{
		// 复杂的无效牌组
		{
			name:             "Invalid 3-card mixed group",
			cards:            []Card{NewCard(Hearts, Ace), NewCard(Spades, Ace), NewCard(Clubs, King)},
			expectedCategory: InvalidCategory,
			expectedValid:    false,
			description:      "两张相同+一张不同，不构成任何有效牌型",
		},
		{
			name:             "Invalid 4-card almost bomb",
			cards:            []Card{NewCard(Hearts, King), NewCard(Spades, King), NewCard(Clubs, King), NewCard(Diamonds, Queen)},
			expectedCategory: InvalidCategory,
			expectedValid:    false,
			description:      "三张相同+一张不同，不是炸弹",
		},
		{
			name:             "Invalid 5-card mixed group",
			cards:            []Card{NewCard(Hearts, Ace), NewCard(Spades, Ace), NewCard(Clubs, King), NewCard(Diamonds, King), NewCard(Hearts, Queen)},
			expectedCategory: InvalidCategory,
			expectedValid:    false,
			description:      "两对+单张，不是有效牌型",
		},
		
		// 边界长度测试
		{
			name:             "Minimum valid straight (5 cards)",
			cards:            []Card{NewCard(Hearts, Two), NewCard(Spades, Three), NewCard(Clubs, Four), NewCard(Diamonds, Five), NewCard(Hearts, Six)},
			expectedCategory: Straight,
			expectedValid:    true,
			description:      "最短有效顺子",
		},
		{
			name:             "Maximum realistic straight (12 cards, 2-K)",
			cards:            []Card{
				NewCard(Hearts, Two), NewCard(Spades, Three), NewCard(Clubs, Four), NewCard(Diamonds, Five),
				NewCard(Hearts, Six), NewCard(Spades, Seven), NewCard(Clubs, Eight), NewCard(Diamonds, Nine),
				NewCard(Hearts, Ten), NewCard(Spades, Jack), NewCard(Clubs, Queen), NewCard(Diamonds, King),
			},
			expectedCategory: Straight,
			expectedValid:    true,
			description:      "最长可能的顺子（不包含A）",
		},
		{
			name:             "Minimum valid pair straight (3 pairs)",
			cards:            []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three),
				NewCard(Clubs, Four), NewCard(Diamonds, Four),
				NewCard(Hearts, Five), NewCard(Spades, Five),
			},
			expectedCategory: PairStraight,
			expectedValid:    true,
			description:      "最短有效连对",
		},
		{
			name:             "Minimum valid triple straight (2 triples)",
			cards:            []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three), NewCard(Clubs, Three),
				NewCard(Diamonds, Four), NewCard(Hearts, Four), NewCard(Spades, Four),
			},
			expectedCategory: TripleStraight,
			expectedValid:    true,
			description:      "最短有效三顺",
		},
		
		// 特殊王牌组合
		{
			name:             "Single small joker",
			cards:            []Card{NewJoker(SmallJoker)},
			expectedCategory: Single,
			expectedValid:    true,
			description:      "单张小王",
		},
		{
			name:             "Single big joker",
			cards:            []Card{NewJoker(BigJoker)},
			expectedCategory: Single,
			expectedValid:    true,
			description:      "单张大王",
		},
		{
			name:             "Two same jokers",
			cards:            []Card{NewJoker(SmallJoker), NewJoker(SmallJoker)},
			expectedCategory: JokerBomb,
			expectedValid:    true,
			description:      "两张相同王牌",
		},
		{
			name:             "Mixed joker types",
			cards:            []Card{NewJoker(SmallJoker), NewJoker(BigJoker)},
			expectedCategory: JokerBomb,
			expectedValid:    true,
			description:      "大小王混合",
		},
		{
			name:             "Five jokers bomb",
			cards:            []Card{
				NewJoker(SmallJoker), NewJoker(BigJoker), NewJoker(SmallJoker), 
				NewJoker(BigJoker), NewJoker(SmallJoker),
			},
			expectedCategory: JokerBomb,
			expectedValid:    true,
			description:      "五张王炸",
		},
		{
			name:             "Joker mixed with normal card",
			cards:            []Card{NewJoker(SmallJoker), NewCard(Hearts, Ace)},
			expectedCategory: InvalidCategory,
			expectedValid:    false,
			description:      "王牌与普通牌混合",
		},
		{
			name:             "Joker mixed with normal pair",
			cards:            []Card{NewJoker(SmallJoker), NewCard(Hearts, Ace), NewCard(Spades, Ace)},
			expectedCategory: InvalidCategory,
			expectedValid:    false,
			description:      "王牌与对子混合",
		},
		
		// 特殊数值边界
		{
			name:             "Ace high straight (ending with Ace)",
			cards:            []Card{
				NewCard(Hearts, Ten), NewCard(Spades, Jack), NewCard(Clubs, Queen), 
				NewCard(Diamonds, King), NewCard(Hearts, Ace),
			},
			expectedCategory: Straight,
			expectedValid:    true,
			description:      "以A结尾的顺子",
		},
		{
			name:             "Two low straight (starting with 2)",
			cards:            []Card{
				NewCard(Hearts, Two), NewCard(Spades, Three), NewCard(Clubs, Four), 
				NewCard(Diamonds, Five), NewCard(Hearts, Six),
			},
			expectedCategory: Straight,
			expectedValid:    true,
			description:      "以2开始的顺子",
		},
		
		// 花色相关测试
		{
			name:             "Same suit straight",
			cards:            []Card{
				NewCard(Hearts, Three), NewCard(Hearts, Four), NewCard(Hearts, Five), 
				NewCard(Hearts, Six), NewCard(Hearts, Seven),
			},
			expectedCategory: Straight,
			expectedValid:    true,
			description:      "同花顺（在当前实现中识别为普通顺子）",
		},
		{
			name:             "All four suits bomb",
			cards:            []Card{
				NewCard(Hearts, King), NewCard(Diamonds, King), 
				NewCard(Clubs, King), NewCard(Spades, King),
			},
			expectedCategory: Bomb,
			expectedValid:    true,
			description:      "四花色炸弹",
		},
		{
			name:             "Same suit bomb (impossible with normal deck)",
			cards:            []Card{
				NewCard(Hearts, King), NewCard(Hearts, King), 
				NewCard(Hearts, King), NewCard(Hearts, King),
			},
			expectedCategory: Bomb,
			expectedValid:    true,
			description:      "同花色炸弹（理论上的测试）",
		},
		
		// 长度边界错误
		{
			name:             "4-card almost straight",
			cards:            []Card{NewCard(Hearts, Three), NewCard(Spades, Four), NewCard(Clubs, Five), NewCard(Diamonds, Six)},
			expectedCategory: InvalidCategory,
			expectedValid:    false,
			description:      "4张连续牌，不足以构成顺子",
		},
		{
			name:             "2-pair almost straight",
			cards:            []Card{NewCard(Hearts, Three), NewCard(Spades, Three), NewCard(Clubs, Four), NewCard(Diamonds, Four)},
			expectedCategory: InvalidCategory,
			expectedValid:    false,
			description:      "2对连续，不足以构成连对",
		},
		{
			name:             "1-triple cannot be triple straight",
			cards:            []Card{NewCard(Hearts, Three), NewCard(Spades, Three), NewCard(Clubs, Three)},
			expectedCategory: Triple,
			expectedValid:    true,
			description:      "单个三张，是三张而不是三顺",
		},
		
		// 间隔和重复错误
		{
			name:             "Straight with one gap",
			cards:            []Card{
				NewCard(Hearts, Three), NewCard(Spades, Four), NewCard(Clubs, Six), 
				NewCard(Diamonds, Seven), NewCard(Hearts, Eight),
			},
			expectedCategory: InvalidCategory,
			expectedValid:    false,
			description:      "顺子中有间隔",
		},
		{
			name:             "Pair straight with gap",
			cards:            []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three),
				NewCard(Clubs, Five), NewCard(Diamonds, Five),
				NewCard(Hearts, Six), NewCard(Spades, Six),
			},
			expectedCategory: InvalidCategory,
			expectedValid:    false,
			description:      "连对中有间隔",
		},
		{
			name:             "Triple straight with gap",
			cards:            []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three), NewCard(Clubs, Three),
				NewCard(Diamonds, Five), NewCard(Hearts, Five), NewCard(Spades, Five),
			},
			expectedCategory: InvalidCategory,
			expectedValid:    false,
			description:      "三顺中有间隔",
		},
		
		// 数量不匹配
		{
			name:             "Pair straight with extra card",
			cards:            []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three),
				NewCard(Clubs, Four), NewCard(Diamonds, Four),
				NewCard(Hearts, Five), NewCard(Spades, Five),
				NewCard(Clubs, Six), // 多余的牌
			},
			expectedCategory: InvalidCategory,
			expectedValid:    false,
			description:      "连对中有多余的牌",
		},
		{
			name:             "Triple straight with uneven triples",
			cards:            []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three), NewCard(Clubs, Three),
				NewCard(Diamonds, Four), NewCard(Hearts, Four), // 只有两张4
			},
			expectedCategory: InvalidCategory,
			expectedValid:    false,
			description:      "三顺中三张数量不均匀",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			group := NewCardGroup(tc.cards)
			
			if group.Category != tc.expectedCategory {
				t.Errorf("Expected category %v, got %v. %s", 
					tc.expectedCategory, group.Category, tc.description)
			}
			
			if group.IsValid() != tc.expectedValid {
				t.Errorf("Expected valid %v, got %v. %s", 
					tc.expectedValid, group.IsValid(), tc.description)
			}
		})
	}
}

func TestExtremeCaseValidations(t *testing.T) {
	testCases := []struct {
		name        string
		cards       []Card
		expectValid bool
		description string
	}{
		{
			name: "Maximum possible joker bomb",
			cards: []Card{
				NewJoker(SmallJoker), NewJoker(BigJoker), NewJoker(SmallJoker), 
				NewJoker(BigJoker), NewJoker(SmallJoker), NewJoker(BigJoker),
			},
			expectValid: true,
			description: "6张王炸（理论最大）",
		},
		{
			name: "Longest possible pair straight",
			cards: func() []Card {
				var cards []Card
				// 从2到A的连对 (12对，24张牌)
				for rank := Two; rank <= Ace; rank++ {
					cards = append(cards, NewCard(Hearts, rank), NewCard(Spades, rank))
				}
				return cards
			}(),
			expectValid: true,
			description: "最长可能的连对（2-A）",
		},
		{
			name: "Longest possible triple straight",
			cards: func() []Card {
				var cards []Card
				// 从2到A的三顺 (12组，36张牌)
				for rank := Two; rank <= Ace; rank++ {
					cards = append(cards, 
						NewCard(Hearts, rank), 
						NewCard(Spades, rank), 
						NewCard(Clubs, rank))
				}
				return cards
			}(),
			expectValid: true,
			description: "最长可能的三顺（2-A）",
		},
		{
			name:        "Empty card group",
			cards:       []Card{},
			expectValid: false,
			description: "空牌组",
		},
		{
			name: "Single card with duplicate (impossible in real game)",
			cards: []Card{
				NewCard(Hearts, Ace), NewCard(Hearts, Ace),
			},
			expectValid: true, // 应该识别为对子
			description: "理论上的重复单牌",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			group := NewCardGroup(tc.cards)
			
			if group.IsValid() != tc.expectValid {
				t.Errorf("Expected valid %v, got %v. %s", 
					tc.expectValid, group.IsValid(), tc.description)
			}
			
			t.Logf("Group: Category=%v, Size=%d, Valid=%v - %s", 
				group.Category, group.Size, group.IsValid(), tc.description)
		})
	}
}

func TestCardGroupAnalysisConsistency(t *testing.T) {
	// 测试分析逻辑的一致性
	testCards := [][]Card{
		{NewCard(Hearts, Ace)}, // 单张
		{NewCard(Hearts, Ace), NewCard(Spades, Ace)}, // 对子
		{NewCard(Hearts, King), NewCard(Spades, King), NewCard(Clubs, King)}, // 三张
		{NewCard(Hearts, King), NewCard(Spades, King), NewCard(Clubs, King), NewCard(Diamonds, King)}, // 炸弹
		{NewJoker(SmallJoker), NewJoker(BigJoker)}, // 王炸
		{NewCard(Hearts, Three), NewCard(Spades, Four), NewCard(Clubs, Five), NewCard(Diamonds, Six), NewCard(Hearts, Seven)}, // 顺子
	}

	for i, cards := range testCards {
		t.Run(fmt.Sprintf("Consistency_%d", i), func(t *testing.T) {
			// 多次创建相同的牌组，应该得到相同的结果
			groups := make([]*CardGroup, 5)
			for j := 0; j < 5; j++ {
				// 复制牌组以避免引用问题
				cardsCopy := make([]Card, len(cards))
				copy(cardsCopy, cards)
				groups[j] = NewCardGroup(cardsCopy)
			}

			// 验证所有分析结果相同
			for j := 1; j < 5; j++ {
				if groups[0].Category != groups[j].Category {
					t.Errorf("Inconsistent category: %v vs %v", groups[0].Category, groups[j].Category)
				}
				if groups[0].Size != groups[j].Size {
					t.Errorf("Inconsistent size: %d vs %d", groups[0].Size, groups[j].Size)
				}
				if groups[0].Rank != groups[j].Rank {
					t.Errorf("Inconsistent rank: %v vs %v", groups[0].Rank, groups[j].Rank)
				}
				if groups[0].IsValid() != groups[j].IsValid() {
					t.Errorf("Inconsistent validity: %v vs %v", groups[0].IsValid(), groups[j].IsValid())
				}
			}
		})
	}
}

func TestCardGroupValidityWithShuffledInput(t *testing.T) {
	// 测试输入顺序不影响合法性判断
	testCases := []struct {
		name           string
		orderedCards   []Card
		shuffledCards  []Card
		expectCategory CardCategory
	}{
		{
			name: "Shuffled straight",
			orderedCards: []Card{
				NewCard(Hearts, Three), NewCard(Spades, Four), NewCard(Clubs, Five), 
				NewCard(Diamonds, Six), NewCard(Hearts, Seven),
			},
			shuffledCards: []Card{
				NewCard(Hearts, Seven), NewCard(Spades, Four), NewCard(Clubs, Five), 
				NewCard(Diamonds, Six), NewCard(Hearts, Three),
			},
			expectCategory: Straight,
		},
		{
			name: "Shuffled pair straight",
			orderedCards: []Card{
				NewCard(Hearts, Three), NewCard(Spades, Three),
				NewCard(Clubs, Four), NewCard(Diamonds, Four),
				NewCard(Hearts, Five), NewCard(Spades, Five),
			},
			shuffledCards: []Card{
				NewCard(Hearts, Five), NewCard(Spades, Three),
				NewCard(Clubs, Four), NewCard(Spades, Five),
				NewCard(Hearts, Three), NewCard(Diamonds, Four),
			},
			expectCategory: PairStraight,
		},
		{
			name: "Shuffled bomb",
			orderedCards: []Card{
				NewCard(Hearts, King), NewCard(Spades, King), 
				NewCard(Clubs, King), NewCard(Diamonds, King),
			},
			shuffledCards: []Card{
				NewCard(Diamonds, King), NewCard(Hearts, King), 
				NewCard(Spades, King), NewCard(Clubs, King),
			},
			expectCategory: Bomb,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			orderedGroup := NewCardGroup(tc.orderedCards)
			shuffledGroup := NewCardGroup(tc.shuffledCards)

			if orderedGroup.Category != tc.expectCategory {
				t.Errorf("Ordered cards: expected category %v, got %v", 
					tc.expectCategory, orderedGroup.Category)
			}

			if shuffledGroup.Category != tc.expectCategory {
				t.Errorf("Shuffled cards: expected category %v, got %v", 
					tc.expectCategory, shuffledGroup.Category)
			}

			if orderedGroup.Category != shuffledGroup.Category {
				t.Errorf("Card order affects analysis: ordered=%v, shuffled=%v", 
					orderedGroup.Category, shuffledGroup.Category)
			}

			if orderedGroup.IsValid() != shuffledGroup.IsValid() {
				t.Errorf("Card order affects validity: ordered=%v, shuffled=%v", 
					orderedGroup.IsValid(), shuffledGroup.IsValid())
			}
		})
	}
}

// 性能测试：确保复杂牌组分析的性能
func BenchmarkComplexCardGroupAnalysis(b *testing.B) {
	// 创建一个复杂的长顺子
	longStraight := make([]Card, 12)
	for i, rank := range []Rank{Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King} {
		longStraight[i] = NewCard(Hearts, rank)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewCardGroup(longStraight)
		_ = group.IsValid()
	}
}

func BenchmarkJokerBombAnalysis(b *testing.B) {
	jokerBomb := []Card{
		NewJoker(SmallJoker), NewJoker(BigJoker), NewJoker(SmallJoker), 
		NewJoker(BigJoker), NewJoker(SmallJoker), NewJoker(BigJoker),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewCardGroup(jokerBomb)
		_ = group.IsValid()
	}
}