package model

import (
	"fmt"
	"strings"
)

const (
	CARDS_PER_PLAYER       = 10
	NUMBER_OF_TEAMS        = 2
	POINTS_TO_WIN          = 41
	MATCH_END_BONUS_POINTS = 3
)

type Point uint8

type Player struct {
	TeamId int    `json:"TeamId"`
	Name   string `json:"Name"`
	Hand   Hand   `json:"Hand"`
}

type PlayerView struct {
	TeamId int    `json:"TeamId"`
	Name   string `json:"Name"`
}

type PlayedCard struct {
	PlayerName string `json:"PlayerName"`
	Card       Card   `json:"Card"`
}

type Game struct {
	Players []Player     `json:"Players"`
	Table   []PlayedCard `json:"Table"`
	// TODO: add a field to track the passed card played in the previeous round.
	MatchPoints   [NUMBER_OF_TEAMS]Point `json:"MatchPoints"`
	TotalPoints   [NUMBER_OF_TEAMS]int   `json:"TotalPoints"`
	FirstPlayer   string                 `json:"FirstPlayer"`
	CurrentPlayer string                 `json:"CurrentPlayer"`
	TrumpSuit     Suit                   `json:"TrumpSuit"`
	WinnerTeam    *int                   `json:"WinnerTeam,omitempty"`
	WinnerPlayers []string               `json:"WinnerPlayers,omitempty"`
}

type GameView struct {
	Players       []PlayerView           `json:"Players"`
	PlayerHand    Hand                   `json:"Hand"`
	Table         []PlayedCard           `json:"Table"`
	MatchPoints   [NUMBER_OF_TEAMS]Point `json:"MatchPoints"`
	TotalPoints   [NUMBER_OF_TEAMS]int   `json:"TotalPoints"`
	FirstPlayer   string                 `json:"FirstPlayer"`
	CurrentPlayer string                 `json:"CurrentPlayer"`
	TrumpSuit     Suit                   `json:"TrumpSuit"`
	WinnerTeam    *int                   `json:"WinnerTeam,omitempty"`
	WinnerPlayers []string               `json:"WinnerPlayers,omitempty"`
}

func (game Game) ViewForPlayer(playerName string) (GameView, error) {
	gameView := GameView{
		Table:         game.Table,
		MatchPoints:   game.MatchPoints,
		TotalPoints:   game.TotalPoints,
		FirstPlayer:   game.FirstPlayer,
		CurrentPlayer: game.CurrentPlayer,
		TrumpSuit:     game.TrumpSuit,
		WinnerTeam:    game.WinnerTeam,
		WinnerPlayers: game.WinnerPlayers,
	}

	playerFound := false
	for _, player := range game.Players {
		gameView.Players = append(gameView.Players, PlayerView{
			TeamId: player.TeamId,
			Name:   player.Name,
		})
		if player.Name == playerName {
			gameView.PlayerHand = player.Hand
			playerFound = true
		}
	}
	if !playerFound {
		return GameView{}, fmt.Errorf("player %q not found in game", playerName)
	}
	return gameView, nil
}

// TODO: check
func (game Game) String() string {
	var sb strings.Builder
	for i, player := range game.Players {
		sb.WriteString("Player " + string(rune(i+1)) + ": " + player.String() + "\n")
	}
	sb.WriteString("Table: " + fmt.Sprintf("%+v", game.Table) + "\n")
	sb.WriteString("Trump Suit: " + game.TrumpSuit.String() + "\n")
	return sb.String()
}

func (player Player) String() string {
	return player.Name + " (Team " + string(rune(player.TeamId+'0')) + ") - Hand: " + player.Hand.String()
}
