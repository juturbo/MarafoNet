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
	defer hub.Cleanup()

	for message := range hub.WriteChannel {
		err := hub.Connection.WriteJSON(message)
		if err != nil {
			return
		}
	}
}

func ServeRead(hub *websockethub.WebSocketHub) {
	defer hub.Cleanup()

	for {
		var envelope WSEnvelope
		err := hub.Connection.ReadJSON(&envelope)
		if err != nil {
			close(hub.WriteChannel)
			break
		}
		var response, payload = HandleWSEnvelope(envelope, hub)
		if response {
			hub.WriteChannel <- payload
		}
	}
}

func HandleWSEnvelope(envelope Envelope, hub *websockethub.WebSocketHub) (bool, json.RawMessage) {
	replyMessageBuilder := NewReplyMessageBuilder()
	authenticated, err := authenticatePlayer(hub, envelope)
	if authenticated {
		replyMessageBuilder.SetType("authentication_success")
	} else {
		replyMessageBuilder.SetType("authentication_failure")
		replyMessageBuilder.SetMessage(err.Error())
		return true, replyMessageBuilder.Build()
	}
	switch {
	case envelope.EqualsType(JoinType):
		gameID, err := hub.StorageService.GetUserCurrentMatchId(context.Background(), envelope.GetPlayerName())
		if err != nil {
			return true, BuildJSONErrorResponse(err.Error())
		}
		if gameID != "" {
			hub.SetWatcherCancelFunc(
				hub.MatchmakingService.JoinQueue(context.Background(), hub.GetPlayerName(), hub.WriteChannel),
			)
		} else {
			hub.SetWatcherCancelFunc(
				hub.MatchmakingService.SetGameWatcher(context.Background(), gameID, hub.WriteChannel),
			)
		}
		return true, replyMessageBuilder.Build()
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
	verified, _ := hub.StorageService.VerifyUser(context.Background(), envelope.GetUser())
	if verified && hub.GetPlayerName() == envelope.GetPlayerName() {
		return true, nil
	} else {
		return false, fmt.Errorf("invalid player identity")
	}
}

func isPlayerNew(hub *websockethub.WebSocketHub, envelope Envelope) bool {
	return hub.GetPlayerName() == "" && envelope.GetPlayerName() != "" && envelope.GetPassword() == ""
}

func authenticatePlayer(hub *websockethub.WebSocketHub, envelope Envelope) (bool, error) {
	isAvailable, err := hub.StorageService.IsUsernameAvailable(context.Background(), envelope.GetPlayerName())
	if err != nil {
		return false, err
	}
	if isPlayerNew(hub, envelope) && isAvailable {
		err := hub.StorageService.RegisterUser(context.Background(), envelope.GetUser())
		if err != nil {
			return false, err
		}
		hub.SetPlayerName(envelope.GetPlayerName())
		return true, nil
	} else {
		check, err := checkPlayerIdentity(hub, envelope)
		return check, err
	}
}
