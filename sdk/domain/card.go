package domain

import (
	"strconv"
)

type Suit int

const (
	Hearts Suit = iota
	Diamonds
	Clubs
	Spades
	Joker
)

func (s Suit) String() string {
	switch s {
	case Hearts:
		return "♥"
	case Diamonds:
		return "♦"
	case Clubs:
		return "♣"
	case Spades:
		return "♠"
	case Joker:
		return "🃏"
	default:
		return "?"
	}
}

type Rank int

const (
	Two Rank = iota + 1
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
	SmallJoker
	BigJoker
)

func (r Rank) String() string {
	switch r {
	case Ace:
		return "A"
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	case SmallJoker:
		return "小王"
	case BigJoker:
		return "大王"
	default:
		return strconv.Itoa(int(r))
	}
}

type Card struct {
	Suit Suit
	Rank Rank
}

func (c Card) String() string {
	if c.Suit == Joker {
		return c.Rank.String()
	}
	return c.Suit.String() + c.Rank.String()
}

func NewCard(suit Suit, rank Rank) Card {
	return Card{Suit: suit, Rank: rank}
}

func NewJoker(rank Rank) Card {
	return Card{Suit: Joker, Rank: rank}
}

func (c Card) IsJoker() bool {
	return c.Suit == Joker
}

func (c Card) IsRedSuit() bool {
	return c.Suit == Hearts || c.Suit == Diamonds
}

func (c Card) IsBlackSuit() bool {
	return c.Suit == Spades || c.Suit == Clubs
}

type CardID int

func (c Card) ID() CardID {
	return CardID(int(c.Suit)*15 + int(c.Rank))
}

func CardFromID(id CardID) Card {
	suitVal := int(id) / 15
	rankVal := int(id) % 15
	return Card{Suit: Suit(suitVal), Rank: Rank(rankVal)}
}