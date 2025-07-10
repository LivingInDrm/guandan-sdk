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
	return CmpEqual
}

func getCardValue(card Card, trump Rank) int {
	if card.IsJoker() {
		if card.Rank == SmallJoker {
			return 1000
		}
		if card.Rank == BigJoker {
			return 1001
		}
	}
	
	if card.Rank == trump {
		return 500 + int(card.Suit)
	}
	
	return int(card.Rank)
}

func CompareCardGroups(a, b *CardGroup, trump Rank) CmpResult {
	if !a.IsValid() || !b.IsValid() {
		return CmpEqual
	}
	
	aKey := a.ComparisonKey()
	bKey := b.ComparisonKey()
	
	if aKey.Category == JokerBomb && bKey.Category != JokerBomb {
		return CmpGreater
	}
	if aKey.Category != JokerBomb && bKey.Category == JokerBomb {
		return CmpLess
	}
	
	if aKey.Category == Bomb && bKey.Category != Bomb && bKey.Category != JokerBomb {
		return CmpGreater
	}
	if aKey.Category != Bomb && aKey.Category != JokerBomb && bKey.Category == Bomb {
		return CmpLess
	}
	
	if aKey.Category == JokerBomb && bKey.Category == JokerBomb {
		if aKey.Size > bKey.Size {
			return CmpGreater
		} else if aKey.Size < bKey.Size {
			return CmpLess
		}
		return CmpEqual
	}
	
	if aKey.Category == Bomb && bKey.Category == Bomb {
		aValue := getBombValue(a, trump)
		bValue := getBombValue(b, trump)
		
		if aValue > bValue {
			return CmpGreater
		} else if aValue < bValue {
			return CmpLess
		}
		return CmpEqual
	}
	
	if aKey.Category != bKey.Category {
		return CmpEqual
	}
	
	if aKey.Size != bKey.Size {
		return CmpEqual
	}
	
	aValue := getGroupValue(a, trump)
	bValue := getGroupValue(b, trump)
	
	if aValue > bValue {
		return CmpGreater
	} else if aValue < bValue {
		return CmpLess
	}
	return CmpEqual
}

func getBombValue(group *CardGroup, trump Rank) int {
	if group.Category != Bomb {
		return 0
	}
	
	rank := group.Rank
	if rank == trump {
		return 1000 + int(rank)
	}
	
	return int(rank)
}

func getGroupValue(group *CardGroup, trump Rank) int {
	if group.Category == JokerBomb {
		return 2000 + group.Size
	}
	
	if group.Category == Bomb {
		return getBombValue(group, trump)
	}
	
	// 对于非炸弹类型，需要考虑王牌的特殊值
	rank := group.Rank
	
	// 处理王牌（小王、大王）
	if rank == SmallJoker {
		return 1000
	}
	if rank == BigJoker {
		return 1001
	}
	
	// 处理主牌
	if rank == trump {
		return 500 + int(rank)
	}
	
	return int(rank)
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