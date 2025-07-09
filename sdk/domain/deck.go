package domain

import (
	"math/rand"
	"time"
)

type Deck struct {
	Cards []Card
	rng   *rand.Rand
}

func NewDeck() *Deck {
	deck := &Deck{
		Cards: make([]Card, 0, 108),
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	deck.initialize()
	return deck
}

func NewDeckWithSeed(seed int64) *Deck {
	deck := &Deck{
		Cards: make([]Card, 0, 108),
		rng:   rand.New(rand.NewSource(seed)),
	}
	deck.initialize()
	return deck
}

func (d *Deck) initialize() {
	suits := []Suit{Hearts, Diamonds, Clubs, Spades}
	ranks := []Rank{Ace, Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King}
	
	for i := 0; i < 2; i++ {
		for _, suit := range suits {
			for _, rank := range ranks {
				d.Cards = append(d.Cards, NewCard(suit, rank))
			}
		}
		
		d.Cards = append(d.Cards, NewJoker(SmallJoker))
		d.Cards = append(d.Cards, NewJoker(BigJoker))
	}
}

func (d *Deck) Shuffle() {
	d.rng.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

func (d *Deck) ShuffleWithSeed(seed int64) {
	d.rng = rand.New(rand.NewSource(seed))
	d.Shuffle()
}

func (d *Deck) Deal(numCards int) []Card {
	if numCards > len(d.Cards) {
		numCards = len(d.Cards)
	}
	
	dealt := make([]Card, numCards)
	copy(dealt, d.Cards[:numCards])
	d.Cards = d.Cards[numCards:]
	
	return dealt
}

func (d *Deck) DealToHands(numPlayers int) [][]Card {
	if numPlayers <= 0 || numPlayers > 4 {
		return nil
	}
	
	cardsPerPlayer := len(d.Cards) / numPlayers
	hands := make([][]Card, numPlayers)
	
	for i := 0; i < numPlayers; i++ {
		hands[i] = d.Deal(cardsPerPlayer)
	}
	
	return hands
}

func (d *Deck) Remaining() int {
	return len(d.Cards)
}

func (d *Deck) IsEmpty() bool {
	return len(d.Cards) == 0
}

func (d *Deck) Reset() {
	d.Cards = d.Cards[:0]
	d.initialize()
}

func (d *Deck) Size() int {
	return len(d.Cards)
}

func (d *Deck) GetCard(index int) Card {
	if index < 0 || index >= len(d.Cards) {
		return Card{}
	}
	return d.Cards[index]
}