package websockethub

import (
	"MarafoNet/internal/matchmaking"
	"MarafoNet/internal/repository"
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	Connection             *websocket.Conn
	WriteChannel           chan json.RawMessage
	WebSocketRepository    repository.WebSocketRepository
	GameService            repository.GameServicer
	MatchmakingService     *matchmaking.MatchmakingHub
	playerName             string
	playerNameOnce         sync.Once
	lobbyWatcherCancelFunc *context.CancelFunc
	closeOnce              sync.Once
	isAuthenticated        bool
	gameId                 string
}

type cancelFunc func()

func CreateWebSocketHub(
	Conn *websocket.Conn,
	GameService repository.GameServicer,
	WebSocketDeps repository.WebSocketRepository,
	MatchmakingService *matchmaking.MatchmakingHub,
) *WebSocketHub {
	var hub WebSocketHub
	hub.Connection = Conn
	hub.WriteChannel = make(chan json.RawMessage, 10)
	hub.GameService = GameService
	hub.WebSocketRepository = WebSocketDeps
	hub.MatchmakingService = MatchmakingService
	hub.closeOnce = sync.Once{}
	hub.isAuthenticated = false
	hub.gameId = ""
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

func (hub *WebSocketHub) SetGameID(gameID string) {
	hub.gameId = gameID
}

func (hub *WebSocketHub) GetGameID() string {
	return hub.gameId
}

// Sets the cancel function for the.current watch associated with the WebSocketHub (that is associated with the connection).
// Overwrites the previous cancel function if it exists.
func (hub *WebSocketHub) SetWatcherCancelFunc(cancelFunc *context.CancelFunc) {
	hub.lobbyWatcherCancelFunc = cancelFunc
}

func (hub *WebSocketHub) IsWatcherCancelFuncSet() bool {
	return hub.lobbyWatcherCancelFunc != nil
}

func (hub *WebSocketHub) CancelWatcher() {
	if hub.lobbyWatcherCancelFunc != nil {
		(*hub.lobbyWatcherCancelFunc)()
		log.Printf("- lobby watcher: cancelled lobby watcher for player %s", hub.GetPlayerName())
		hub.lobbyWatcherCancelFunc = nil
	}
}

func (hub *WebSocketHub) Cleanup() {
	hub.closeOnce.Do(func() {
		closeConnection(hub)
		log.Printf("connection cleaned up correctly for client %s", hub.Connection.RemoteAddr())
	})
}

// Closes the connection and everything related to it: watchers, channels, etc...
func closeConnection(hub *WebSocketHub) {
	close(hub.WriteChannel)
	hub.Connection.Close()
	hub.CancelWatcher()
	hub.WebSocketRepository.RemoveUserFromQueue(context.Background(), hub.playerName)
	hub.WebSocketRepository.OnUserDisconnect(context.Background(), hub.playerName)
}
