package matchmaking

import (
	"MarafoNet/internal/service"
	"context"
	"encoding/json"
	"log"
	"sync"
)

type MatchmakingHub struct {
	ctx          context.Context
	cancel       context.CancelFunc
	etcdService  *service.EtcdService
	gameService  *service.GameService
	queueWatcher context.CancelFunc
}

type GameUpdateMessage struct {
	Type string          `json:"type"`
	Game json.RawMessage `json:"game"`
}

// Callback function type for when game ID is assigned
type OnGameIDCallback func(gameID string)

type handler func(ctx context.Context, id string) (<-chan []byte, context.CancelFunc)

// Returns a new Matchmaking hub
func NewMatchmakingHub(etcdService *service.EtcdService, gameService *service.GameService) *MatchmakingHub {
	ctx, cancel := context.WithCancel(context.Background())
	log.Printf("started Matchmaking service")
	return &MatchmakingHub{
		cancel:      cancel,
		ctx:         ctx,
		etcdService: etcdService,
		gameService: gameService,
	}
}

func (hub *MatchmakingHub) GetStorageService() *service.EtcdService {
	return hub.etcdService
}

func (hub *MatchmakingHub) GetGameService() *service.GameService {
	return hub.gameService
}

// Starts the matchmaking service as a Go Routine.
// It will start to check for players in queue and create games.
// Can be stopped by calling StopMatchmaking().
func (hub *MatchmakingHub) StartMatchmaking(wg *sync.WaitGroup) {
	queueChannel, cancelQueueWatcher := hub.GetStorageService().WatchUserQueue(context.Background())
	usersQueue, err := hub.GetStorageService().GetUserQueue(context.Background())
	if err != nil {
		log.Printf("- matchmaking: Error getting initial user queue: %v", err)
		return
	}
	hub.checkQueueAndStartGame(usersQueue)
	go func() {
		for users := range queueChannel {
			log.Printf("- matchmaking: Current users in queue: %v", users)
			select {
			case <-hub.ctx.Done():
				cancelQueueWatcher()
				wg.Done()
				log.Printf("- matchmaking: Stopping matchmaking service")
				return
			default:
				hub.checkQueueAndStartGame(users)
			}
		}
	}()
}

func (hub *MatchmakingHub) checkQueueAndStartGame(users []string) {
	if len(users) >= 4 {
		log.Printf("- matchmaking: Found 4 players in queue, starting a game with players: %v", users[:4])
		gameID, _ := hub.GetGameService().StartGame(context.Background(), users[:4])
		for _, user := range users[:4] {
			hub.GetStorageService().RemoveUserFromQueue(context.Background(), user)
			hub.GetStorageService().SetUserCurrentGameId(context.Background(), user, gameID)
		}
		usersStillInQueue, _ := hub.GetStorageService().GetUserQueue(context.Background())
		log.Printf("- matchmaking: Current users in queue after matchmaking: %v", usersStillInQueue)
		for _, user := range users[:4] {
			gameID, _ := hub.GetStorageService().GetUserCurrentGameId(context.Background(), user)
			log.Printf("- matchmaking: user %s joined game %s", user, gameID)
		}
	}
}

// Stops the matchmaking service.
func (hub *MatchmakingHub) StopMatchmaking() {
	hub.cancel()
}

func (hub *MatchmakingHub) StartTimeoutWatcher(wg *sync.WaitGroup) {
	timeoutWatcher, cancelTimeoutWatcher := hub.GetStorageService().WatchUserTimeoutLease(context.Background())
	go func() {
		log.Printf("- timeout watcher: started watching timeout lease")
		for {
			select {
			case <-hub.ctx.Done():
				cancelTimeoutWatcher()
				wg.Done()
				log.Printf("- timeout watcher: Stopping timeout watcher service")
				return
			default:
				timeoutEvent, ok := <-timeoutWatcher
				if !ok {
					log.Printf("- timeout watcher: timeout channel closed")
					return
				}
				log.Printf("- timeout watcher: received timeout event for user %s in game %s", timeoutEvent.Username, timeoutEvent.GameID)
				err := hub.GetGameService().ForfeitGame(context.Background(), timeoutEvent.GameID, timeoutEvent.Username)
				if err != nil {
					log.Printf("- timeout watcher: error forfeiting game %s for user %s: %v", timeoutEvent.GameID, timeoutEvent.Username, err)
				} else {
					log.Printf("- timeout watcher: successfully forfeited game %s for user %s", timeoutEvent.GameID, timeoutEvent.Username)
				}
			}
		}
	}()
}

