package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuitConstants(t *testing.T) {
	assert.Equal(t, Suit(1), Clubs)
	assert.Equal(t, Suit(2), Cups)
	assert.Equal(t, Suit(3), Coins)
	assert.Equal(t, Suit(4), Swords)
}

func TestSuitRange(t *testing.T) {
	assert.Equal(t, Clubs, StartSuit)
	assert.Equal(t, Swords, EndSuit)
	assert.True(t, StartSuit <= Clubs && Clubs <= EndSuit)
	assert.True(t, StartSuit <= Swords && Swords <= EndSuit)
}

func TestRankConstants(t *testing.T) {
	assert.Equal(t, Rank(1), Ace)
	assert.Equal(t, Rank(2), Two)
	assert.Equal(t, Rank(3), Three)
	assert.Equal(t, Rank(4), Four)
	assert.Equal(t, Rank(5), Five)
	assert.Equal(t, Rank(6), Six)
	assert.Equal(t, Rank(7), Seven)
	assert.Equal(t, Rank(8), Jack)
	assert.Equal(t, Rank(9), Knight)
	assert.Equal(t, Rank(10), King)
}

func TestRankRange(t *testing.T) {
	assert.Equal(t, Ace, StartRank)
	assert.Equal(t, King, EndRank)
	assert.True(t, StartRank <= Ace && Ace <= EndRank)
	assert.True(t, StartRank <= King && King <= EndRank)
}

func TestAllMinorPointCards(t *testing.T) {
	minorCards := []Rank{Two, Three, Jack, Knight, King}
	for _, rank := range minorCards {
		card := Card{Rank: rank}
		assert.Equal(t, Point(MINOR_POINTS), card.PointValue())
	}
}

func TestAllBlankPointCards(t *testing.T) {
	blankCards := []Rank{Four, Five, Six, Seven}
	for _, rank := range blankCards {
		card := Card{Rank: rank}
		assert.Equal(t, Point(BLANK_POINTS), card.PointValue())
	}
}

func TestInvalidRankPower(t *testing.T) {
	invalidCard := Card{Rank: Rank(0)}
	assert.Equal(t, 0, invalidCard.Power())
}
