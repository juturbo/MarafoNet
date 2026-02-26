package service

import (
	"MarafoNet/internal/model"
	deckUtils "MarafoNet/internal/utils/deck"
	"errors"
	"fmt"
	"math/rand"
)

/*starts a full 41 points game*/
func StartGame(players []model.Player) model.Match {
	match := initializeGame(players)
	return startMatch(match)
}

func SetTrumpSuit(match model.Match, playerName string, suit model.Suit) (model.Match, error) {
	if isTrumpSuitChosen(match) {
		return match, errors.New("trump suit has already been chosen")
	}
	if isFirstPlayerTurn(match, playerName) {
		match.TrumpSuit = suit
	}
	return match, nil
}

func PlayCard(match model.Match, playerName string, card model.Card) (model.Match, error) {
	if !isTheCardPlayable(match, playerName, card) {
		return match, errors.New("the card is not playable")
	}
	match.Table = append(match.Table, model.PlayedCard{
		PlayerName: playerName,
		Card:       card,
	})
	match.Players = removeCardFromPlayerHand(match.Players, playerName, card)
	if isTableFull(match) {
		match = calculateTrickWinnerAndUpdate(match)
	}
	if isMatchOver(match) {
		match = calculateMatchPointsAndReset(match)
		if match.WinnerPlayers != nil {
			return match, nil
		}
		match = startMatch(match)
		return match, nil
	}
	return match, nil
}

func initializeGame(players []model.Player) model.Match {
	var match model.Match
	match.Players = initializePlayers(players)
	match.FirstPlayer = extractFirstPlayer(match.Players)
	return match
}

func initializePlayers(players []model.Player) []model.Player {
	var initializedPlayers []model.Player
	for i := range players {
		initializedPlayers = append(initializedPlayers, model.Player{
			TeamId: i % 2,
			Name:   players[i].Name,
			Hand:   nil,
		})
	}
	return initializedPlayers
}

/*starts a match in a game*/
func startMatch(match model.Match) model.Match {
	deck := initializeDeck()
	match.Players, deck = distributeCards(deck, match.Players)
	return match
}

