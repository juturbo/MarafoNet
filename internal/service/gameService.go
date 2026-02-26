package service

import (
	"MarafoNet/internal/model"
	gameLogic "MarafoNet/internal/utils/gameLogic"
	"encoding/json"
)

/*
Return solo problemi/cose da inviare solo al client che ha fatto la chiamata (tramite backend)
altrimenti se non dà errori chiamo metodo etcdService per aggiornare etcd con le nuove informazioni del match
*/
func StartGame(playerInfo json.RawMessage) json.RawMessage {
	players := decodePlayersFromJson(playerInfo)
	match := gameLogic.StartGame(players)
	return encodeMatchAsJSON(match)
}

func SetTrumpSuit(gameInfo json.RawMessage, playerName string, suit model.Suit) json.RawMessage {
	match := decodeMatchFromJson(gameInfo)
	match, err := gameLogic.SetTrumpSuit(match, playerName, suit)
	if err != nil {
		panic(err)
	}
	return encodeMatchAsJSON(match)
}

func PlayCard(gameInfo json.RawMessage, playerName string, card model.Card) json.RawMessage {
	match := decodeMatchFromJson(gameInfo)
	match, err := gameLogic.PlayCard(match, playerName, card)
	if err != nil {
		panic(err)
	}
	if match.WinnerPlayers != nil {
		//Showcase winner players
	}
	return encodeMatchAsJSON(match)
}

func decodePlayersFromJson(playerNames json.RawMessage) []model.Player {
	var players []model.Player
	err := json.Unmarshal(playerNames, &players)
	if err != nil {
		panic(err)
	}
	return players
}

func decodeMatchFromJson(gameInfo json.RawMessage) model.Match {
	var match model.Match
	err := json.Unmarshal(gameInfo, &match)
	if err != nil {
		panic(err)
	}
	return match
}

func encodeMatchAsJSON(match model.Match) json.RawMessage {
	matchBytes, err := json.Marshal(match)
	if err != nil {
		panic(err)
	}
	return matchBytes
}
