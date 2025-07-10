package domain

import (
	"testing"
)

// 测试tribute系统的各种场景
func TestTributeScenarios(t *testing.T) {
	testCases := []struct {
		name                string
		lastRankings        []SeatID // 上局排名 [1st, 2nd, 3rd, 4th]
		playerBigJokers     map[SeatID]int // 每个玩家的大王数量
		expectedScenario    TributeScenario
		expectedImmunity    bool
		expectedTributes    map[SeatID]SeatID // from -> to
		expectedReturnCount int
	}{
		{
			name:             "Double Down - with immunity",
			lastRankings:     []SeatID{SeatEast, SeatWest, SeatSouth, SeatNorth}, // East&West(1&2) vs South&North(3&4)
			playerBigJokers:  map[SeatID]int{SeatSouth: 1, SeatNorth: 1}, // 3+4合计2张大王
			expectedScenario: TributeScenarioDoubleDown,
			expectedImmunity: true,
			expectedTributes: map[SeatID]SeatID{}, // 免贡
			expectedReturnCount: 0,
		},
		{
			name:             "Double Down - without immunity", 
			lastRankings:     []SeatID{SeatEast, SeatWest, SeatSouth, SeatNorth}, // East&West(1&2) vs South&North(3&4)
			playerBigJokers:  map[SeatID]int{SeatSouth: 1, SeatNorth: 0}, // 3+4合计1张大王
			expectedScenario: TributeScenarioDoubleDown,
			expectedImmunity: false,
			expectedTributes: map[SeatID]SeatID{SeatSouth: SeatEast, SeatNorth: SeatWest}, // 3->1, 4->2
			expectedReturnCount: 2, // 1、2各还贡
		},
		{
			name:             "Single Last - with immunity",
			lastRankings:     []SeatID{SeatEast, SeatSouth, SeatWest, SeatNorth}, // East&West(1&3) vs South&North(2&4)
			playerBigJokers:  map[SeatID]int{SeatNorth: 2}, // 4单独握2张大王  
			expectedScenario: TributeScenarioSingleLast,
			expectedImmunity: true,
			expectedTributes: map[SeatID]SeatID{}, // 免贡
			expectedReturnCount: 0,
		},
		{
			name:             "Single Last - without immunity",
			lastRankings:     []SeatID{SeatEast, SeatSouth, SeatWest, SeatNorth}, // East&West(1&3) vs South&North(2&4)
			playerBigJokers:  map[SeatID]int{SeatNorth: 1}, // 4单独握1张大王
			expectedScenario: TributeScenarioSingleLast,
			expectedImmunity: false,
			expectedTributes: map[SeatID]SeatID{SeatNorth: SeatEast}, // 4->1
			expectedReturnCount: 1, // 1还贡
		},
		{
			name:             "Partner Last - with immunity",
			lastRankings:     []SeatID{SeatEast, SeatNorth, SeatSouth, SeatWest}, // 1&4 vs 2&3
			playerBigJokers:  map[SeatID]int{SeatSouth: 2}, // 3单独握2张大王
			expectedScenario: TributeScenarioPartnerLast,
			expectedImmunity: true,
			expectedTributes: map[SeatID]SeatID{}, // 免贡
			expectedReturnCount: 0,
		},
		{
			name:             "Partner Last - without immunity",
			lastRankings:     []SeatID{SeatEast, SeatNorth, SeatSouth, SeatWest},
			playerBigJokers:  map[SeatID]int{SeatSouth: 1}, // 3单独握1张大王
			expectedScenario: TributeScenarioPartnerLast,
			expectedImmunity: false,
			expectedTributes: map[SeatID]SeatID{SeatSouth: SeatEast}, // 3->1
			expectedReturnCount: 1, // 1还贡
		},
		{
			name:             "No tribute scenario - first deal",
			lastRankings:     nil, // 首局
			playerBigJokers:  map[SeatID]int{},
			expectedScenario: TributeScenarioNone,
			expectedImmunity: false,
			expectedTributes: map[SeatID]SeatID{}, // 无需贡牌
			expectedReturnCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 测试场景识别
			scenario := DetermineTributeScenario(tc.lastRankings)
			if scenario != tc.expectedScenario {
				t.Errorf("Expected scenario %v, got %v", tc.expectedScenario, scenario)
			}

			// 测试immunity检查
			immunity := CheckTributeImmunity(scenario, tc.playerBigJokers)
			if immunity != tc.expectedImmunity {
				t.Errorf("Expected immunity %v, got %v", tc.expectedImmunity, immunity)
			}

			// 测试tribute要求计算
			if !immunity {
				tributes := CalculateTributeRequirements(scenario)
				if len(tributes) != len(tc.expectedTributes) {
					t.Errorf("Expected %d tributes, got %d", len(tc.expectedTributes), len(tributes))
				}
				for from, to := range tc.expectedTributes {
					if tributes[from] != to {
						t.Errorf("Expected tribute %v->%v, got %v->%v", from, to, from, tributes[from])
					}
				}
			}
		})
	}
}

