package domain

type CmpResult int

const (
	CmpEqual CmpResult = iota
	CmpLess
	CmpGreater
)

func (r CmpResult) String() string {
	switch r {
	case CmpEqual:
		return "Equal"
	case CmpLess:
		return "Less"
	case CmpGreater:
		return "Greater"
	default:
		return "Unknown"
	}
}

func CompareCards(a, b Card, trump Rank) CmpResult {
	aValue := getCardValue(a, trump)
	bValue := getCardValue(b, trump)
	
	if aValue < bValue {
		return CmpLess
	} else if aValue > bValue {
		return CmpGreater
	}
	
	// 花色不影响牌的大小比较，相同rank值的牌被认为是相等的
	return CmpEqual
}

func getCardValue(card Card, trump Rank) int {
	if card.IsJoker() {
		if card.Rank == SmallJoker {
			return 14
		}
		if card.Rank == BigJoker {
			return 15
		}
	}
	
	if card.Rank == trump {
		return 13 // 主牌统一返回13，花色比较在CompareCards中特殊处理
	}
	
	return int(card.Rank)
}

func CompareCardGroups(a, b *CardGroup, trump Rank) CmpResult {
	if !a.IsValid() || !b.IsValid() {
		return CmpEqual
	}
	
	aKey := a.ComparisonKey()
	bKey := b.ComparisonKey()
	
	// Step 1: Compare CAT values (0-5) - higher CAT always wins
	if aKey.CAT > bKey.CAT {
		return CmpGreater
	} else if aKey.CAT < bKey.CAT {
		return CmpLess
	}
	
	// Step 2: Same CAT - check if they can be compared
	// Special case: bombs of same CAT can be compared by size
	if aKey.CAT == bKey.CAT && (aKey.CAT == 5 || aKey.CAT == 4 || aKey.CAT == 2 || aKey.CAT == 1) {
		// For bombs (including joker bombs), compare by size first
		if aKey.Size > bKey.Size {
			return CmpGreater
		} else if aKey.Size < bKey.Size {
			return CmpLess
		}
		// Same size bombs, compare by rank
	} else {
		// For non-bomb types, they must have same Category and Size to be comparable
		if aKey.Category != bKey.Category || aKey.Size != bKey.Size {
			return CmpEqual // Cannot compare different card types
		}
	}
	
	// Step 3: Same CAT, Category and SIZE, compare RANK
	aRankValue := getGroupRankValue(a, trump)
	bRankValue := getGroupRankValue(b, trump)
	
	if aRankValue > bRankValue {
		return CmpGreater
	} else if aRankValue < bRankValue {
		return CmpLess
	}
	
	return CmpEqual
}


// getGroupRankValue returns the RANK value for group comparison according to Guandan rules
func getGroupRankValue(group *CardGroup, trump Rank) int {
	switch group.Category {
	case JokerBomb:
		return 0 // 王炸之间无点数比较，固定返回0
	
	case Bomb:
		// 炸弹点值：使用该炸弹的点值，主牌炸弹优先于同点数普通炸弹
		rank := group.Rank
		if rank == trump {
			return 13 // 主牌炸弹使用值13
		}
		return int(rank) // 普通炸弹使用原始rank值(0-12)
	
	case Straight:
		// 顺子最大牌点值，包括同花顺
		return getMaxCardValueInGroup(group, trump)
	
	case PairStraight:
		// 三连对：最高对子点值
		return getMaxCardValueInGroup(group, trump)
	
	case TripleStraight:
		// 钢板（二连三）：最高三同张点值
		return getMaxCardValueInGroup(group, trump)
	
	case Triple:
		// 三带二或三同张：三同张点值（忽略对子）
		if group.Size == 5 {
			// 三带二：找到三同张的点值
			return getTripleValueInGroup(group, trump)
		}
		// 三同张
		return getCardValue(Card{Rank: group.Rank}, trump)
	
	default:
		// 其他牌型（单张、对子）：该点值
		return getCardValue(Card{Rank: group.Rank}, trump)
	}
}

