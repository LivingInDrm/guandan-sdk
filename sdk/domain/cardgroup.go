package domain

import (
	"sort"
)

type CardCategory int

const (
	InvalidCategory CardCategory = iota
	Single
	Pair
	Triple
	Straight
	PairStraight
	TripleStraight
	Bomb
	JokerBomb
)

func (c CardCategory) String() string {
	switch c {
	case Single:
		return "Single"
	case Pair:
		return "Pair"
	case Triple:
		return "Triple"
	case Straight:
		return "Straight"
	case PairStraight:
		return "PairStraight"
	case TripleStraight:
		return "TripleStraight"
	case Bomb:
		return "Bomb"
	case JokerBomb:
		return "JokerBomb"
	default:
		return "Invalid"
	}
}

type CardGroup struct {
	Cards []Card
	Category CardCategory
	Size int
	Rank Rank
}

func NewCardGroup(cards []Card) *CardGroup {
	if len(cards) == 0 {
		return &CardGroup{
			Cards: cards,
			Category: InvalidCategory,
			Size: 0,
			Rank: 0,
		}
	}
	
	cg := &CardGroup{
		Cards: make([]Card, len(cards)),
	}
	copy(cg.Cards, cards)
	
	cg.analyze()
	return cg
}

func (cg *CardGroup) analyze() {
	if len(cg.Cards) == 0 {
		cg.Category = InvalidCategory
		return
	}
	
	cg.Size = len(cg.Cards)
	
	switch cg.Size {
	case 1:
		cg.analyzeSingle()
	case 2:
		if cg.isJokerBomb() {
			cg.Category = JokerBomb
		} else {
			cg.analyzePair()
		}
	case 3:
		if cg.isJokerBomb() {
			cg.Category = JokerBomb
		} else {
			cg.analyzeTriple()
		}
	case 4:
		if cg.isJokerBomb() {
			cg.Category = JokerBomb
		} else {
			cg.analyzeFour()
		}
	default:
		cg.analyzeLonger()
	}
}

func (cg *CardGroup) analyzeSingle() {
	cg.Category = Single
	cg.Rank = cg.Cards[0].Rank
}

func (cg *CardGroup) analyzePair() {
	ranks := cg.getRankCounts()
	
	if len(ranks) == 1 {
		for rank := range ranks {
			if ranks[rank] == 2 {
				cg.Category = Pair
				cg.Rank = rank
				return
			}
		}
	}
	
	cg.Category = InvalidCategory
}

func (cg *CardGroup) analyzeTriple() {
	ranks := cg.getRankCounts()
	
	if len(ranks) == 1 {
		for rank := range ranks {
			if ranks[rank] == 3 {
				cg.Category = Triple
				cg.Rank = rank
				return
			}
		}
	}
	
	cg.Category = InvalidCategory
}

func (cg *CardGroup) analyzeFour() {
	ranks := cg.getRankCounts()
	
	if len(ranks) == 1 {
		for rank := range ranks {
			if ranks[rank] == 4 {
				cg.Category = Bomb
				cg.Rank = rank
				return
			}
		}
	}
	
	cg.Category = InvalidCategory
}

func (cg *CardGroup) analyzeLonger() {
	if cg.isJokerBomb() {
		cg.Category = JokerBomb
		return
	}
	
	if cg.isStraight() {
		cg.Category = Straight
		return
	}
	
	if cg.isPairStraight() {
		cg.Category = PairStraight
		return
	}
	
	if cg.isTripleStraight() {
		cg.Category = TripleStraight
		return
	}
	
	cg.Category = InvalidCategory
}

func (cg *CardGroup) isJokerBomb() bool {
	jokerCount := 0
	for _, card := range cg.Cards {
		if card.IsJoker() {
			jokerCount++
		}
	}
	return jokerCount >= 2 && jokerCount == len(cg.Cards)
}

func (cg *CardGroup) isStraight() bool {
	if len(cg.Cards) < 5 {
		return false
	}
	
	ranks := cg.getSortedRanks()
	if len(ranks) != len(cg.Cards) {
		return false
	}
	
	for i := 1; i < len(ranks); i++ {
		if ranks[i] != ranks[i-1]+1 {
			return false
		}
	}
	
	cg.Rank = ranks[0]
	return true
}

func (cg *CardGroup) isPairStraight() bool {
	if len(cg.Cards) < 6 || len(cg.Cards)%2 != 0 {
		return false
	}
	
	ranks := cg.getRankCounts()
	sortedRanks := make([]Rank, 0, len(ranks))
	
	for rank, count := range ranks {
		if count != 2 {
			return false
		}
		sortedRanks = append(sortedRanks, rank)
	}
	
	sort.Slice(sortedRanks, func(i, j int) bool {
		return sortedRanks[i] < sortedRanks[j]
	})
	
	for i := 1; i < len(sortedRanks); i++ {
		if sortedRanks[i] != sortedRanks[i-1]+1 {
			return false
		}
	}
	
	cg.Rank = sortedRanks[0]
	return true
}

func (cg *CardGroup) isTripleStraight() bool {
	if len(cg.Cards) < 6 || len(cg.Cards)%3 != 0 {
		return false
	}
	
	ranks := cg.getRankCounts()
	sortedRanks := make([]Rank, 0, len(ranks))
	
	for rank, count := range ranks {
		if count != 3 {
			return false
		}
		sortedRanks = append(sortedRanks, rank)
	}
	
	sort.Slice(sortedRanks, func(i, j int) bool {
		return sortedRanks[i] < sortedRanks[j]
	})
	
	for i := 1; i < len(sortedRanks); i++ {
		if sortedRanks[i] != sortedRanks[i-1]+1 {
			return false
		}
	}
	
	cg.Rank = sortedRanks[0]
	return true
}

func (cg *CardGroup) getRankCounts() map[Rank]int {
	counts := make(map[Rank]int)
	for _, card := range cg.Cards {
		counts[card.Rank]++
	}
	return counts
}

func (cg *CardGroup) getSortedRanks() []Rank {
	ranks := make([]Rank, len(cg.Cards))
	for i, card := range cg.Cards {
		ranks[i] = card.Rank
	}
	
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i] < ranks[j]
	})
	
	return ranks
}

func (cg *CardGroup) IsValid() bool {
	return cg.Category != InvalidCategory
}

func (cg *CardGroup) IsBomb() bool {
	return cg.Category == Bomb || cg.Category == JokerBomb
}

func (cg *CardGroup) String() string {
	return cg.Category.String()
}

type ComparisonKey struct {
	Category CardCategory
	Size     int
	Rank     Rank
}

func (cg *CardGroup) ComparisonKey() ComparisonKey {
	return ComparisonKey{
		Category: cg.Category,
		Size:     cg.Size,
		Rank:     cg.Rank,
	}
}