package domain

import "fmt"

type SeatID int

const (
	SeatEast SeatID = iota
	SeatSouth
	SeatWest
	SeatNorth
)

func (s SeatID) String() string {
	switch s {
	case SeatEast:
		return "East"
	case SeatSouth:
		return "South"
	case SeatWest:
		return "West"
	case SeatNorth:
		return "North"
	default:
		return "Unknown"
	}
}

func (s SeatID) Next() SeatID {
	return SeatID((int(s) + 1) % 4)
}

func (s SeatID) Previous() SeatID {
	return SeatID((int(s) + 3) % 4)
}

func (s SeatID) Opposite() SeatID {
	return SeatID((int(s) + 2) % 4)
}

func (s SeatID) IsValid() bool {
	return s >= SeatEast && s <= SeatNorth
}

type TeamID int

const (
	TeamEastWest TeamID = iota
	TeamSouthNorth
)

func (t TeamID) String() string {
	switch t {
	case TeamEastWest:
		return "East-West"
	case TeamSouthNorth:
		return "South-North"
	default:
		return "Unknown"
	}
}

func (t TeamID) OpposingTeam() TeamID {
	if t == TeamEastWest {
		return TeamSouthNorth
	}
	return TeamEastWest
}

func GetTeamFromSeat(seat SeatID) TeamID {
	if seat == SeatEast || seat == SeatWest {
		return TeamEastWest
	}
	return TeamSouthNorth
}

type Player struct {
	ID       string
	Name     string
	SeatID   SeatID
	TeamID   TeamID
	Level    Rank
	Hand     []Card
	IsOnline bool
}

func NewPlayer(id, name string, seat SeatID) *Player {
	return &Player{
		ID:       id,
		Name:     name,
		SeatID:   seat,
		TeamID:   GetTeamFromSeat(seat),
		Level:    Two,
		Hand:     make([]Card, 0),
		IsOnline: true,
	}
}

func (p *Player) String() string {
	return fmt.Sprintf("%s(%s)", p.Name, p.SeatID.String())
}

func (p *Player) GetPartnerSeat() SeatID {
	return p.SeatID.Opposite()
}

func (p *Player) AddCards(cards []Card) {
	p.Hand = append(p.Hand, cards...)
}

func (p *Player) RemoveCards(cards []Card) bool {
	for _, cardToRemove := range cards {
		found := false
		for i, handCard := range p.Hand {
			if handCard.Suit == cardToRemove.Suit && handCard.Rank == cardToRemove.Rank {
				p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (p *Player) HasCard(card Card) bool {
	for _, handCard := range p.Hand {
		if handCard.Suit == card.Suit && handCard.Rank == card.Rank {
			return true
		}
	}
	return false
}

func (p *Player) HasCards(cards []Card) bool {
	for _, card := range cards {
		if !p.HasCard(card) {
			return false
		}
	}
	return true
}

func (p *Player) HandSize() int {
	return len(p.Hand)
}

func (p *Player) IsHandEmpty() bool {
	return len(p.Hand) == 0
}

func (p *Player) ClearHand() {
	p.Hand = p.Hand[:0]
}

func (p *Player) GetHand() []Card {
	hand := make([]Card, len(p.Hand))
	copy(hand, p.Hand)
	return hand
}

type Team struct {
	ID      TeamID
	Players [2]*Player
	Level   Rank
}

func NewTeam(id TeamID) *Team {
	return &Team{
		ID:      id,
		Players: [2]*Player{},
		Level:   Two,
	}
}

func (t *Team) AddPlayer(player *Player) bool {
	if t.Players[0] == nil {
		t.Players[0] = player
		return true
	} else if t.Players[1] == nil {
		t.Players[1] = player
		return true
	}
	return false
}

func (t *Team) GetPlayers() []*Player {
	players := make([]*Player, 0, 2)
	for _, player := range t.Players {
		if player != nil {
			players = append(players, player)
		}
	}
	return players
}

func (t *Team) GetPlayerBySeat(seat SeatID) *Player {
	for _, player := range t.Players {
		if player != nil && player.SeatID == seat {
			return player
		}
	}
	return nil
}

func (t *Team) String() string {
	return t.ID.String()
}

type PlayerArray [4]*Player

func NewPlayerArray() PlayerArray {
	return PlayerArray{}
}

func (pa *PlayerArray) Set(seat SeatID, player *Player) {
	if seat.IsValid() {
		pa[seat] = player
	}
}

func (pa *PlayerArray) Get(seat SeatID) *Player {
	if seat.IsValid() {
		return pa[seat]
	}
	return nil
}

func (pa *PlayerArray) All() []*Player {
	players := make([]*Player, 0, 4)
	for _, player := range pa {
		if player != nil {
			players = append(players, player)
		}
	}
	return players
}

func (pa *PlayerArray) IsComplete() bool {
	for _, player := range pa {
		if player == nil {
			return false
		}
	}
	return true
}