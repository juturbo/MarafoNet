package service

import (
	"MarafoNet/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializeGame(t *testing.T) {
	expectedPlayers := []model.Player{
		{Name: "Player 1", TeamId: 0},
		{Name: "Player 2", TeamId: 1},
		{Name: "Player 3", TeamId: 0},
		{Name: "Player 4", TeamId: 1},
	}
	match := initializeGame(expectedPlayers)
	for i := range expectedPlayers {
		assert.Equal(t, expectedPlayers[i], match.Players[i], "Expected player %v, got %v", expectedPlayers[i], match.Players[i])
	}
}

/*
// starts a full 41 points game
func StartGame(match model.Match) model.Match {
	players := initializePlayers(match.Players) // might be removed in future. Double check
	match.Players = players
	return StartMatch(match)
}

// starts a match in a game
func StartMatch(match model.Match) model.Match {
	deck := initializeDeck()
	match.Players, deck = distributeCards(deck, match.Players)
	printMatch(match, deck)
	return match
}

func initializePlayers(players []model.Player) []model.Player {
	for i := range len(players) {
		players[i] = model.Player{
			TeamId:      i % 2,
			Name:        players[i].Name,
			Hand:        nil,
			MatchPoints: 0,
			TotalPoints: 0,
		}
	}
	return players
}

func inizializeDeck() model.Deck {
	return deckUtils.NewShuffledDeck()
}

func distributeCards(deck model.Deck, players []model.Player) ([]model.Player, model.Deck) {
	for i := range len(players) {
		players[i].Hand, deck = deckUtils.DrawCards(deck, constant.CardsPerPlayer)
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
*/
