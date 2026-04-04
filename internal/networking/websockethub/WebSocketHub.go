package websockethub

import (
	"MarafoNet/internal/matchmaking"
	"MarafoNet/internal/service"
	"context"
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
	playerNameOnce     sync.Once
	cancelFunc         context.CancelFunc
	closeOnce          sync.Once
	isAuthenticated    bool
	matchID            string
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
	hub.closeOnce = sync.Once{}
	hub.isAuthenticated = false
	hub.matchID = ""
	return &hub
}

func (hub *WebSocketHub) GetPlayerName() string {
	return hub.playerName
}

func (hub *WebSocketHub) SetPlayerName(playerName string) {
	hub.playerNameOnce.Do(func() {
		hub.playerName = playerName
	})
}

func (hub *WebSocketHub) SetAuthenticated() {
	hub.isAuthenticated = true
}

func (hub *WebSocketHub) IsAuthenticated() bool {
	return hub.isAuthenticated
}

func (hub *WebSocketHub) SetMatchID(matchID string) {
	hub.matchID = matchID
}

func (hub *WebSocketHub) GetMatchID() string {
	return hub.matchID
}

// Sets the cancel function for the.current watch associated with the WebSocketHub (that is associated with the connection).
// Overwrites the previous cancel function if it exists.
func (hub *WebSocketHub) SetWatcherCancelFunc(cancelFunc context.CancelFunc) {
	hub.cancelFunc = cancelFunc
}

func (hub *WebSocketHub) CancelWatcher() {
	if hub.cancelFunc != nil {
		hub.cancelFunc()
		hub.cancelFunc = nil
	}
}

func (hub *WebSocketHub) Cleanup() {
	hub.closeOnce.Do(func() {
		closeConnection(hub)
	})
}

// Closes the connection and everything related to it: watchers, channels, etc...
func closeConnection(hub *WebSocketHub) {
	hub.Connection.Close()
	hub.CancelWatcher()
	hub.StorageService.RemoveUserFromQueue(context.Background(), hub.playerName)
}
