package matchmaking

import (
	"MarafoNet/internal/service"
	"context"
	"encoding/json"
)

type MatchmakingHub struct {
	ctx         context.Context
	cancel      context.CancelFunc
	etcdService *service.EtcdService
}

type handler func(ctx context.Context, id string) (<-chan []byte, context.CancelFunc)

// Returns a new Matchmaking hub
func NewMatchmakingHub(etcdService *service.EtcdService) *MatchmakingHub {
	ctx, cancel := context.WithCancel(context.Background())
	return &MatchmakingHub{
		cancel:      cancel,
		ctx:         ctx,
		etcdService: etcdService,
	}
}

func (hub *MatchmakingHub) GetStorageService() *service.EtcdService {
	return hub.etcdService
}

// Starts the matchmaking service as a Go Routine.
// It will start to check for players in queue and create games.
// Can be stopped by calling StopMatchmaking().
func (hub *MatchmakingHub) StartMatchmaking() {

}

// Stops the matchmaking service.
func (hub *MatchmakingHub) StopMatchmaking() {
	hub.cancel()
}

// Sets a watcher on the requested game, sending the information down the write channel.
func (hub *MatchmakingHub) SetGameWatcher(ctx context.Context, matchId string, writeChannel chan json.RawMessage) context.CancelFunc {
	return startWatcher(ctx, hub.GetStorageService().WatchMatch, matchId, writeChannel)
}

// Adds the player to the matchmaking queue, once a game is found, the write channel will be used to create
// a watcher for the game.
func (hub *MatchmakingHub) JoinQueue(ctx context.Context, playerName string, writeChannel chan json.RawMessage) context.CancelFunc {
	//startWatcher(ctx, hub.GetStorageService().method, playerName, writeChannel)
	return nil
}

func startWatcher(ctx context.Context, fun handler, arg string, writeChannel chan json.RawMessage) context.CancelFunc {
	watchChannel, cancelFunc := fun(ctx, arg)
	go func() {
		for json := range watchChannel {
			writeChannel <- json
		}
	}()
	return cancelFunc
}

func matchmake() {

}
