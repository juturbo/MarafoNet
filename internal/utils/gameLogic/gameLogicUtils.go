package service

import (
	"MarafoNet/internal/model"
	deckUtils "MarafoNet/internal/utils/deck"
	"errors"
	"fmt"
	"math/rand"
)

func InitializeGame(players []model.Player) model.Match {
	var match model.Match
	match.Players = initializePlayers(players)
	match.FirstPlayer = extractFirstPlayer(match.Players)
	printMatch(match)
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
	printMatch(match)
	return match
}

func SetTrumpSuit(match model.Match, playerName string, suit string) model.Match {
	var trumpSuitAlreadyChosen = match.TrumpSuit != ""
	if trumpSuitAlreadyChosen {
		//TODO: handle this case properly, maybe return an error instead of panicking
		panic(errors.New("trump suit has already been chosen"))
	}
	var isFirstPlayer = match.FirstPlayer == playerName
	if isFirstPlayer {
		match.TrumpSuit = suit
	}
	return match
}

func isTheCardPlayable(match model.Match, playerName string, card model.Card) bool {
	var isValid = false
	var isTrumpSuitChosen = match.TrumpSuit != ""
	if !isTrumpSuitChosen {
		isValid = false
		return isValid
	}
	if !isPlayerTurnValid(match, playerName) {
		isValid = false
		return isValid
	}
	if !playerHasCardInHand(match.Players, playerName, card) {
		isValid = false
		return isValid
	}
	isValid = true
	return isValid
}

func isPlayerTurnValid(match model.Match, playerName string) bool {
	var isValid = false
	var tableIsFull = len(match.Table) >= len(match.Players)
	if tableIsFull {
		isValid = false
		return isValid
	}
	var isFirstTurn = len(match.Table) == 0
	if isFirstTurn {
		isValid = match.FirstPlayer == playerName
		return isValid
	}
	var currentPlayer = lastPlayerToPlay(match)
	isValid = currentPlayer == playerName
	return isValid
}

func playerHasCardInHand(players []model.Player, playerName string, card model.Card) bool {
	var isValid = false
	for _, player := range players {
		var playerNameMatches = player.Name == playerName
		if playerNameMatches {
			for _, cardInHand := range player.Hand {
				var cardMatchesRank = cardInHand.Rank == card.Rank
				var cardMatchesSuit = cardInHand.Suit == card.Suit
				if cardMatchesRank && cardMatchesSuit {
					isValid = true
					return isValid
				}
			}
		}
	}
	isValid = false
	return isValid
}

func lastPlayerToPlay(match model.Match) string {
	var tableIsEmpty = len(match.Table) == 0
	if tableIsEmpty {
		panic(errors.New("no cards have been played yet"))
	}
	indexOfLastPlayer := len(match.Table) - 1
	var lastPlayerName = match.Table[indexOfLastPlayer].PlayerName
	return lastPlayerName
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

func extractFirstPlayer(player []model.Player) string {
	randomIndex := rand.Intn(len(player))
	return player[randomIndex].Name
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

func printMatch(match model.Match) {
	for i := 0; i < len(match.Players); i++ {
		fmt.Printf("Player %d (%s): %v\n", i+1, match.Players[i].Name, match.Players[i].Hand)
	}
	fmt.Printf("Table: %+v\n", match.Table)
	fmt.Printf("Trump Suit: %s\n", match.TrumpSuit)
	fmt.Printf("First Player: %s\n", match.FirstPlayer)
	//fmt.Printf("Remaining deck: %v\n", deck)
}
