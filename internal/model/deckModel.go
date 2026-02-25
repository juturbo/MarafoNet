//go:generate stringer -type=Suit,Rank

package model

import (
	"strings"
)

const (
	AcePoints   = 3
	MinorPoints = 1
	BlankPoints = 0
)

const (
	Clubs Suit = iota + 1
	Cups
	Coins
	Swords
)

const (
	Ace Rank = iota + 1
	Two
	Three
	Four
	Five
	Six
	Seven
	Jack
	Knight
	King
)

const (
	StartSuit = Clubs
	EndSuit   = Swords
)

const (
	StartRank = Ace
	EndRank   = King
)

type Card struct {
	Suit Suit
	Rank Rank
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

func (card1 Card) Equal(card2 Card) bool {
	var isValid = false
	var cardMatchesRank = card1.Rank == card2.Rank
	var cardMatchesSuit = card1.Suit == card2.Suit
	if cardMatchesRank && cardMatchesSuit {
		isValid = true
		return isValid
	}
	isValid = false
	return isValid
}

func (card1 Card) IsHigherThan(card2 Card, trumpSuit Suit) bool {
	var card1IsTrump = card1.Suit == trumpSuit
	var card2IsTrump = card2.Suit == trumpSuit
	if card1IsTrump && !card2IsTrump {
		return true
	}
	if !card1IsTrump && card2IsTrump {
		return false
	}
	if card1.Suit == card2.Suit {
		return card1.Rank > card2.Rank
	}
	return false
}

func (card Card) PointValue() Point {
	switch card.Rank {
	case Ace:
		return AcePoints
	case Two, Three, Jack, Knight, King:
		return MinorPoints
	default:
		return BlankPoints
	}
}
