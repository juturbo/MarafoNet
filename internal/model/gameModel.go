package model

import (
	"fmt"
	"strings"
)

const (
	CardsPerPlayer = 10
	NumberOfTeams  = 2
	PointsToWin    = 41
)

type Point uint8

type Player struct {
	TeamId int    `json:"TeamId"`
	Name   string `json:"Name"`
	Hand   Hand   `json:"Hand"`
}

type PlayedCard struct {
	PlayerName string `json:"PlayerName"`
	Card       Card   `json:"Card"`
}

type Game struct {
	Players       []Player             `json:"Players"`
	Table         []PlayedCard         `json:"Table"`
	MatchPoints   [NumberOfTeams]Point `json:"MatchPoints"`
	TotalPoints   [NumberOfTeams]int   `json:"TotalPoints"`
	FirstPlayer   string               `json:"FirstPlayer"`
	TrumpSuit     Suit                 `json:"TrumpSuit"`
	WinnerTeam    *int                 `json:"WinnerTeam,omitempty"`
	WinnerPlayers []string             `json:"WinnerPlayers,omitempty"`
}

// TODO: check
func (match Game) String() string {
	var sb strings.Builder
	for i, player := range match.Players {
		sb.WriteString("Player " + string(rune(i+1)) + ": " + player.String() + "\n")
	}
	sb.WriteString("Table: " + fmt.Sprintf("%+v", match.Table) + "\n")
	sb.WriteString("Trump Suit: " + match.TrumpSuit.String() + "\n")
	return sb.String()
}

func (player Player) String() string {
	return player.Name + " (Team " + string(rune(player.TeamId+'0')) + ") - Hand: " + player.Hand.String()
}
