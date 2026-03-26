package websockethub

import (
	"MarafoNet/internal/matchmaking"
	"MarafoNet/internal/service"
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	Connection         *websocket.Conn
	WriteChannel       chan json.RawMessage
	StorageService     *service.EtcdService
	GameService        *service.GameService
	MatchmakingService *matchmaking.MatchmakingHub
	playerName         string
	once               sync.Once
}

func CreateWebSocketHub(
	Conn *websocket.Conn,
	GameService *service.GameService,
	StorageService *service.EtcdService,
	MatchmakingService *matchmaking.MatchmakingHub,
) *WebSocketHub {
	var hub WebSocketHub
	hub.Connection = Conn
	hub.WriteChannel = make(chan json.RawMessage)
	hub.GameService = GameService
	hub.StorageService = StorageService
	hub.MatchmakingService = MatchmakingService
	return &hub
}

func (hub *WebSocketHub) GetPlayerName() string {
	return hub.playerName
}

func (hub *WebSocketHub) SetPlayerID(playerName string) {
	hub.once.Do(func() {
		hub.playerName = playerName
	})
}
