package service

import (
	"MarafoNet/internal/model"
	deckUtils "MarafoNet/internal/utils/deck"
	"errors"
	"fmt"
	"math/rand"
)

/*starts a full 41 points game*/
func StartGame(playerNames []string) (model.Game, error) {
	players := make([]model.Player, len(playerNames))

	for i, name := range playerNames {
		players[i] = model.Player{Name: name}
	}

	game, err := initializeGame(players)
	if err != nil {
		return game, err
	}

	return startMatch(game), nil
}

func IsGameEnded(game model.Game) bool {
	return game.WinnerPlayers != nil
}

func ForfeitGame(game model.Game, playerName string) (model.Game, error) {
	playerTeam := -1
	for _, player := range game.Players {
		if player.Name == playerName {
			playerTeam = player.TeamId
			break
		}
	}
	if playerTeam == -1 {
		return game, fmt.Errorf("player %s not found in game", playerName)
	}
	winnerTeam := 1 - playerTeam
	game = setMatchWinner(game, winnerTeam)
	return game, nil
}

func SetTrumpSuit(game model.Game, playerName string, suit model.Suit) (model.Game, error) {
	if isTrumpSuitChosen(game) {
		return game, errors.New("trump suit has already been chosen")
	}
	if !isFirstPlayerTurn(game, playerName) {
		return game, errors.New("only the first player can choose the trump suit")
	}
	game.TrumpSuit = suit
	return game, nil
}

func PlayCard(game model.Game, playerName string, card model.Card) (model.Game, error) {
	if !isTheCardPlayable(game, playerName, card) {
		return game, errors.New("the card is not playable")
	}
	if isEligibleForMarafona(game, playerName, card) {
		teamId := getPlayerTeamId(game, playerName)
		game.MatchPoints[teamId] += model.Point(model.MARAFONA_POINTS)
	}
	game.Table = append(game.Table, model.PlayedCard{
		PlayerName: playerName,
		Card:       card,
	})
	game.Players = removeCardFromPlayerHand(game.Players, playerName, card)
	game.CurrentPlayer = getNextCurrentPlayerName(game)
	if isTableFull(game) {
		game.LastTrick = append(game.LastTrick, game.Table...)
		game = calculateTrickWinnerAndUpdate(game)
	}
	if isMatchOver(game) {
		game = calculateMatchPointsAndReset(game)
		if IsGameEnded(game) {
			return game, nil
		}
		game = startMatch(game)
		return game, nil
	}
	return game, nil
}

func initializeGame(players []model.Player) (model.Game, error) {
	var game model.Game
	var err error
	game.Players, err = initializePlayers(players)
	if err != nil {
		return game, err
	}
	game.FirstPlayer, err = extractFirstPlayer(game.Players)
	if err != nil {
		return game, err
	}
	game.CurrentPlayer = game.FirstPlayer
	return game, nil
}

func initializePlayers(players []model.Player) ([]model.Player, error) {
	if len(players) == 0 {
		return nil, errors.New("at least one player is required to start the game")
	}
	var initializedPlayers []model.Player
	for i := range players {
		initializedPlayers = append(initializedPlayers, model.Player{
			TeamId: i % 2,
			Name:   players[i].Name,
			Hand:   nil,
		})
	}
	return initializedPlayers, nil
}

/*starts a match in a game*/
func startMatch(game model.Game) model.Game {
	deck := initializeDeck()
	game.Players, deck = distributeCards(deck, game.Players)
	game.CurrentPlayer = game.FirstPlayer
	return game
}

func extractFirstPlayer(player []model.Player) (string, error) {
	if len(player) == 0 {
		return "", errors.New("there must be at least one player to extract the first player")
	}
	randomIndex := rand.Intn(len(player))
	return player[randomIndex].Name, nil
}

func getNextFirstPlayerName(game model.Game) string {
	return getNextPlayer(game, game.FirstPlayer)
}

func getNextCurrentPlayerName(game model.Game) string {
	return getNextPlayer(game, game.CurrentPlayer)
}

