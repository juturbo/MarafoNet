package service

import (
	"MarafoNet/internal/model"
	deckUtils "MarafoNet/internal/utils/deck"
	"fmt"
)

func InitializeGame(players []model.Player) model.Match {
	var match model.Match
	match.Players = initializePlayers(players)
	return match
}

/*starts a full 41 points game*/
func StartGame(match model.Match) model.Match {
	players := initializePlayers(match.Players) // might be removed in future. Double check
	match.Players = players
	return StartMatch(match)
}

/*starts a match in a game*/
func StartMatch(match model.Match) model.Match {
	deck := initializeDeck()
	match.Players, deck = distributeCards(deck, match.Players)
	printMatch(match, deck)
	return match
}

func initializePlayers(players []model.Player) []model.Player {
	var initializedPlayers []model.Player
	for i := range players {
		initializedPlayers = append(initializedPlayers, model.Player{
			TeamId:      i % 2,
			Name:        players[i].Name,
			Hand:        nil,
			MatchPoints: 0,
			TotalPoints: 0,
		})
	}
	return initializedPlayers
}

func initializeDeck() model.Deck {
	return deckUtils.NewShuffledDeck()
}

func distributeCards(deck model.Deck, players []model.Player) ([]model.Player, model.Deck) {
	for i := range len(players) {
		players[i].Hand, deck = deckUtils.DrawCards(deck, model.CardsPerPlayer)
	}
	return players, deck
}

func printMatch(match model.Match, deck model.Deck) {
	for i := 0; i < len(match.Players); i++ {
		fmt.Printf("Player %d (%s): %v\n", i+1, match.Players[i].Name, match.Players[i].Hand)
	}
	fmt.Printf("Table: %+v\n", match.Table)
	fmt.Printf("Remaining deck: %v\n", deck)
}
