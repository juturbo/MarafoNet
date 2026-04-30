package networking

import (
	"MarafoNet/internal/matchmaking"
	"MarafoNet/internal/networking/websockethub"
	"MarafoNet/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

var once sync.Once

// Calls goroutines to serve read and write channels for one WebSocket connection.
func ServeWS(
	Conn *websocket.Conn,
	GracefulShutdownContext context.Context,
	GameService *service.GameService,
	StorageService *service.EtcdService,
	MatchmakingService *matchmaking.MatchmakingHub,
) {
	hub := websockethub.CreateWebSocketHub(Conn, GameService, StorageService, MatchmakingService)
	log.Printf("New WebSocket connection established with client %s", hub.Connection.RemoteAddr())
	go ServeWrite(hub, GracefulShutdownContext)
	go ServeRead(hub, GracefulShutdownContext)

}

func ServeWrite(hub *websockethub.WebSocketHub, GracefulShutdownContext context.Context) {
	defer hub.Cleanup()

	for {
		select {
		case <-GracefulShutdownContext.Done():
			log.Printf("graceful shutdown signal received, closing write channel for client %s", hub.Connection.RemoteAddr())
			return
		case message := <-hub.WriteChannel:
			err := hub.Connection.WriteJSON(message)
			if err != nil {
				return
			}
		}
	}
}

func ServeRead(hub *websockethub.WebSocketHub, GracefulShutdownContext context.Context) {
	defer hub.Cleanup()

	for {
		select {
		case <-GracefulShutdownContext.Done():
			log.Printf("graceful shutdown signal received, closing read channel for client %s", hub.Connection.RemoteAddr())
			return
		default:
			var envelope WSEnvelope
			err := hub.Connection.ReadJSON(&envelope)
			if err != nil {
				log.Printf("error reading message from client %s: %v. Closing connection", hub.Connection.RemoteAddr(), err.Error())
				hub.Cleanup()
				return
			}
			var response, payload = HandleWSEnvelope(envelope, hub)
			if response {
				hub.WriteChannel <- payload
			}
		}
	}
}

func HandleWSEnvelope(envelope Envelope, hub *websockethub.WebSocketHub) (bool, json.RawMessage) {
	if !hub.IsAuthenticated() {
		return checkAuthenticationMessage(hub, envelope)
	}
	switch {
	case envelope.EqualsType(JoinType):
		log.Printf(" - wss: received join request from player %s", hub.GetPlayerName())
		gameID, err := hub.StorageService.GetUserCurrentGameId(context.Background(), hub.GetPlayerName())
		log.Printf(" - wss: player %s current game ID is %s", hub.GetPlayerName(), gameID)
		if err != nil {
			return true, BuildJSONErrorResponse(err.Error())
		}
		if gameID == "" && !hub.IsWatcherCancelFuncSet() {
			putUserInQueue(hub)
		} else if !hub.IsWatcherCancelFuncSet() {
			hub.SetWatcherCancelFunc(
				hub.MatchmakingService.SetGameWatcher(context.Background(), gameID, hub.GetPlayerName(), func() {
					hub.SetWatcherCancelFunc(nil)
				}, hub.WriteChannel),
			)
			hub.SetGameID(gameID)
		}
	case envelope.EqualsType(PlayCardType):
		card, marshalingError := PayloadFromJSON(envelope.GetPayload())
		if marshalingError != nil {
			return true, BuildJSONErrorResponse(marshalingError.Error())
		}
		err := hub.GameService.PlayCard(context.Background(), hub.GetGameID(), hub.GetPlayerName(), card)
		if err != nil {
			return true, BuildJSONErrorResponse(err.Error())
		}
	case envelope.EqualsType(SetTrumpType):
		var payload SetTrumpPayLoad
		json.Unmarshal(envelope.GetPayload(), &payload)
		err := hub.GameService.SetTrumpSuit(context.Background(), hub.GetGameID(), hub.GetPlayerName(), payload.Suit)
		if err != nil {
			return true, BuildJSONErrorResponse(err.Error())
		}
	case envelope.EqualsType(PlayAgainType) || envelope.EqualsType(QuitType):
		log.Printf(" - wss: received %s request from player %s", envelope.GetMessageType(), hub.GetPlayerName())
		gameID, err := hub.StorageService.GetUserCurrentGameId(context.Background(), hub.GetPlayerName())
		if err != nil {
			return true, BuildJSONErrorResponse(err.Error())
		}
		if gameID != "" && isGameOver(hub, gameID) {
			if hub.IsWatcherCancelFuncSet() {
				hub.CancelWatcher()
			}
			hub.StorageService.RemoveUserCurrentGameId(context.Background(), hub.GetPlayerName())
			if envelope.EqualsType(PlayAgainType) {
				putUserInQueue(hub)
			}
		} else {
			return true, BuildJSONErrorResponse("cannot play again until current game is over or if no gameId is set")
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
		err := hub.StorageService.RegisterUser(context.Background(), envelope.GetUser())
		if err != nil {
			replyMessageBuilder.SetType("register_failed")
			replyMessageBuilder.SetMessage(fmt.Sprintf("user registration failed for username %s. Error: %s", envelope.GetPlayerName(), err.Error()))
			return true, replyMessageBuilder.Build()
		}
		replyMessageBuilder.SetType("register_success")
		replyMessageBuilder.SetMessage(fmt.Sprintf("username %s successfully registered", envelope.GetPlayerName()))
	case envelope.EqualsType(LoginType):
		err := hub.StorageService.LoginUser(context.Background(), envelope.GetUser())
		if err != nil {
			replyMessageBuilder.SetType("login_failed")
			replyMessageBuilder.SetMessage(fmt.Sprintf("authentication failed for username %s. Error: %s", envelope.GetPlayerName(), err.Error()))
			return true, replyMessageBuilder.Build()
		}
		hub.SetAuthenticated()
		hub.SetPlayerName(envelope.GetPlayerName())
		replyMessageBuilder.SetType("login_success")
		replyMessageBuilder.SetMessage(fmt.Sprintf("username %s successfully authenticated", envelope.GetPlayerName()))
		return true, replyMessageBuilder.Build()
	default:
		replyMessageBuilder.SetType("invalid_request")
		replyMessageBuilder.SetMessage(fmt.Sprintf("invalid authentication message type %s", envelope.GetMessageType()))
	}
	return true, replyMessageBuilder.Build()
}

// Check if the player can register in the connection with the provided username.
func isPlayerNew(hub *websockethub.WebSocketHub, envelope Envelope) bool {
	return hub.GetPlayerName() == "" && envelope.GetPlayerName() != "" && envelope.GetPassword() != ""
}

func isGameOver(hub *websockethub.WebSocketHub, gameId string) bool {
	game, _, err := hub.StorageService.GetGameJsonAndRevision(context.Background(), gameId)
	if err != nil {
		log.Printf("error fetching game data for game ID %s: %v", gameId, err.Error())
		return false
	}
	result, err := hub.GameService.IsGameEnded(game)
	if err != nil {
		log.Printf("error checking if game is over for game ID %s: %v", gameId, err.Error())
		return false
	}
	return result
}

func putUserInQueue(hub *websockethub.WebSocketHub) {
	hub.SetWatcherCancelFunc(
		hub.MatchmakingService.JoinQueue(context.Background(), hub.GetPlayerName(), hub.WriteChannel, func() {
			hub.SetWatcherCancelFunc(nil)
		}, func(gameID string) {
			hub.SetGameID(gameID)
		}),
	)
}