func getNextPlayer(game model.Game, playerName string) string {
	var currentPlayerIndex int
	for i, player := range game.Players {
		if player.Name == playerName {
			currentPlayerIndex = i
			break
		}
	}
	var nextPlayerIndex = (currentPlayerIndex + 1) % len(game.Players)
	return game.Players[nextPlayerIndex].Name
}

func initializeDeck() model.Deck {
	return deckUtils.NewShuffledDeck()
}

func distributeCards(deck model.Deck, players []model.Player) ([]model.Player, model.Deck) {
	for i := range len(players) {
		players[i].Hand, deck = deckUtils.DrawCards(deck, model.CARDS_PER_PLAYER)
	}
	return players, deck
}

func isTheCardPlayable(game model.Game, playerName string, card model.Card) bool {
	var isValid = false
	if !isTrumpSuitChosen(game) {
		isValid = false
		return isValid
	}
	if !isPlayerTurnValid(game, playerName) {
		isValid = false
		return isValid
	}
	if !playerHasCardInHand(game.Players, playerName, card) {
		isValid = false
		return isValid
	}
	if !isFirstPlayerTurn(game, playerName) {
		if !isCardOfLeadingSuit(game, playerName, card) {
			isValid = false
			return isValid
		}
	}
	isValid = true
	return isValid
}

func isTrumpSuitChosen(game model.Game) bool {
	return game.TrumpSuit != 0
}

