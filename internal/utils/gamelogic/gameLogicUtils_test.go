package service

import (
	"MarafoNet/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializeGame(t *testing.T) {
	names := []string{"Alice", "Bob", "Carol", "Dave"}
	players := makePlayers(names...)
	game, err := initializeGame(players)
	assert.NoError(t, err)
	assert.Equal(t, players, game.Players)
}

func TestStartMatchDealsCards(t *testing.T) {
	game := makeGame("Alice", "Bob", "Carol", "Dave")
	game = startMatch(game)
	for _, player := range game.Players {
		assert.Equal(t, model.CARDS_PER_PLAYER, len(player.Hand), "Each player should have %d cards", model.CARDS_PER_PLAYER)
	}
}

func TestSetTrumpSuitOnlyFirstPlayer(t *testing.T) {
	game := makeGame("Alice", "Bob", "Carol", "Dave")
	var err error
	game, err = SetTrumpSuit(game, "Alice", model.Swords)
	assert.NoError(t, err)
	assert.Equal(t, model.Swords, game.TrumpSuit)
	_, err = SetTrumpSuit(game, "Bob", model.Cups)
	assert.Error(t, err)
}

func TestPlayCardRemovesCardAndTotals(t *testing.T) {
	game := makeGame("Alice", "Bob", "Carol", "Dave")
	aliceHand := []model.Card{{Suit: model.Swords, Rank: model.Ace}}
	bobHand := []model.Card{{Suit: model.Swords, Rank: model.King}}
	carolHand := []model.Card{{Suit: model.Swords, Rank: model.Three}}
	daveHand := []model.Card{{Suit: model.Swords, Rank: model.Seven}}
	hands := [][]model.Card{aliceHand, bobHand, carolHand, daveHand}
	setHands(&game, hands)
	game.TrumpSuit = model.Swords
	var err error
	game, err = PlayCard(game, "Alice", aliceHand[0])
	assert.NoError(t, err)
	for _, player := range game.Players {
		if player.Name == "Alice" {
			assert.Equal(t, 0, len(player.Hand))
		}
	}
}

func TestCalculateTrickWinner(t *testing.T) {
	players := makePlayers("Alice", "Bob", "Carol", "Dave")
	game := model.Game{Players: players, TrumpSuit: model.Cups}
	game.Table = []model.PlayedCard{
		{PlayerName: "Alice", Card: model.Card{Suit: model.Swords, Rank: model.Seven}},
		{PlayerName: "Bob", Card: model.Card{Suit: model.Swords, Rank: model.Knight}},
		{PlayerName: "Carol", Card: model.Card{Suit: model.Swords, Rank: model.King}},
		{PlayerName: "Dave", Card: model.Card{Suit: model.Swords, Rank: model.Ace}},
	}
	winner := getTrickWinner(game)
	assert.Equal(t, "Dave", winner)
}

func TestTrickWinnerWithTrump(t *testing.T) {
	game := makeGame("Alice", "Bob", "Carol", "Dave")
	game.TrumpSuit = model.Cups
	game.Table = []model.PlayedCard{
		{PlayerName: "Alice", Card: model.Card{Suit: model.Swords, Rank: model.Seven}},
		{PlayerName: "Bob", Card: model.Card{Suit: model.Swords, Rank: model.King}},
		{PlayerName: "Carol", Card: model.Card{Suit: model.Coins, Rank: model.Ace}},
		{PlayerName: "Dave", Card: model.Card{Suit: model.Cups, Rank: model.Two}},
	}
	winner := getTrickWinner(game)
	assert.Equal(t, "Dave", winner)
}

func TestCalculateTrickPoints(t *testing.T) {
	game := makeGame("Alice", "Bob", "Carol", "Dave")
	game.Table = []model.PlayedCard{
		{PlayerName: "Alice", Card: model.Card{Suit: model.Swords, Rank: model.Ace}},
		{PlayerName: "Bob", Card: model.Card{Suit: model.Swords, Rank: model.Two}},
		{PlayerName: "Carol", Card: model.Card{Suit: model.Swords, Rank: model.Four}},
		{PlayerName: "Dave", Card: model.Card{Suit: model.Swords, Rank: model.King}},
	}
	pts := calculateTrickPoints(game)
	assert.Equal(t, model.Point(5), pts)
}

func TestGetCurrentPlayerBehavior(t *testing.T) {
	game := makeGame("Alice", "Bob", "Carol", "Dave")
	assert.Equal(t, "Alice", game.CurrentPlayer)
	hand := make([]model.Card, model.CARDS_PER_PLAYER)
	hand[0] = model.Card{Suit: model.Swords, Rank: model.Two}
	hand[1] = model.Card{Suit: model.Swords, Rank: model.Ace}
	hand[2] = model.Card{Suit: model.Swords, Rank: model.Three}
	for i := 3; i < model.CARDS_PER_PLAYER; i++ {
		hand[i] = model.Card{Suit: model.Swords, Rank: model.Four}
	}
	setHands(&game, [][]model.Card{hand})
	game, err := SetTrumpSuit(game, "Alice", model.Clubs)
	assert.NoError(t, err)
	game, err = PlayCard(game, "Alice", model.Card{Suit: model.Swords, Rank: model.Ace})
	assert.NoError(t, err)
	assert.Equal(t, "Bob", game.CurrentPlayer)
}

