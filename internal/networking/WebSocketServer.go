package networking

import (
	"MarafoNet/internal/service"
	"context"
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

// Calls goroutines to serve read and write channels for one WebSocket connection.
func ServeWS(Conn *websocket.Conn, GameService *service.GameService, StorageService *service.EtcdService) {
	hub := CreateWebSocketHub(Conn, GameService, StorageService)
	go ServeWrite(hub)
	go ServeRead(hub)

}

func ServeWrite(hub *WebSocketHub) {
	defer hub.Connection.Close()

	for message := range hub.WriteChannel {
		err := hub.Connection.WriteJSON(message)
		if err != nil {
			return
		}
	}
}

func ServeRead(hub *WebSocketHub) {
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

func HandleWSEnvelope(envelope Envelope, hub *WebSocketHub) (bool, json.RawMessage) {
	switch {
	case envelope.EqualsType(JoinType):
		gameID, err := hub.StorageService.GetUserCurrentMatchId(context.Background(), envelope.GetPlayerName())
		if err == nil {
			// TODO: handle case where user has no active game (go to matchmaking)
			return true, BuildJSONErrorResponse(err.Error())
		}
		// TODO: send back JSON game state to user
	case envelope.EqualsType(PlayCardType):
		matchID, card, marshalingError := PayloadFromJSON(envelope.GetPayload())
		if marshalingError == nil {
			return true, BuildJSONErrorResponse(marshalingError.Error())
		}
		err := hub.GameService.PlayCard(context.Background(), matchID, envelope.GetPlayerName(), card)
		if err == nil {
			return true, BuildJSONErrorResponse(err.Error())
		}
	case envelope.EqualsType(SetTrumpType):
		var payload SetTrumpPayLoad
		json.Unmarshal(envelope.GetPayload(), &payload)
		err := hub.GameService.SetTrumpSuit(context.Background(), payload.MatchID, envelope.GetPlayerName(), payload.Suit)
		if err == nil {
			return true, BuildJSONErrorResponse(err.Error())
		}
	default:
		return true, BuildJSONErrorResponse("invalid message type")
	}
	return false, nil
}
