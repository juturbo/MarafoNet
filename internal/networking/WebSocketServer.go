package networking

import (
	"MarafoNet/internal/matchmaking"
	"MarafoNet/internal/networking/websockethub"
	"MarafoNet/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

var once sync.Once

// Calls goroutines to serve read and write channels for one WebSocket connection.
func ServeWS(
	Conn *websocket.Conn,
	GameService *service.GameService,
	StorageService *service.EtcdService,
	MatchmakingService *matchmaking.MatchmakingHub,
) {
	hub := websockethub.CreateWebSocketHub(Conn, GameService, StorageService, MatchmakingService)
	go ServeWrite(hub)
	go ServeRead(hub)

}

func ServeWrite(hub *websockethub.WebSocketHub) {
	defer once.Do(func() {
		closeConnection(hub)
	})

	for message := range hub.WriteChannel {
		err := hub.Connection.WriteJSON(message)
		if err != nil {
			return
		}
	}
}

func ServeRead(hub *websockethub.WebSocketHub) {
	defer once.Do(func() {
		closeConnection(hub)
	})

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
	if bool, err := checkPlayerIdentity(hub, envelope); bool {
		return true, BuildJSONErrorResponse(err.Error())
	}
	switch {
	case envelope.EqualsType(JoinType):
		gameID, err := hub.StorageService.GetUserCurrentMatchId(context.Background(), envelope.GetPlayerName())
		if err != nil {
			// TODO: manage errors from matchmaking calls
			hub.MatchmakingService.JoinQueue(context.Background(), hub.GetPlayerName(), hub.WriteChannel)
			return true, BuildJSONErrorResponse(err.Error())
		} else {
			_, err := hub.SetWatcherCancelFunc(
				hub.MatchmakingService.SetGameWatcher(context.Background(), gameID, hub.WriteChannel),
			)
			if err != nil {
				return true, BuildJSONErrorResponse(err.Error())
			}
		}
	case envelope.EqualsType(PlayCardType):
		matchID, card, marshalingError := PayloadFromJSON(envelope.GetPayload())
		if marshalingError != nil {
			return true, BuildJSONErrorResponse(marshalingError.Error())
		}
		err := hub.GameService.PlayCard(context.Background(), matchID, hub.GetPlayerName(), card)
		if err != nil {
			return true, BuildJSONErrorResponse(err.Error())
		}
	case envelope.EqualsType(SetTrumpType):
		var payload SetTrumpPayLoad
		json.Unmarshal(envelope.GetPayload(), &payload)
		err := hub.GameService.SetTrumpSuit(context.Background(), payload.MatchID, hub.GetPlayerName(), payload.Suit)
		if err != nil {
			return true, BuildJSONErrorResponse(err.Error())
		}
	default:
		return true, BuildJSONErrorResponse("invalid message type")
	}
	return false, nil
}

// Checks the player's identity against the name associated with the connection in WebSocketHub.
// If there's no name associated with the connection, then the one sent is set and a new UUIDv4 is generated
// and associated with the player's name.
func checkPlayerIdentity(hub *websockethub.WebSocketHub, envelope Envelope) (bool, error) {
	// 1. Check if the name in the message is the same as the one in the WebSocketHub.
	// Check also if the UUIDv4 passed is the same as the one in etcd.
	if hub.GetPlayerName() != envelope.GetPlayerName() {
		return true, fmt.Errorf("player identity does not match the existing one for this connection")
	} else if hub.GetPlayerName() == "" {
		// 2. If no name is associated with the connection, set it and generate a new UUIDv4 for the player.
		hub.SetPlayerID(envelope.GetPlayerName())
	}
	return false, nil
}

// Closes the connection and everything related to it: watchers, channels, etc...
func closeConnection(hub *websockethub.WebSocketHub) {
	hub.Connection.Close()
	hub.CancelWatcher()
}
