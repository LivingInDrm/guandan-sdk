package domain

import (
	"fmt"
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

// ParseCard parses a card from its string representation
func ParseCard(cardStr string) (Card, error) {
	if len(cardStr) == 0 {
		return Card{}, fmt.Errorf("empty card string")
	}
	
	// Handle jokers
	if cardStr == "小王" || cardStr == "SJ" {
		return NewJoker(SmallJoker), nil
	}
	if cardStr == "大王" || cardStr == "BJ" {
		return NewJoker(BigJoker), nil
	}
	
	// Parse suit and rank
	if len(cardStr) < 2 {
		return Card{}, fmt.Errorf("invalid card string: %s", cardStr)
	}
	
	suitStr := cardStr[:1]
	rankStr := cardStr[1:]
	
	// Parse suit
	var suit Suit
	switch suitStr {
	case "♥", "H":
		suit = Hearts
	case "♦", "D":
		suit = Diamonds
	case "♣", "C":
		suit = Clubs
	case "♠", "S":
		suit = Spades
	default:
		return Card{}, fmt.Errorf("invalid suit: %s", suitStr)
	}
	
	// Parse rank
	var rank Rank
	switch rankStr {
	case "A":
		rank = Ace
	case "2":
		rank = Two
	case "3":
		rank = Three
	case "4":
		rank = Four
	case "5":
		rank = Five
	case "6":
		rank = Six
	case "7":
		rank = Seven
	case "8":
		rank = Eight
	case "9":
		rank = Nine
	case "10", "T":
		rank = Ten
	case "J":
		rank = Jack
	case "Q":
		rank = Queen
	case "K":
		rank = King
	default:
		return Card{}, fmt.Errorf("invalid rank: %s", rankStr)
	}
	
	return NewCard(suit, rank), nil
}