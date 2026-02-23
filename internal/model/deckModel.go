//go:generate stringer -type=Suit,Rank

package model

import (
	"strings"
)

const (
	Bastoni Suit = iota + 1
	Coppe
	Denari
	Spade
)

const (
	Asso Rank = iota + 1
	Due
	Tre
	Quattro
	Cinque
	Sei
	Sette
	Fante
	Cavallo
	Re
)

const (
	StartSuit = Bastoni
	EndSuit   = Spade
)

const (
	StartRank = Asso
	EndRank   = Re
)

type Card struct {
	Suit
	Rank
}

type Suit uint8
type Rank uint8

type Hand []Card

type Deck []Card

func (card Card) String() string {
	return card.Rank.String() + " di " + card.Suit.String()
}

func (hand Hand) String() string {
	strs := make([]string, len(hand))
	for i, card := range hand {
		strs[i] = card.String()
	}
	return strings.Join(strs, ", ")
}
