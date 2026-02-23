package model

import (
	"fmt"
	"strings"
)

const CardsPerPlayer = 10

type Player struct {
	TeamId      int    `json:"TeamId"`
	Name        string `json:"Name"`
	Hand        Hand   `json:"Hand"`
	MatchPoints int    `json:"MatchPoints"`
	TotalPoints int    `json:"TotalPoints"`
}

type PlayedCard struct {
	PlayerName string `json:"PlayerName"`
	Card       Card   `json:"Card"`
}

type Match struct {
	Players     []Player     `json:"players"`
	Table       []PlayedCard `json:"table"`
	FirstPlayer string       `json:"firstPlayer"`
	TrumpSuit   string       `json:"trumpSuit"`
}

// TODO: check
func (match Match) String() string {
	var sb strings.Builder
	for i, player := range match.Players {
		sb.WriteString("Player " + string(rune(i+1)) + ": " + player.String() + "\n")
	}
	sb.WriteString("Table: " + fmt.Sprintf("%+v", match.Table) + "\n")
	sb.WriteString("Trump Suit: " + match.TrumpSuit + "\n")
	return sb.String()
}

func (player Player) String() string {
	return player.Name + " (Team " + string(rune(player.TeamId+'0')) + ") - Hand: " + player.Hand.String()
}
