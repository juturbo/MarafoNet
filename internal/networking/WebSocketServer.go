package networking

import (
	"MarafoNet/internal/matchmaking"
	"MarafoNet/internal/networking/websockethub"
	"MarafoNet/internal/service"
	"context"
	"encoding/json"

	"github.com/gorilla/websocket"
)

// Calls goroutines to serve read and write channels for one WebSocket connection.
func ServeWS(Conn *websocket.Conn, GameService *service.GameService, StorageService *service.EtcdService) {
	hub := websockethub.CreateWebSocketHub(Conn, GameService, StorageService)
	go ServeWrite(hub)
	go ServeRead(hub)

}

func ServeWrite(hub *websockethub.WebSocketHub) {
	defer hub.Connection.Close()

	for message := range hub.WriteChannel {
		err := hub.Connection.WriteJSON(message)
		if err != nil {
			return
		}
	}
}

func ServeRead(hub *websockethub.WebSocketHub) {
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

func HandleWSEnvelope(envelope Envelope, hub *websockethub.WebSocketHub) (bool, json.RawMessage) {
	// TODO: check if the username is already in the server and if the user is connected from another connection (stored in etcd?)
	if hub.GetPlayerName() != envelope.GetPlayerName() {
		return true, BuildJSONErrorResponse("player name does not match existing player name for this connection")
	} else if hub.GetPlayerName() == "" {
		hub.SetPlayerID(envelope.GetPlayerName())
	}
	switch {
	case envelope.EqualsType(JoinType):
		gameID, err := hub.StorageService.GetUserCurrentMatchId(context.Background(), envelope.GetPlayerName())
		if err != nil {
			// TODO: handle case where user has no active game (go to matchmaking)
			return true, BuildJSONErrorResponse(err.Error())
		} else {
			matchmaking.SetGameWatcher(context.Background(), gameID, hub.WriteChannel)
		}
	case envelope.EqualsType(PlayCardType):
		matchID, card, marshalingError := PayloadFromJSON(envelope.GetPayload())
		if marshalingError != nil {
			return true, BuildJSONErrorResponse(marshalingError.Error())
		}
		err := hub.GameService.PlayCard(context.Background(), matchID, envelope.GetPlayerName(), card)
		if err != nil {
			return true, BuildJSONErrorResponse(err.Error())
		}
	case envelope.EqualsType(SetTrumpType):
		var payload SetTrumpPayLoad
		json.Unmarshal(envelope.GetPayload(), &payload)
		err := hub.GameService.SetTrumpSuit(context.Background(), payload.MatchID, envelope.GetPlayerName(), payload.Suit)
		if err != nil {
			return true, BuildJSONErrorResponse(err.Error())
		}
	default:
		return true, BuildJSONErrorResponse("invalid message type")
	}
	return false, nil
}
