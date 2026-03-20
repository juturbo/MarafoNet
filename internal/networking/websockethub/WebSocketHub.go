package websockethub

import (
	"MarafoNet/internal/service"
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	Connection     *websocket.Conn
	WriteChannel   chan json.RawMessage
	StorageService *service.EtcdService
	GameService    *service.GameService
	playerName     string
	once           sync.Once
}

func CreateWebSocketHub(Conn *websocket.Conn, GameService *service.GameService, StorageService *service.EtcdService) *WebSocketHub {
	var hub WebSocketHub
	hub.Connection = Conn
	hub.WriteChannel = make(chan json.RawMessage)
	hub.GameService = GameService
	hub.StorageService = StorageService
	return &hub
}

func (hub *WebSocketHub) GetPlayerName() string {
	return hub.playerName
}

func (hub *WebSocketHub) setPlayerID(playerName string) {
	hub.once.Do(func() {
		hub.playerName = playerName
	})
}