func TestTributeCardSelection(t *testing.T) {
	testCases := []struct {
		name              string
		hand              []Card
		trump             Rank
		expectedCard      Card
		expectedValid     bool
		description       string
	}{
		{
			name: "Select highest non-trump card",
			hand: []Card{
				NewCard(Hearts, Ace),   // 红桃A (trump)
				NewCard(Spades, King),  // 黑桃K
				NewCard(Clubs, Queen),  // 梅花Q
				NewCard(Diamonds, Jack), // 方片J
			},
			trump:         Ace,
			expectedCard:  NewCard(Spades, King), // 应该选择黑桃K（除了红桃A外最大）
			expectedValid: true,
			description:   "应该选择除红桃trump外最大的牌",
		},
		{
			name: "Select highest when no trump suit in hand",
			hand: []Card{
				NewCard(Spades, King),
				NewCard(Clubs, Queen),
				NewCard(Diamonds, Jack),
				NewCard(Hearts, Ten),
			},
			trump:         Ace, // A是主牌，但手中没有A
			expectedCard:  NewCard(Spades, King), // 应该选择最大的K
			expectedValid: true,
			description:   "手中没有trump时，选择最大的牌",
		},
		{
			name: "Select highest excluding Hearts trump",
			hand: []Card{
				NewCard(Hearts, Two),    // 红桃2 (trump)
				NewCard(Spades, Two),    // 黑桃2 (trump)
				NewCard(Clubs, Two),     // 梅花2 (trump)
				NewCard(Diamonds, Ace),  // 方片A
			},
			trump:         Two,
			expectedCard:  NewCard(Spades, Two), // 应该选择黑桃2（非红桃的trump牌优先级最高）
			expectedValid: true,
			description:   "除了红桃trump外，非红桃trump牌优先级最高",
		},
		{
			name: "Only have Hearts trump cards",
			hand: []Card{
				NewCard(Hearts, Two),   // 红桃2 (trump)
				NewCard(Hearts, Three), // 红桃3
			},
			trump:         Two,
			expectedCard:  NewCard(Hearts, Three), // 只能选择红桃3
			expectedValid: true,
			description:   "只有红桃trump时，选择非trump的红桃牌",
		},
		{
			name: "Invalid - only Hearts trump rank",
			hand: []Card{
				NewCard(Hearts, Two), // 红桃2 (trump)
			},
			trump:         Two,
			expectedCard:  NewCard(Hearts, Two), // 被迫选择红桃trump
			expectedValid: false, // 这种情况应该被标记为无效
			description:   "只有红桃trump时无法合法选择",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			selectedCard, valid := SelectTributeCard(tc.hand, tc.trump)
			
			if valid != tc.expectedValid {
				t.Errorf("Expected valid %v, got %v. %s", tc.expectedValid, valid, tc.description)
			}

			if valid && selectedCard != tc.expectedCard {
				t.Errorf("Expected card %v, got %v. %s", tc.expectedCard, selectedCard, tc.description)
			}
		})
	}
}

