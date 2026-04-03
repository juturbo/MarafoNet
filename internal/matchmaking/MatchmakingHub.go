package matchmaking

import (
	"MarafoNet/internal/service"
	"context"
	"encoding/json"
)

type MatchmakingHub struct {
	ctx          context.Context
	cancel       context.CancelFunc
	etcdService  *service.EtcdService
	gameService  *service.GameService
	queueWatcher context.CancelFunc
}

type MatchUpdateMessage struct {
	Type  string          `json:"type"`
	Match json.RawMessage `json:"match"`
}

type handler func(ctx context.Context, id string) (<-chan []byte, context.CancelFunc)

// Returns a new Matchmaking hub
func NewMatchmakingHub(etcdService *service.EtcdService, gameService *service.GameService) *MatchmakingHub {
	ctx, cancel := context.WithCancel(context.Background())
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
func (hub *MatchmakingHub) StartMatchmaking() {
	queueChannel, cancelQueueWatcher := hub.GetStorageService().WatchUserQueue(context.Background())
	for users := range queueChannel {
		select {
		case <-hub.ctx.Done():
			cancelQueueWatcher()
			return
		default:
			if len(users) >= 4 {
				hub.GetGameService().StartGame(context.Background(), users[:4])
				for _, user := range users[:4] {
					hub.GetStorageService().RemoveUserFromQueue(context.Background(), user)
				}
			}
		}
	}
}

// Stops the matchmaking service.
func (hub *MatchmakingHub) StopMatchmaking() {
	hub.cancel()
}

// Sets a watcher on the requested game, sending the information down the write channel.
func (hub *MatchmakingHub) SetGameWatcher(ctx context.Context, matchId string, writeChannel chan json.RawMessage) context.CancelFunc {
	return startWatcher(ctx, hub.GetStorageService().WatchGame, matchId, writeChannel)
}

// Adds the player to the matchmaking queue, once a game is found, the write channel will be used to create
// a watcher for the game.
func (hub *MatchmakingHub) JoinQueue(ctx context.Context, playerName string, writeChannel chan json.RawMessage) context.CancelFunc {
	return startWatcher(ctx, hub.GetStorageService().WatchUserLobby, playerName, writeChannel)
}

func startWatcher(ctx context.Context, fun handler, arg string, writeChannel chan json.RawMessage) context.CancelFunc {
	watchChannel, cancelFunc := fun(ctx, arg)
	go func() {
		for update := range watchChannel {
			message := MatchUpdateMessage{
				Type:  "match_update",
				Match: update,
			}
			payload, _ := json.Marshal(message)
			writeChannel <- payload
		}
	}()
	return cancelFunc
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