func (hub *MatchmakingHub) StopTimeoutWatcher() {
	hub.cancel()
}

// Sets a watcher on the requested game, sending the information down the write channel..
func (hub *MatchmakingHub) SetGameWatcher(ctx context.Context, gameId string, playerId string, cleanUpFunc func(), writeChannel chan json.RawMessage) *context.CancelFunc {
	watchChannel, cancelFunc := hub.GetStorageService().WatchGame(ctx, gameId)
	gameJSON, _, _ := hub.GetStorageService().GetGameJsonAndRevision(ctx, gameId)
	gameViewJson, err := hub.gameService.GetGameView(gameJSON, playerId)
	if err != nil {
		log.Printf("- game watcher: error getting game view for player %s in game %s: %v", playerId, gameId, err)
	}
	go func() {
		log.Printf("- game watcher: setting game watcher for game ID: %s", gameId)
		for {
			gameUpdateJson, ok := <-watchChannel
			if !ok {
				log.Printf("- game watcher: watch channel closed for game ID: %s", gameId)
				return
			}
			gameViewJson, err := hub.gameService.GetGameView(gameUpdateJson, playerId)
			if err != nil {
				log.Printf("- game watcher: error getting game view for player %s in game %s: %v", playerId, gameId, err)
				continue
			}
			sendGameView(gameViewJson, writeChannel)
			gameOver, err := hub.GetGameService().IsGameEnded(gameUpdateJson)
			if err != nil {
				log.Printf("- game watcher: error checking if game is over for game ID: %s, error: %v", gameId, err)
				continue
			}
			if gameOver {
				log.Printf("- game watcher: game over for game ID: %s, stopping watcher", gameId)
				hub.GetStorageService().RemoveUserCurrentGameId(ctx, playerId)
				cleanUpFunc()
				cancelFunc()
				return
			}
		}
	}()
	sendGameView(gameViewJson, writeChannel)
	return &cancelFunc
}

func sendGameView(gameViewJson []byte, writeChannel chan json.RawMessage) {
	message := GameUpdateMessage{
		Type: "game_update",
		Game: gameViewJson,
	}
	payload, _ := json.Marshal(message)
	writeChannel <- payload
}

// Adds the player to the matchmaking queue, once a game is found, the write channel will be used to create
// a watcher for the game. The onGameID callback will be called with the gameID once assigned.
func (hub *MatchmakingHub) JoinQueue(ctx context.Context, playerName string, writeChannel chan json.RawMessage, cleanUpFunc func(), onGameID OnGameIDCallback) *context.CancelFunc {
	hub.GetStorageService().PutUserIntoQueue(context.Background(), playerName)
	lobbyChannel, watcherCancelFunc := hub.GetStorageService().WatchUserLobby(ctx, playerName)
	cancelFunc := watcherCancelFunc
	go func() {
		log.Printf("- lobby watcher: started watching lobby for player %s", playerName)
		for {
			lobbyUpdate, ok := <-lobbyChannel
			if !ok {
				log.Printf("- lobby watcher: lobby channel closed for player %s", playerName)
				return
			}
			cancelFunc = *hub.SetGameWatcher(ctx, string(lobbyUpdate), playerName, cleanUpFunc, writeChannel)
			if onGameID != nil {
				onGameID(string(lobbyUpdate))
				log.Printf("- lobby watcher: lobby update for player %s: %s, starting watcher and returning", playerName, string(lobbyUpdate))
				return
			}
		}
	}()
	return &cancelFunc
}

func newPlayerList(players []string) json.RawMessage {
	var playerList []map[string]string
	for _, player := range players {
		playerMap := map[string]string{"Name": player}
		playerList = append(playerList, playerMap)
	}
	result, _ := json.Marshal(playerList)
	return result
}
