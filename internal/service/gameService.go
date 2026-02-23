package service

import (
	"MarafoNet/internal/model"
	gameLogic "MarafoNet/internal/utils/gameLogic"
	"encoding/json"
)

/*bastano i 4 nomi dei player*/
func SetUpGame(playerInfo json.RawMessage) json.RawMessage {
	players := decodePlayersFromJson(playerInfo)
	match := gameLogic.InitializeGame(players)
	return encodeMatchAsJSON(match)
}

func StartGame(gameInfo json.RawMessage) json.RawMessage {
	match := decodeMatchFromJson(gameInfo)
	match = gameLogic.StartGame(match)
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
