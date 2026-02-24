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

func SetTrumpSuit(match model.Match, playerName string, suit model.Suit) model.Match {
	if isTrumpSuitChosen(match) {
		//TODO: handle this case properly, maybe return an error instead of panicking
		panic(errors.New("trump suit has already been chosen"))
	}
	if isFirstPlayerTurn(match, playerName) {
		match.TrumpSuit = suit
	}
	return match
}

func PlayCard(match model.Match, playerName string, card model.Card) model.Match {
	if !isTheCardPlayable(match, playerName, card) {
		//TODO: handle this case properly, maybe return an error instead of panicking
		panic(errors.New("the card is not playable"))
	}
	match.Table = append(match.Table, model.PlayedCard{
		PlayerName: playerName,
		Card:       card,
	})
	removeCardFromPlayerHand(match.Players, playerName, card)
	if isTableFull(match) {
		calculateTrickWinnerAndUpdate(match)
	}
	/*
		TODO: Se nessuno ha più carte in mano ho finito il match -> aggiorno i total points dei giocatori (match/3),
		resetto i match points, resetto la mano dei giocatori, resetto il tavolo,
		estraggo un nuovo primo giocatore e ricomincio con la scelta del seme di briscola
		match.Table = nil
		match.FirstPlayer = nextPlayerName(match)
	*/
	return match
}

func removeCardFromPlayerHand(player []model.Player, playerName string, card model.Card) {
	for _, player := range player {
		if player.Name == playerName {
			for j, cardInHand := range player.Hand {
				if cardInHand.Equal(card) {
					player.Hand = append(player.Hand[:j], player.Hand[j+1:]...)
					return
				}
			}
		}
	}
}

func calculateTrickWinnerAndUpdate(match model.Match) {
	winningPlayerName := getTrickWinner(match)
	winningTeamId := getTrickWinningTeamId(match, winningPlayerName)
	trickPoints := calculateTrickPoints(match)
	updateMatchPoints(match, winningTeamId, trickPoints)
}

func nextPlayerName(match model.Match) string {
	var currentFirstPlayerIndex int
	for i, player := range match.Players {
		if player.Name == match.FirstPlayer {
			currentFirstPlayerIndex = i
			break
		}
	}
	var nextPlayerIndex = (currentFirstPlayerIndex + 1) % len(match.Players)
	return match.Players[nextPlayerIndex].Name
}

func updateMatchPoints(match model.Match, winningTeamId int, trickPoints model.Point) {
	for _, player := range match.Players {
		if player.TeamId == winningTeamId {
			player.MatchPoints += trickPoints
		}
	}
}

func calculateTrickPoints(match model.Match) model.Point {
	var trickPoints model.Point
	for _, playedCard := range match.Table {
		trickPoints += playedCard.Card.PointValue()
	}
	return trickPoints
}

func getTrickWinningTeamId(match model.Match, winningPlayerName string) int {
	var winningTeamId int
	for _, player := range match.Players {
		if player.Name == winningPlayerName {
			winningTeamId = player.TeamId
			break
		}
	}
	return winningTeamId
}

func getTrickWinner(match model.Match) string {
	var winningCard model.Card
	var winningPlayerName string
	winningCard = match.Table[0].Card
	winningPlayerName = match.Table[0].PlayerName
	for i := 1; i < len(match.Table); i++ {
		card := match.Table[i].Card
		if card.IsHigherThan(winningCard, match.TrumpSuit) {
			winningCard = card
			winningPlayerName = match.Table[i].PlayerName
		}
	}
	return winningPlayerName
}

func isTrumpSuitChosen(match model.Match) bool {
	return match.TrumpSuit != 0
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
