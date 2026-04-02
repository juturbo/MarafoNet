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
	if !hub.IsAuthenticated() {
		return checkAuthenticationMessage(hub, envelope)
	}
	switch {
	case envelope.EqualsType(JoinType):
		gameID, err := hub.StorageService.GetUserCurrentMatchId(context.Background(), hub.GetPlayerName())
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

// Checks the authentication message sent by the client and performs register or log-in operations accordingly.
// Registers the user both in the storaga service and in the connection's context.
func checkAuthenticationMessage(hub *websockethub.WebSocketHub, envelope Envelope) (bool, json.RawMessage) {
	replyMessageBuilder := NewReplyMessageBuilder()
	switch {
	case envelope.EqualsType(RegisterType) && isPlayerNew(hub, envelope):
		isAvailable, err := hub.StorageService.IsUsernameAvailable(context.Background(), envelope.GetUser().Name)
		if err == nil && isAvailable {
			err := hub.StorageService.RegisterUser(context.Background(), envelope.GetUser())
			if err != nil {
				return true, BuildJSONErrorResponse(err.Error())
			}
			hub.SetPlayerName(envelope.GetPlayerName())
			replyMessageBuilder.SetType("register_success")
			replyMessageBuilder.SetMessage(fmt.Sprintf("username %s successfully registered", envelope.GetPlayerName()))
		}
		return true, BuildJSONErrorResponse(err.Error())
	case envelope.EqualsType(LoginType):
		authenticated, err := checkPlayerIdentity(hub, envelope)
		if err == nil && authenticated {
			hub.SetAuthenticated()
			replyMessageBuilder.SetType("login_success")
			replyMessageBuilder.SetMessage(fmt.Sprintf("username %s successfully authenticated", envelope.GetPlayerName()))
		}
		replyMessageBuilder.SetMessage(fmt.Sprintf("authentication failed for username %s", envelope.GetPlayerName()))
		replyMessageBuilder.SetType("login_failed")
	default:
		replyMessageBuilder.SetType("invalid_request")
		replyMessageBuilder.SetMessage("An authentication error occoured")
	}
	return true, replyMessageBuilder.Build()
}

// Checks the player's identity against the name associated with the connection in WebSocketHub.
func checkPlayerIdentity(hub *websockethub.WebSocketHub, envelope Envelope) (bool, error) {
	verified, err := hub.StorageService.VerifyUser(context.Background(), envelope.GetUser())
	if err == nil && verified && hub.GetPlayerName() == envelope.GetPlayerName() {
		return true, nil
	} else {
		return false, fmt.Errorf("invalid player identity - %s", err.Error())
	}
}

// Check if the player can register in the connection with the provided username.
func isPlayerNew(hub *websockethub.WebSocketHub, envelope Envelope) bool {
	return hub.GetPlayerName() == "" && envelope.GetPlayerName() != "" && envelope.GetPassword() == ""
}
