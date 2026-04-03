package service

import (
	"MarafoNet/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializeGame(t *testing.T) {
	names := []string{"Alice", "Bob", "Carol", "Dave"}
	players := mkPlayers(names...)
	match, err := initializeGame(players)
	assert.NoError(t, err)
	assert.Equal(t, players, match.Players)
}

func TestStartMatchDealsCards(t *testing.T) {
	match := mkMatch("Alice", "Bob", "Carol", "Dave")
	match = startMatch(match)
	for _, player := range match.Players {
		assert.Equal(t, model.CardsPerPlayer, len(player.Hand), "Each player should have %d cards", model.CardsPerPlayer)
	}
}

func TestSetTrumpSuitOnlyFirstPlayer(t *testing.T) {
	match := mkMatch("Alice", "Bob", "Carol", "Dave")
	var err error
	match, err = SetTrumpSuit(match, "Alice", model.Swords)
	assert.NoError(t, err)
	assert.Equal(t, model.Swords, match.TrumpSuit)
	_, err = SetTrumpSuit(match, "Bob", model.Cups)
	assert.Error(t, err)
}

func TestPlayCardRemovesCardAndTotals(t *testing.T) {
	match := mkMatch("Alice", "Bob", "Carol", "Dave")
	aliceHand := []model.Card{{Suit: model.Swords, Rank: model.Ace}}
	bobHand := []model.Card{{Suit: model.Swords, Rank: model.King}}
	carolHand := []model.Card{{Suit: model.Swords, Rank: model.Three}}
	daveHand := []model.Card{{Suit: model.Swords, Rank: model.Seven}}
	hands := [][]model.Card{aliceHand, bobHand, carolHand, daveHand}
	setHands(&match, hands)
	match.TrumpSuit = model.Swords
	var err error
	match, err = PlayCard(match, "Alice", aliceHand[0])
	assert.NoError(t, err)
	for _, player := range match.Players {
		if player.Name == "Alice" {
			assert.Equal(t, 0, len(player.Hand))
		}
	}
}

func TestCalculateTrickWinner(t *testing.T) {
	players := mkPlayers("Alice", "Bob", "Carol", "Dave")
	match := model.Game{Players: players, TrumpSuit: model.Cups}
	match.Table = []model.PlayedCard{
		{PlayerName: "Alice", Card: model.Card{Suit: model.Swords, Rank: model.Seven}},
		{PlayerName: "Bob", Card: model.Card{Suit: model.Swords, Rank: model.Knight}},
		{PlayerName: "Carol", Card: model.Card{Suit: model.Swords, Rank: model.King}},
		{PlayerName: "Dave", Card: model.Card{Suit: model.Coins, Rank: model.Ace}},
	}
	winner := getTrickWinner(match)
	assert.Equal(t, "Carol", winner)
}

func TestTrickWinnerWithTrump(t *testing.T) {
	match := mkMatch("Alice", "Bob", "Carol", "Dave")
	match.TrumpSuit = model.Cups
	match.Table = []model.PlayedCard{
		{PlayerName: "Alice", Card: model.Card{Suit: model.Swords, Rank: model.Seven}},
		{PlayerName: "Bob", Card: model.Card{Suit: model.Swords, Rank: model.King}},
		{PlayerName: "Carol", Card: model.Card{Suit: model.Coins, Rank: model.Ace}},
		{PlayerName: "Dave", Card: model.Card{Suit: model.Cups, Rank: model.Two}},
	}
	winner := getTrickWinner(match)
	assert.Equal(t, "Dave", winner)
}

func TestCalculateTrickPoints(t *testing.T) {
	match := mkMatch("Alice", "Bob", "Carol", "Dave")
	match.Table = []model.PlayedCard{
		{PlayerName: "Alice", Card: model.Card{Suit: model.Swords, Rank: model.Ace}},
		{PlayerName: "Bob", Card: model.Card{Suit: model.Swords, Rank: model.Two}},
		{PlayerName: "Carol", Card: model.Card{Suit: model.Swords, Rank: model.Four}},
		{PlayerName: "Dave", Card: model.Card{Suit: model.Swords, Rank: model.King}},
	}
	pts := calculateTrickPoints(match)
	assert.Equal(t, model.Point(5), pts)
}

func TestGetCurrentPlayerBehavior(t *testing.T) {
	match := mkMatch("Alice", "Bob", "Carol", "Dave")
	assert.Equal(t, "Alice", getCurrentPlayer(match))
	match.Table = []model.PlayedCard{{PlayerName: "Alice", Card: model.Card{Suit: model.Swords, Rank: model.Ace}}}
	assert.Equal(t, "Bob", getCurrentPlayer(match))
}

func TestIsCardOfLeadingSuitAndPlayable(t *testing.T) {
	match := mkMatch("Alice", "Bob", "Carol", "Dave")
	setHands(&match, [][]model.Card{{{Suit: model.Swords, Rank: model.Ace}, {Suit: model.Cups, Rank: model.Two}}, {{Suit: model.Swords, Rank: model.King}}})
	match.TrumpSuit = model.Cups
	match.Table = []model.PlayedCard{{PlayerName: "Alice", Card: model.Card{Suit: model.Swords, Rank: model.Four}}}
	isLeading := isCardOfLeadingSuit(match, "Bob", model.Card{Suit: model.Cups, Rank: model.Two})
	assert.False(t, isLeading)
	isLeading = isCardOfLeadingSuit(match, "Bob", model.Card{Suit: model.Swords, Rank: model.King})
	assert.True(t, isLeading)
}