// getMaxCardValueInGroup returns the maximum card value in the group
func getMaxCardValueInGroup(group *CardGroup, trump Rank) int {
	maxValue := -1
	for _, card := range group.Cards {
		value := getCardValue(card, trump)
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}

// getTripleValueInGroup finds the rank that appears 3 times in a 三带二 group
func getTripleValueInGroup(group *CardGroup, trump Rank) int {
	rankCounts := make(map[Rank]int)
	for _, card := range group.Cards {
		rankCounts[card.Rank]++
	}
	
	for rank, count := range rankCounts {
		if count == 3 {
			return getCardValue(Card{Rank: rank}, trump)
		}
	}
	
	// Fallback: should not happen in valid 三带二
	return getCardValue(Card{Rank: group.Rank}, trump)
}


func CanBeat(hand, tablePlay *CardGroup, trump Rank) bool {
	if hand == nil || !hand.IsValid() {
		return false
	}
	
	if tablePlay == nil || !tablePlay.IsValid() {
		return true
	}
	
	result := CompareCardGroups(hand, tablePlay, trump)
	return result == CmpGreater
}

func CanFollow(hand, tablePlay *CardGroup, trump Rank) bool {
	if hand == nil || !hand.IsValid() {
		return false
	}
	
	if tablePlay == nil || !tablePlay.IsValid() {
		return true
	}
	
	if hand.IsBomb() && !tablePlay.IsBomb() {
		return true
	}
	
	if hand.Category != tablePlay.Category {
		return false
	}
	
	if hand.Size != tablePlay.Size {
		return false
	}
	
	return CanBeat(hand, tablePlay, trump)
}

func GetPlayableCards(hand []Card, tablePlay *CardGroup, trump Rank) [][]Card {
	if len(hand) == 0 {
		return nil
	}
	
	var playable [][]Card
	
	for i := 1; i <= len(hand); i++ {
		combinations := generateCombinations(hand, i)
		for _, combo := range combinations {
			group := NewCardGroup(combo)
			if group.IsValid() && CanFollow(group, tablePlay, trump) {
				playable = append(playable, combo)
			}
		}
	}
	
	return playable
}

func generateCombinations(cards []Card, size int) [][]Card {
	if size == 0 {
		return [][]Card{{}}
	}
	
	if len(cards) == 0 {
		return nil
	}
	
	var result [][]Card
	
	first := cards[0]
	rest := cards[1:]
	
	combosWithFirst := generateCombinations(rest, size-1)
	for _, combo := range combosWithFirst {
		newCombo := append([]Card{first}, combo...)
		result = append(result, newCombo)
	}
	
	combosWithoutFirst := generateCombinations(rest, size)
	result = append(result, combosWithoutFirst...)
	
	return result
}

func IsTrump(card Card, trump Rank) bool {
	if card.IsJoker() {
		return true
	}
	return card.Rank == trump
}

func IsHigherTrump(a, b Card, trump Rank) bool {
	if !IsTrump(a, trump) || !IsTrump(b, trump) {
		return false
	}
	
	return CompareCards(a, b, trump) == CmpGreater
}

func GetTrumpCards(cards []Card, trump Rank) []Card {
	var trumpCards []Card
	for _, card := range cards {
		if IsTrump(card, trump) {
			trumpCards = append(trumpCards, card)
		}
	}
	return trumpCards
}

func GetNonTrumpCards(cards []Card, trump Rank) []Card {
	var nonTrumpCards []Card
	for _, card := range cards {
		if !IsTrump(card, trump) {
			nonTrumpCards = append(nonTrumpCards, card)
		}
	}
	return nonTrumpCards
}

func CountTrumps(cards []Card, trump Rank) int {
	count := 0
	for _, card := range cards {
		if IsTrump(card, trump) {
			count++
		}
	}
	return count
}

func HasBomb(cards []Card, trump Rank) bool {
	for i := 1; i <= len(cards); i++ {
		combinations := generateCombinations(cards, i)
		for _, combo := range combinations {
			group := NewCardGroup(combo)
			if group.IsBomb() {
				return true
			}
		}
	}
	return false
}

func FindBombs(cards []Card, trump Rank) []*CardGroup {
	var bombs []*CardGroup
	
	for i := 1; i <= len(cards); i++ {
		combinations := generateCombinations(cards, i)
		for _, combo := range combinations {
			group := NewCardGroup(combo)
			if group.IsBomb() {
				bombs = append(bombs, group)
			}
		}
	}
	
	return bombs
}