func TestIsCardOfLeadingSuitAndPlayable(t *testing.T) {
	game := makeGame("Alice", "Bob", "Carol", "Dave")
	setHands(&game, [][]model.Card{{{Suit: model.Swords, Rank: model.Ace}, {Suit: model.Cups, Rank: model.Two}}, {{Suit: model.Swords, Rank: model.King}}})
	game.TrumpSuit = model.Cups
	game.Table = []model.PlayedCard{{PlayerName: "Alice", Card: model.Card{Suit: model.Swords, Rank: model.Four}}}
	isLeading := isCardOfLeadingSuit(game, "Bob", model.Card{Suit: model.Cups, Rank: model.Two})
	assert.False(t, isLeading)
	isLeading = isCardOfLeadingSuit(game, "Bob", model.Card{Suit: model.Swords, Rank: model.King})
	assert.True(t, isLeading)
}

func TestIsTheCardPlayableVarious(t *testing.T) {
	game := makeGame("Alice", "Bob", "Carol", "Dave")
	setHands(&game, [][]model.Card{{{Suit: model.Swords, Rank: model.Ace}}, {{Suit: model.Swords, Rank: model.King}}})
	ok := isTheCardPlayable(game, "Alice", model.Card{Suit: model.Swords, Rank: model.Ace})
	assert.False(t, ok)
	game.TrumpSuit = model.Cups
	ok = isTheCardPlayable(game, "Bob", model.Card{Suit: model.Swords, Rank: model.King})
	assert.False(t, ok)
	ok = isTheCardPlayable(game, "Alice", model.Card{Suit: model.Swords, Rank: model.Ace})
	assert.True(t, ok)
}

func TestMarafonaBonusAwarded(t *testing.T) {
	game := makeGame("Alice", "Bob", "Carol", "Dave")
	hand := make([]model.Card, model.CARDS_PER_PLAYER)
	hand[0] = model.Card{Suit: model.Cups, Rank: model.Ace}
	hand[1] = model.Card{Suit: model.Cups, Rank: model.Two}
	hand[2] = model.Card{Suit: model.Cups, Rank: model.Three}
	for i := 3; i < model.CARDS_PER_PLAYER; i++ {
		hand[i] = model.Card{Suit: model.Swords, Rank: model.Four}
	}
	setHands(&game, [][]model.Card{hand})
	game.TrumpSuit = model.Cups

	updated, err := PlayCard(game, "Alice", hand[0])
	assert.NoError(t, err)
	teamId := getPlayerTeamId(game, "Alice")
	assert.Equal(t, model.Point(model.MARAFONA_POINTS), updated.MatchPoints[teamId])
}

func TestMarafonaNotAwardedIfNotAllCards(t *testing.T) {
	game := makeGame("Alice", "Bob", "Carol", "Dave")
	hand := []model.Card{{Suit: model.Cups, Rank: model.Ace}, {Suit: model.Cups, Rank: model.Two}, {Suit: model.Cups, Rank: model.Three}}
	setHands(&game, [][]model.Card{hand})
	game.TrumpSuit = model.Cups
	updated, err := PlayCard(game, "Alice", hand[0])
	assert.NoError(t, err)
	teamId := getPlayerTeamId(game, "Alice")
	assert.Equal(t, model.Point(0), updated.MatchPoints[teamId])
}

func TestMarafonaNotAwardedIfNotFirstPlayer(t *testing.T) {
	game := makeGame("Alice", "Bob", "Carol", "Dave")
	hand := make([]model.Card, model.CARDS_PER_PLAYER)
	hand[0] = model.Card{Suit: model.Cups, Rank: model.Ace}
	hand[1] = model.Card{Suit: model.Cups, Rank: model.Two}
	hand[2] = model.Card{Suit: model.Cups, Rank: model.Three}
	for i := 3; i < model.CARDS_PER_PLAYER; i++ {
		hand[i] = model.Card{Suit: model.Swords, Rank: model.Four}
	}
	setHands(&game, [][]model.Card{{}, hand})
	game.TrumpSuit = model.Cups
	game.Table = []model.PlayedCard{{PlayerName: "Alice", Card: model.Card{Suit: model.Cups, Rank: model.Four}}}
	updated, _ := PlayCard(game, "Bob", hand[0])
	teamId := getPlayerTeamId(game, "Bob")
	assert.Equal(t, model.Point(0), updated.MatchPoints[teamId])
}

func TestMarafonaNotAwardedIfFirstCardNotAce(t *testing.T) {
	game := makeGame("Alice", "Bob", "Carol", "Dave")
	hand := make([]model.Card, model.CARDS_PER_PLAYER)
	hand[0] = model.Card{Suit: model.Cups, Rank: model.Two}
	hand[1] = model.Card{Suit: model.Cups, Rank: model.Ace}
	hand[2] = model.Card{Suit: model.Cups, Rank: model.Three}
	for i := 3; i < model.CARDS_PER_PLAYER; i++ {
		hand[i] = model.Card{Suit: model.Swords, Rank: model.Four}
	}
	setHands(&game, [][]model.Card{hand})
	game.TrumpSuit = model.Cups
	updated, err := PlayCard(game, "Alice", hand[0])
	assert.NoError(t, err)
	teamId := getPlayerTeamId(game, "Alice")
	assert.Equal(t, model.Point(0), updated.MatchPoints[teamId])
}

func makePlayers(names ...string) []model.Player {
	players := make([]model.Player, len(names))
	for i, n := range names {
		players[i] = model.Player{Name: n, TeamId: i % 2}
	}
	return players
}

func makeGame(names ...string) model.Game {
	players := makePlayers(names...)
	return model.Game{Players: players, FirstPlayer: names[0], CurrentPlayer: names[0]}
}

func setHands(game *model.Game, hands [][]model.Card) {
	for i := range hands {
		if i < len(game.Players) {
			game.Players[i].Hand = model.Hand(hands[i])
		}
	}
}