func TestReturnTributeCardSelection(t *testing.T) {
	testCases := []struct {
		name           string
		hand           []Card
		selectedCard   Card
		expectedValid  bool
		description    string
	}{
		{
			name: "Valid return tribute - rank 10",
			hand: []Card{
				NewCard(Hearts, Ten),
				NewCard(Spades, King),
				NewCard(Clubs, Ace),
			},
			selectedCard:  NewCard(Hearts, Ten),
			expectedValid: true,
			description:   "10点数的牌可以还贡",
		},
		{
			name: "Valid return tribute - rank 5",
			hand: []Card{
				NewCard(Hearts, Five),
				NewCard(Spades, King),
				NewCard(Clubs, Ace),
			},
			selectedCard:  NewCard(Hearts, Five),
			expectedValid: true,
			description:   "5点数的牌可以还贡",
		},
		{
			name: "Invalid return tribute - rank Jack",
			hand: []Card{
				NewCard(Hearts, Jack),
				NewCard(Spades, King),
				NewCard(Clubs, Ace),
			},
			selectedCard:  NewCard(Hearts, Jack),
			expectedValid: false,
			description:   "J(11点数)的牌不能还贡",
		},
		{
			name: "Valid return tribute - rank 2",
			hand: []Card{
				NewCard(Hearts, Two),
				NewCard(Spades, King),
				NewCard(Clubs, Ace),
			},
			selectedCard:  NewCard(Hearts, Two),
			expectedValid: true,
			description:   "2点数的牌可以还贡",
		},
		{
			name: "Invalid return tribute - not in hand",
			hand: []Card{
				NewCard(Hearts, Five),
				NewCard(Spades, King),
				NewCard(Clubs, Ace),
			},
			selectedCard:  NewCard(Hearts, Ten), // 手中没有这张牌
			expectedValid: false,
			description:   "不在手中的牌不能还贡",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid := IsValidReturnTributeCard(tc.hand, tc.selectedCard)
			
			if valid != tc.expectedValid {
				t.Errorf("Expected valid %v, got %v. %s", tc.expectedValid, valid, tc.description)
			}
		})
	}
}

func TestTributePhaseIntegration(t *testing.T) {
	// 集成测试：完整的tribute流程
	t.Run("Complete tribute flow", func(t *testing.T) {
		// 模拟Double Down场景，无immunity
		lastRankings := []SeatID{SeatEast, SeatWest, SeatSouth, SeatNorth} // East&West(1&2) vs South&North(3&4)
		playerBigJokers := map[SeatID]int{
			SeatSouth: 1, SeatNorth: 0, // 败方队伍合计1张大王，无immunity
		}

		// 1. 场景识别
		scenario := DetermineTributeScenario(lastRankings)
		if scenario != TributeScenarioDoubleDown {
			t.Errorf("Expected DoubleDown scenario, got %v", scenario)
		}

		// 2. Immunity检查
		immunity := CheckTributeImmunity(scenario, playerBigJokers)
		if immunity {
			t.Error("Should not have immunity with only 1 big joker")
		}

		// 3. 计算tribute要求
		tributes := CalculateTributeRequirements(scenario)
		expectedTributes := map[SeatID]SeatID{
			SeatSouth: SeatEast, // 3->1
			SeatNorth: SeatWest, // 4->2
		}
		
		for from, to := range expectedTributes {
			if tributes[from] != to {
				t.Errorf("Expected tribute %v->%v, got %v->%v", from, to, from, tributes[from])
			}
		}

		// 4. 验证手牌数量保持不变
		// 这部分需要在实际的状态机中测试
		t.Log("Tribute flow validation completed")
	})
}