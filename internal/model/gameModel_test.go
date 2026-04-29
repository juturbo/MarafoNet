package model

import "testing"

func TestGameViewForPlayer(t *testing.T) {
	winnerTeam := 1
	game := Game{
		Players: []Player{
			{Name: "alice", TeamId: 0, Hand: Hand{{Suit: Clubs, Rank: Ace}}},
			{Name: "bob", TeamId: 1, Hand: Hand{{Suit: Cups, Rank: King}}},
		},
		Table:         []PlayedCard{{PlayerName: "alice", Card: Card{Suit: Clubs, Rank: Ace}}},
		MatchPoints:   [NUMBER_OF_TEAMS]Point{10, 11},
		TotalPoints:   [NUMBER_OF_TEAMS]int{20, 21},
		FirstPlayer:   "alice",
		CurrentPlayer: "bob",
		TrumpSuit:     Cups,
		WinnerTeam:    &winnerTeam,
		WinnerPlayers: []string{"bob"},
	}

	view, err := game.ViewForPlayer("bob")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if view.Player.Name != "bob" {
		t.Fatalf("expected player bob, got %q", view.Player.Name)
	}
	if len(view.Player.Hand) != 1 || view.Player.Hand[0].Rank != King {
		t.Fatalf("expected bob hand to be preserved, got %+v", view.Player.Hand)
	}
	if len(view.Table) != 1 || view.Table[0].PlayerName != "alice" {
		t.Fatalf("expected table to be copied, got %+v", view.Table)
	}
	if view.CurrentPlayer != "bob" || view.FirstPlayer != "alice" {
		t.Fatalf("expected turn metadata to be copied, got %+v", view)
	}
	if view.WinnerTeam == nil || *view.WinnerTeam != 1 {
		t.Fatalf("expected winner team to be copied, got %+v", view.WinnerTeam)
	}

	if _, err := game.ViewForPlayer("carol"); err == nil {
		t.Fatal("expected an error for an unknown player")
	}
}