func extractFirstPlayer(player []model.Player) string {
	randomIndex := rand.Intn(len(player))
	return player[randomIndex].Name
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

func initializeDeck() model.Deck {
	return deckUtils.NewShuffledDeck()
}

func distributeCards(deck model.Deck, players []model.Player) ([]model.Player, model.Deck) {
	for i := range len(players) {
		players[i].Hand, deck = deckUtils.DrawCards(deck, model.CardsPerPlayer)
	}
	return players, deck
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

func isTrumpSuitChosen(match model.Match) bool {
	return match.TrumpSuit != 0
}

func isPlayerTurnValid(match model.Match, playerName string) bool {
	if isTableFull(match) {
		return false
	}
	return getCurrentPlayer(match) == playerName
}

func getCurrentPlayer(match model.Match) string {
	if isTableEmpty(match) {
		return match.FirstPlayer
	}
	indexOfLastPlayer := len(match.Table) - 1
	var lastPlayerName = match.Table[indexOfLastPlayer].PlayerName
	currentPlayerName := match.FirstPlayer
	for i, player := range match.Players {
		if player.Name == lastPlayerName {
			currentPlayerName = match.Players[(i+1)%len(match.Players)].Name
			break
		}
	}
	return currentPlayerName
}

func playerHasCardInHand(players []model.Player, playerName string, playedCard model.Card) bool {
	var hasCardInHand = false
	var cardInHandPredicate = func(cardInHand model.Card) bool {
		return cardInHand.Equal(playedCard)
	}
	hasCardInHand = playerSatisfies(players, playerName, cardInHandPredicate)
	return hasCardInHand
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

func isFirstPlayerTurn(match model.Match, playerName string) bool {
	return match.FirstPlayer == playerName
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

func isTableFull(match model.Match) bool {
	return len(match.Table) >= len(match.Players)
}

func isTableEmpty(match model.Match) bool {
	return len(match.Table) == 0
}

func removeCardFromPlayerHand(players []model.Player, playerName string, card model.Card) []model.Player {
	for i, player := range players {
		if player.Name == playerName {
			for j, cardInHand := range player.Hand {
				if cardInHand.Equal(card) {
					players[i].Hand = append(player.Hand[:j], player.Hand[j+1:]...)
					return players
				}
			}
		}
	}
	return players
}

func calculateTrickWinnerAndUpdate(match model.Match) model.Match {
	winningPlayerName := getTrickWinner(match)
	winningTeamId := getTrickWinningTeamId(match, winningPlayerName)
	trickPoints := calculateTrickPoints(match)
	match.MatchPoints[winningTeamId] += trickPoints
	match.Table = nil
	return match
}

func getTrickWinner(match model.Match) string {
	winningCard := match.Table[0].Card
	winningPlayerName := match.Table[0].PlayerName
	for i := 1; i < len(match.Table); i++ {
		card := match.Table[i].Card
		if card.IsHigherThan(winningCard, match.TrumpSuit) {
			winningCard = card
			winningPlayerName = match.Table[i].PlayerName
		}
	}
	return winningPlayerName
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

func calculateTrickPoints(match model.Match) model.Point {
	var trickPoints model.Point
	for _, playedCard := range match.Table {
		trickPoints += playedCard.Card.PointValue()
	}
	return trickPoints
}

func isMatchOver(match model.Match) bool {
	for _, player := range match.Players {
		if len(player.Hand) > 0 {
			return false
		}
	}
	return true
}

func calculateMatchPointsAndReset(match model.Match) model.Match {
	for i := range model.NumberOfTeams {
		match.TotalPoints[i] += int(match.MatchPoints[i] / model.AcePoints)
		match.MatchPoints[i] = 0
	}
	match, isVictory := checkVictoryAndUpdate(match)
	if isVictory {
		return match
	}
	for i := range match.Players {
		match.Players[i].Hand = nil
	}
	match.Table = nil
	match.FirstPlayer = nextPlayerName(match)
	match.TrumpSuit = 0
	return match
}

func checkVictoryAndUpdate(match model.Match) (model.Match, bool) {
	var teamsOver []int
	for i := range match.TotalPoints {
		if match.TotalPoints[i] >= model.PointsToWin {
			teamsOver = append(teamsOver, i)
		}
	}
	firstTeamPoints := match.TotalPoints[0]
	secondTeamPoints := match.TotalPoints[1]
	teamsHaveSamePoints := firstTeamPoints == secondTeamPoints
	noTeamIsOver := len(teamsOver) == 0
	if teamsHaveSamePoints || noTeamIsOver {
		return match, false
	}
	var winner int
	oneTeamIsOver := len(teamsOver) == 1
	if oneTeamIsOver {
		winner = teamsOver[0]
	}
	if firstTeamPoints > secondTeamPoints {
		winner = 0
	}
	if secondTeamPoints > firstTeamPoints {
		winner = 1
	}
	match.WinnerTeam = &winner
	match.WinnerPlayers = nil
	for _, p := range match.Players {
		if p.TeamId == winner {
			match.WinnerPlayers = append(match.WinnerPlayers, p.Name)
		}
	}
	return match, true
}

func printMatch(match model.Match) {
	for i := 0; i < len(match.Players); i++ {
		fmt.Printf("Team %d\t%s\tHand: %v\n", match.Players[i].TeamId+1, match.Players[i].Name, match.Players[i].Hand)
	}
	for i := range model.NumberOfTeams {
		fmt.Printf("Team %d Match Points: %d, Total Points: %d\n", i+1, match.MatchPoints[i], match.TotalPoints[i])
	}
	fmt.Printf("Table: %+v\n", match.Table)
	fmt.Printf("Trump Suit: %s\n", match.TrumpSuit)
	fmt.Printf("First Player: %s\n\n\n", match.FirstPlayer)
	//fmt.Printf("Remaining deck: %v\n", deck)
}
