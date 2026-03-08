package service

import (
	"MarafoNet/internal/model"
	"fmt"
	"math/rand"
)

func NewShuffledDeck() model.Deck {
	deck := newSortedDeck()
	deck = shuffleDeck(deck)
	return deck
}

func DrawCards(deck model.Deck, n int) (model.Hand, model.Deck) {
	var hand model.Hand
	for range n {
		var card model.Card
		card, deck = drawCard(deck)
		hand = append(hand, card)
	}
	return hand, deck
}

func shuffleDeck(deck model.Deck) model.Deck {
	for i := range deck {
		j := rand.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}
	return deck
}

func newSortedDeck() model.Deck {
	var deck model.Deck
	for suit := model.StartSuit; suit <= model.EndSuit; suit++ {
		for rank := model.StartRank; rank <= model.EndRank; rank++ {
			deck = append(deck, model.Card{Suit: suit, Rank: rank})
		}
	}
	return deck
}

func printDeck(deck model.Deck) {
	for _, card := range deck {
		fmt.Printf("%s\n", card.String())
	}
}

func drawCard(deck model.Deck) (model.Card, model.Deck) {
	card := deck[0]
	deck = deck[1:]
	return card, deck
}
