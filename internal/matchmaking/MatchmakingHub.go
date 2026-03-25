package matchmaking

import (
	"context"
	"encoding/json"
)

type MatchmakingHub struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// Returns a new Matchmaking hub
func NewMatchmakingHub() *MatchmakingHub {
	ctx, cancel := context.WithCancel(context.Background())
	return &MatchmakingHub{
		cancel: cancel,
		ctx:    ctx,
	}
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
func SetGameWatcher(ctx context.Context, matchId string, writeChannel chan json.RawMessage) error {
	return nil
}

// Adds the player to the matchmaking queue, once a game is found, the write channel will be used to create
// a watcher for the game.
func JoinQueue(ctx context.Context, playerName string, writeChannel chan json.RawMessage) error {
	return nil
}

func matchmake() {

}