func isPlayerTurnValid(game model.Game, playerName string) bool {
	if isTableFull(game) {
		return false
	}
	return game.CurrentPlayer == playerName
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

func isFirstPlayerTurn(game model.Game, playerName string) bool {
	return game.FirstPlayer == playerName
}

func getPlayerTeamId(game model.Game, playerName string) int {
	for _, p := range game.Players {
		if p.Name == playerName {
			return p.TeamId
		}
	}
	return 0
}

func isEligibleForMarafona(game model.Game, playerName string, card model.Card) bool {
	if !isTrumpSuitChosen(game) {
		return false
	}
	if !isFirstPlayerTurn(game, playerName) {
		return false
	}
	if !isTableEmpty(game) {
		return false
	}
	if card.Suit != game.TrumpSuit || card.Rank != model.Ace {
		return false
	}
	for _, player := range game.Players {
		if player.Name != playerName {
			continue
		}
		if len(player.Hand) != model.CARDS_PER_PLAYER {
			return false
		}
		hasAce := false
		hasTwo := false
		hasThree := false
		for _, card := range player.Hand {
			if card.Suit != game.TrumpSuit {
				continue
			}
			switch card.Rank {
			case model.Ace:
				hasAce = true
			case model.Two:
				hasTwo = true
			case model.Three:
				hasThree = true
			}
		}
		return hasAce && hasTwo && hasThree
	}
	return false
}

/*
A card is of the leading suit if it has the same suit as the first card played in the trick.
If a player has a card of the leading suit, they must play it.
If they don't have a card of the leading suit, they can play any card.
*/
func isCardOfLeadingSuit(game model.Game, playerName string, card model.Card) bool {
	var isValid = false
	if !isTableEmpty(game) {
		var leadingSuit = game.Table[0].Card.Suit
		var cardSuitIsLeadingSuit = card.Suit == leadingSuit
		if !cardSuitIsLeadingSuit {
			if playerHasCardOfLeadingSuit(game.Players, playerName, leadingSuit) {
				isValid = false
				return isValid
			}
		}
	}
	isValid = true
	return isValid
}

func playerHasCardOfLeadingSuit(players []model.Player, playerName string, leadingSuit model.Suit) bool {
	var hasCardOfLeadingSuit = false
	var leadingSuitPredicate = func(card model.Card) bool {
		return card.Suit == leadingSuit
	}
	hasCardOfLeadingSuit = playerSatisfies(players, playerName, leadingSuitPredicate)
	return hasCardOfLeadingSuit
}

func isTableFull(game model.Game) bool {
	return len(game.Table) >= len(game.Players)
}

func isTableEmpty(game model.Game) bool {
	return len(game.Table) == 0
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

func calculateTrickWinnerAndUpdate(game model.Game) model.Game {
	winningPlayerName := getTrickWinner(game)
	winningTeamId := getTrickWinningTeamId(game, winningPlayerName)
	trickPoints := calculateTrickPoints(game)
	if isMatchOver(game) {
		trickPoints += model.Point(model.MATCH_END_BONUS_POINTS)
	}
	game.MatchPoints[winningTeamId] += trickPoints
	game.Table = nil
	// The winner of the trick becomes the player who starts the next trick
	game.CurrentPlayer = winningPlayerName
	return game
}

func getTrickWinner(game model.Game) string {
	winningCard := game.Table[0].Card
	winningPlayerName := game.Table[0].PlayerName
	for i := 1; i < len(game.Table); i++ {
		card := game.Table[i].Card
		if card.IsHigherThan(winningCard, game.TrumpSuit) {
			winningCard = card
			winningPlayerName = game.Table[i].PlayerName
		}
	}
	return winningPlayerName
}

func getTrickWinningTeamId(game model.Game, winningPlayerName string) int {
	var winningTeamId int
	for _, player := range game.Players {
		if player.Name == winningPlayerName {
			winningTeamId = player.TeamId
			return winningTeamId
		}
	}
	return winningTeamId
}

func calculateTrickPoints(game model.Game) model.Point {
	var trickPoints model.Point
	for _, playedCard := range game.Table {
		trickPoints += playedCard.Card.PointValue()
	}
	return trickPoints
}

func isMatchOver(game model.Game) bool {
	for _, player := range game.Players {
		if len(player.Hand) > 0 {
			return false
		}
	}
	return true
}

func calculateMatchPointsAndReset(game model.Game) model.Game {
	for i := range model.NUMBER_OF_TEAMS {
		game.TotalPoints[i] += int(game.MatchPoints[i] / model.ACE_POINTS)
		game.MatchPoints[i] = 0
	}
	game, isVictory := checkVictoryAndUpdate(game)
	if isVictory {
		return game
	}
	for i := range game.Players {
		game.Players[i].Hand = nil
	}
	game.Table = nil
	game.LastTrick = nil
	game.FirstPlayer = getNextFirstPlayerName(game)
	game.CurrentPlayer = game.FirstPlayer
	game.TrumpSuit = 0
	return game
}

func checkVictoryAndUpdate(game model.Game) (model.Game, bool) {
	var teamsOver []int
	for i := range game.TotalPoints {
		if game.TotalPoints[i] >= model.POINTS_TO_WIN {
			teamsOver = append(teamsOver, i)
		}
	}
	firstTeamPoints := game.TotalPoints[0]
	secondTeamPoints := game.TotalPoints[1]
	teamsHaveSamePoints := firstTeamPoints == secondTeamPoints
	noTeamIsOver := len(teamsOver) == 0
	if teamsHaveSamePoints || noTeamIsOver {
		return game, false
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
	game = setMatchWinner(game, winner)
	return game, true
}

func setMatchWinner(game model.Game, winnerTeam int) model.Game {
	game.WinnerTeam = &winnerTeam
	game.WinnerPlayers = nil
	for _, player := range game.Players {
		if player.TeamId == winnerTeam {
			game.WinnerPlayers = append(game.WinnerPlayers, player.Name)
		}
	}
	return game
}

func printMatch(game model.Game) {
	for i := 0; i < len(game.Players); i++ {
		fmt.Printf("Team %d\t%s\tHand: %v\n", game.Players[i].TeamId+1, game.Players[i].Name, game.Players[i].Hand)
	}
	for i := range model.NUMBER_OF_TEAMS {
		fmt.Printf("Team %d Match Points: %d, Total Points: %d\n", i+1, game.MatchPoints[i], game.TotalPoints[i])
	}
	fmt.Printf("Table: %+v\n", game.Table)
	fmt.Printf("Trump Suit: %s\n", game.TrumpSuit)
	fmt.Printf("First Player: %s\n\n\n", game.FirstPlayer)
	//fmt.Printf("Remaining deck: %v\n", deck)
}
