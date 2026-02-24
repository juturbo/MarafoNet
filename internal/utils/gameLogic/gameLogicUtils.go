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
	if isTrumpSuitChosen(match) {
		//TODO: handle this case properly, maybe return an error instead of panicking
		panic(errors.New("trump suit has already been chosen"))
	}
	if isFirstPlayerTurn(match, playerName) {
		match.TrumpSuit = suit
	}
	return match
}

func isTrumpSuitChosen(match model.Match) bool {
	return match.TrumpSuit != ""
}

func isFirstPlayerTurn(match model.Match, playerName string) bool {
	return match.FirstPlayer == playerName
}

func isTheCardPlayable(match model.Match, playerName string, card model.Card) bool {
	var isValid = false
	if !isTrumpSuitChosen(match) {
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
	if !isFirstPlayerTurn(match, playerName) {
		if !isCardOfLeadingSuit(match, playerName, card) {
			isValid = false
			return isValid
		}
	}
	isValid = true
	return isValid
}

func isPlayerTurnValid(match model.Match, playerName string) bool {
	if isTableFull(match) {
		return false
	}
	return getCurrentPlayer(match) == playerName
}

func playerHasCardInHand(players []model.Player, playerName string, playedCard model.Card) bool {
	var hasCardInHand = false
	var cardInHandPredicate = func(cardInHand model.Card) bool {
		return cardInHand.Equal(playedCard)
	}
	hasCardInHand = playerSatisfies(players, playerName, cardInHandPredicate)
	return hasCardInHand
}

/*
A card is of the leading suit if it has the same suit as the first card played in the trick.
If a player has a card of the leading suit, they must play it.
If they don't have a card of the leading suit, they can play any card.
*/
func isCardOfLeadingSuit(match model.Match, playerName string, card model.Card) bool {
	var isValid = false
	if !isTableEmpty(match) {
		var leadingSuit = match.Table[0].Card.Suit
		var cardSuitIsLeadingSuit = card.Suit == leadingSuit
		if !cardSuitIsLeadingSuit {
			if playerHasCardOfLeadingSuit(match.Players, playerName, leadingSuit) {
				isValid = false
				return isValid
			}
		}
	}
	isValid = true
	return isValid
	/*
		if isTableEmpty(match) { // if the table is empty, any card can be played, so we consider it valid
			isValid = true
			return isValid
		}
		var leadingSuit = match.Table[0].Card.Suit
		var cardSuitIsLeadingSuit = card.Suit == leadingSuit
		if cardSuitIsLeadingSuit {
			isValid = true
			return isValid
		}
		if !playerHasCardOfLeadingSuit(match.Players, playerName, leadingSuit) {
			isValid = true
			return isValid
		}
		isValid = false
		return isValid
	*/
}

func playerHasCardOfLeadingSuit(players []model.Player, playerName string, leadingSuit model.Suit) bool {
	var hasCardOfLeadingSuit = false
	var leadingSuitPredicate = func(card model.Card) bool {
		return card.Suit == leadingSuit
	}
	hasCardOfLeadingSuit = playerSatisfies(players, playerName, leadingSuitPredicate)
	return hasCardOfLeadingSuit
}

func playerSatisfies(players []model.Player, playerName string, predicate func(model.Card) bool) bool {
	var isValid = false
	for _, player := range players {
		if player.Name == playerName {
			for _, cardInHand := range player.Hand {
				if predicate(cardInHand) {
					isValid = true
					return isValid
				}
			}
			isValid = false
			return isValid
		}
	}
	isValid = false
	return isValid
}

func getCurrentPlayer(match model.Match) string {
	if isTableEmpty(match) {
		return match.FirstPlayer
	}
	indexOfLastPlayer := len(match.Table) - 1
	var lastPlayerName = match.Table[indexOfLastPlayer].PlayerName
	return lastPlayerName
}

func isTableFull(match model.Match) bool {
	return len(match.Table) >= len(match.Players)
}

func isTableEmpty(match model.Match) bool {
	return len(match.Table) == 0
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