func TestIsTheCardPlayableVarious(t *testing.T) {
	match := mkMatch("Alice", "Bob", "Carol", "Dave")
	setHands(&match, [][]model.Card{{{Suit: model.Swords, Rank: model.Ace}}, {{Suit: model.Swords, Rank: model.King}}})
	ok := isTheCardPlayable(match, "Alice", model.Card{Suit: model.Swords, Rank: model.Ace})
	assert.False(t, ok)
	match.TrumpSuit = model.Cups
	ok = isTheCardPlayable(match, "Bob", model.Card{Suit: model.Swords, Rank: model.King})
	assert.False(t, ok)
	ok = isTheCardPlayable(match, "Alice", model.Card{Suit: model.Swords, Rank: model.Ace})
	assert.True(t, ok)
}

func TestMarafonaBonusAwarded(t *testing.T) {
	match := mkMatch("Alice", "Bob", "Carol", "Dave")
	hand := make([]model.Card, model.CardsPerPlayer)
	hand[0] = model.Card{Suit: model.Cups, Rank: model.Ace}
	hand[1] = model.Card{Suit: model.Cups, Rank: model.Two}
	hand[2] = model.Card{Suit: model.Cups, Rank: model.Three}
	for i := 3; i < model.CardsPerPlayer; i++ {
		hand[i] = model.Card{Suit: model.Swords, Rank: model.Four}
	}
	setHands(&match, [][]model.Card{hand})
	match.TrumpSuit = model.Cups

	updated, err := PlayCard(match, "Alice", hand[0])
	assert.NoError(t, err)
	teamId := getPlayerTeamId(match, "Alice")
	assert.Equal(t, model.Point(model.MARAFONA_POINTS), updated.MatchPoints[teamId])
}

func TestMarafonaNotAwardedIfNotAllCards(t *testing.T) {
	match := mkMatch("Alice", "Bob", "Carol", "Dave")
	hand := []model.Card{{Suit: model.Cups, Rank: model.Ace}, {Suit: model.Cups, Rank: model.Two}, {Suit: model.Cups, Rank: model.Three}}
	setHands(&match, [][]model.Card{hand})
	match.TrumpSuit = model.Cups
	updated, err := PlayCard(match, "Alice", hand[0])
	assert.NoError(t, err)
	teamId := getPlayerTeamId(match, "Alice")
	assert.Equal(t, model.Point(0), updated.MatchPoints[teamId])
}

func TestMarafonaNotAwardedIfNotFirstPlayer(t *testing.T) {
	match := mkMatch("Alice", "Bob", "Carol", "Dave")
	hand := make([]model.Card, model.CardsPerPlayer)
	hand[0] = model.Card{Suit: model.Cups, Rank: model.Ace}
	hand[1] = model.Card{Suit: model.Cups, Rank: model.Two}
	hand[2] = model.Card{Suit: model.Cups, Rank: model.Three}
	for i := 3; i < model.CardsPerPlayer; i++ {
		hand[i] = model.Card{Suit: model.Swords, Rank: model.Four}
	}
	setHands(&match, [][]model.Card{{}, hand})
	match.TrumpSuit = model.Cups
	match.Table = []model.PlayedCard{{PlayerName: "Alice", Card: model.Card{Suit: model.Cups, Rank: model.Four}}}
	updated, err := PlayCard(match, "Bob", hand[0])
	assert.NoError(t, err)
	teamId := getPlayerTeamId(match, "Bob")
	assert.Equal(t, model.Point(0), updated.MatchPoints[teamId])
}

func TestMarafonaNotAwardedIfFirstCardNotAce(t *testing.T) {
	match := mkMatch("Alice", "Bob", "Carol", "Dave")
	hand := make([]model.Card, model.CardsPerPlayer)
	hand[0] = model.Card{Suit: model.Cups, Rank: model.Two}
	hand[1] = model.Card{Suit: model.Cups, Rank: model.Ace}
	hand[2] = model.Card{Suit: model.Cups, Rank: model.Three}
	for i := 3; i < model.CardsPerPlayer; i++ {
		hand[i] = model.Card{Suit: model.Swords, Rank: model.Four}
	}
	setHands(&match, [][]model.Card{hand})
	match.TrumpSuit = model.Cups
	updated, err := PlayCard(match, "Alice", hand[0])
	assert.NoError(t, err)
	teamId := getPlayerTeamId(match, "Alice")
	assert.Equal(t, model.Point(0), updated.MatchPoints[teamId])
}

func mkPlayers(names ...string) []model.Player {
	players := make([]model.Player, len(names))
	for i, n := range names {
		players[i] = model.Player{Name: n, TeamId: i % 2}
	}
	return players
}

func mkMatch(names ...string) model.Game {
	players := mkPlayers(names...)
	return model.Game{Players: players, FirstPlayer: names[0]}
}

func setHands(m *model.Game, hands [][]model.Card) {
	for i := range hands {
		if i < len(m.Players) {
			m.Players[i].Hand = model.Hand(hands[i])
		}
	}
}
