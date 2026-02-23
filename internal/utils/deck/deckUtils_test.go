package service

import (
	"MarafoNet/internal/model"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintCard(t *testing.T) {
	card := model.Card{Suit: model.Bastoni, Rank: model.Asso}
	expected := "Asso di Bastoni"
	actual := card.String()
	assert.Equal(t, expected, actual, "Expected %s, got %s", expected, actual)
}

func TestDeckLenght(t *testing.T) {
	deck := NewShuffledDeck()
	expectedCardCount := int(model.EndRank) * int(model.EndSuit)
	assert.Equal(t, expectedCardCount, len(deck), "Expected %d cards, got %d", expectedCardCount, len(deck))
}

func TestDeckCorrectness(t *testing.T) {
	deck := NewShuffledDeck()
	expected := newSortedDeck()
	sortedDeck := defaultSort(deck)
	for i := range expected {
		assert.Equal(t, expected[i], sortedDeck[i], "Expected card %v, got %v", expected[i], sortedDeck[i])
	}
}

// Might fail. To be removed after testing.
func TestDeckShuffling(t *testing.T) {
	iterations := 100
	for range iterations {
		firstDeck := NewShuffledDeck()
		secondDeck := NewShuffledDeck()
		assert.NotEqual(t, firstDeck, secondDeck, "Expected shuffled decks to be different")
	}
}

func TestDrawCards(t *testing.T) {
	numberOfCardsToDraw := 10
	deck := NewShuffledDeck()
	expectedCardsInHand := numberOfCardsToDraw
	expectedDeckDimension := len(deck) - expectedCardsInHand
	hand, remainingDeck := DrawCards(deck, numberOfCardsToDraw)
	actualCardsInHand := len(hand)
	actualDeckDimension := len(remainingDeck)
	assert.Equal(t, expectedCardsInHand, actualCardsInHand, "Expected hand to have %d cards, got %d", expectedCardsInHand, actualCardsInHand)
	assert.Equal(t, expectedDeckDimension, actualDeckDimension, "Expected remaining deck to have %d cards, got %d", expectedDeckDimension, actualDeckDimension)
}

func defaultSort(deck []model.Card) []model.Card {
	orderedDeck := make([]model.Card, len(deck))
	copy(orderedDeck, deck)
	sort.Slice(orderedDeck, less(orderedDeck))
	return orderedDeck
}

func less(deck []model.Card) func(i, j int) bool {
	return func(i, j int) bool {
		return absRank(deck[i]) < absRank(deck[j])
	}
}

func absRank(c model.Card) int {
	return int(c.Suit)*int(model.EndRank) + int(c.Rank)
}
