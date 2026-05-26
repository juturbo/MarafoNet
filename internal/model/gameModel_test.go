package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCard_Equal(t *testing.T) {
	card1 := Card{Suit: Clubs, Rank: Ace}
	card2 := Card{Suit: Clubs, Rank: Ace}
	card3 := Card{Suit: Cups, Rank: Ace}

	assert.True(t, card1.Equal(card2))
	assert.False(t, card1.Equal(card3))
}

func TestCard_PointValue(t *testing.T) {
	tests := []struct {
		card     Card
		expected Point
	}{
		{Card{Rank: Ace}, ACE_POINTS},
		{Card{Rank: Two}, MINOR_POINTS},
		{Card{Rank: Three}, MINOR_POINTS},
		{Card{Rank: Four}, BLANK_POINTS},
		{Card{Rank: Five}, BLANK_POINTS},
		{Card{Rank: Six}, BLANK_POINTS},
		{Card{Rank: Seven}, BLANK_POINTS},
		{Card{Rank: Jack}, MINOR_POINTS},
		{Card{Rank: Knight}, MINOR_POINTS},
		{Card{Rank: King}, MINOR_POINTS},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.card.PointValue())
	}
}

func TestCard_Power(t *testing.T) {
	tests := []struct {
		rank     Rank
		expected int
	}{
		{Three, 10},
		{Two, 9},
		{Ace, 8},
		{King, 7},
		{Knight, 6},
		{Jack, 5},
		{Seven, 4},
		{Six, 3},
		{Five, 2},
		{Four, 1},
	}

	for _, test := range tests {
		card := Card{Rank: test.rank}
		assert.Equal(t, test.expected, card.Power())
	}
}

func TestCard_IsHigherThan(t *testing.T) {
	tests := []struct {
		card1          Card
		card2          Card
		trumpSuit      Suit
		expectedResult bool
	}{
		{
			// Same suit, higher power
			Card{Suit: Clubs, Rank: Three},
			Card{Suit: Clubs, Rank: Two},
			Cups,
			true,
		},
		{
			// Same suit, lower power
			Card{Suit: Clubs, Rank: Two},
			Card{Suit: Clubs, Rank: Three},
			Cups,
			false,
		},
		{
			// Card1 is trump, Card2 is not
			Card{Suit: Clubs, Rank: Four},
			Card{Suit: Cups, Rank: Ace},
			Clubs,
			true,
		},
		{
			// Card2 is trump, Card1 is not
			Card{Suit: Cups, Rank: Ace},
			Card{Suit: Clubs, Rank: Four},
			Clubs,
			false,
		},
		{
			// Different suits, neither trump
			Card{Suit: Clubs, Rank: Ace},
			Card{Suit: Cups, Rank: Ace},
			Swords,
			false,
		},
		{
			// Both trump
			Card{Suit: Clubs, Rank: Three},
			Card{Suit: Clubs, Rank: Two},
			Clubs,
			true,
		},
	}

	for _, test := range tests {
		result := test.card1.IsHigherThan(test.card2, test.trumpSuit)
		assert.Equal(t, test.expectedResult, result)
	}
}

func TestGame_ViewForPlayer_Success(t *testing.T) {
	name := "Alice"
	teamId := 0
	game := Game{
		Players: []Player{
			{Name: name, TeamId: teamId, Hand: Hand{{Suit: Clubs, Rank: Ace}}},
			{Name: "Bob", TeamId: 1, Hand: Hand{{Suit: Cups, Rank: Two}}},
		},
		Table:         []PlayedCard{},
		TotalPoints:   [2]int{10, 20},
		FirstPlayer:   name,
		CurrentPlayer: "Bob",
		TrumpSuit:     Clubs,
	}

	gameView, err := game.ViewForPlayer(name)
	assert.NoError(t, err)
	assert.Equal(t, name, gameView.Players[0].Name)
	assert.Equal(t, len(game.Players[0].Hand), len(gameView.PlayerHand))
	assert.Equal(t, Ace, gameView.PlayerHand[0].Rank)
}

func TestGame_ViewForPlayer_PlayerNotFound(t *testing.T) {
	name := "Alice"
	teamId := 0
	game := Game{
		Players: []Player{
			{Name: name, TeamId: teamId},
		},
	}

	_, err := game.ViewForPlayer("NonExistent")
	assert.Error(t, err)
}

func TestGame_ViewForPlayer_CopiesPublicInfo(t *testing.T) {
	winnerTeam := 0
	name := "Alice"
	winners := []string{name}
	game := Game{
		Players: []Player{
			{Name: name, TeamId: 0, Hand: Hand{}},
		},
		Table:         []PlayedCard{},
		LastTrick:     []PlayedCard{},
		TotalPoints:   [2]int{41, 10},
		FirstPlayer:   name,
		CurrentPlayer: name,
		TrumpSuit:     Clubs,
		WinnerTeam:    &winnerTeam,
		WinnerPlayers: winners,
	}

	gameView, err := game.ViewForPlayer(name)
	assert.NoError(t, err)
	assert.Equal(t, game.TotalPoints, gameView.TotalPoints)
	assert.Equal(t, game.FirstPlayer, gameView.FirstPlayer)
	assert.Equal(t, game.CurrentPlayer, gameView.CurrentPlayer)
	assert.Equal(t, game.TrumpSuit, gameView.TrumpSuit)
	assert.Equal(t, winnerTeam, *gameView.WinnerTeam)
	assert.Equal(t, winners, gameView.WinnerPlayers)
}
