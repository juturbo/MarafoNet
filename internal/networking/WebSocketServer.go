package networking

import (
	"MarafoNet/internal/service"
	"context"
	"encoding/json"

	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	Connection     *websocket.Conn
	WriteChannel   chan json.RawMessage
	StorageService *service.EtcdService
	GameService    *service.GameService
}

func CreateWebSocketHub(Conn *websocket.Conn, GameService *service.GameService, StorageService *service.EtcdService) WebSocketHub {
	var hub WebSocketHub
	hub.Connection = Conn
	hub.WriteChannel = make(chan json.RawMessage)
	hub.GameService = GameService
	hub.StorageService = StorageService
	return hub
}

// Calls goroutines to serve read and write channels for one WebSocket connection.
func ServeWS(Conn *websocket.Conn, GameService *service.GameService, StorageService *service.EtcdService) {
	hub := CreateWebSocketHub(Conn, GameService, StorageService)
	go ServeWrite(hub)
	go ServeRead(hub)

}

func ServeWrite(hub WebSocketHub) {
	defer hub.Connection.Close()

	for message := range hub.WriteChannel {
		err := hub.Connection.WriteJSON(message)
		if err != nil {
			return
		}
	}
}

func ServeRead(hub WebSocketHub) {
	defer hub.Connection.Close()

	for {
		var envelope WSEnvelope
		err := hub.Connection.ReadJSON(&envelope)
		if err != nil {
			break
		}
		var response, payload = HandleWSEnvelope(envelope, hub)
		if response {
			hub.WriteChannel <- payload
		}
	}
}

func HandleWSEnvelope(envelope Envelope, hub WebSocketHub) (bool, json.RawMessage) {
	switch {
	case envelope.EqualsType(JoinType):
		var joinPayload JoinPayload
		json.Unmarshal(envelope.GetPayload(), &joinPayload)
		gameID, err := hub.StorageService.GetUserCurrentMatchId(context.Background(), joinPayload.PlayerName)
		if err != nil {
			// TODO: handle case where user has no active game (go to matchmaking)
			panic("unimplemented")

		}
		// TODO: send back JSON game state to user
	case envelope.EqualsType(PlayCardType):
	case envelope.EqualsType(SetTrumpType):
		panic("unimplemented")
	default:
		panic("unknown envelope type")
	}
}
