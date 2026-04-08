package matchmaking

import (
	"MarafoNet/internal/service"
	"context"
	"encoding/json"
	"log"
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
func (hub *MatchmakingHub) StartMatchmaking() {
	queueChannel, cancelQueueWatcher := hub.GetStorageService().WatchUserQueue(context.Background())
	go func() {
		for users := range queueChannel {
			log.Printf("Current users in queue: %v", users)
			select {
			case <-hub.ctx.Done():
				cancelQueueWatcher()
				return
			default:
				if len(users) >= 4 {
					log.Printf("Found 4 players in queue, starting a game with players: %v", users[:4])
					matchID, _ := hub.GetGameService().StartGame(context.Background(), users[:4])
					for _, user := range users[:4] {
						hub.GetStorageService().RemoveUserFromQueue(context.Background(), user)
						hub.GetStorageService().SetUserCurrentMatchId(context.Background(), user, matchID)
					}
				}
			}
		}
	}()
}

// Stops the matchmaking service.
func (hub *MatchmakingHub) StopMatchmaking() {
	hub.cancel()
}

// Sets a watcher on the requested game, sending the information down the write channel..
func (hub *MatchmakingHub) SetGameWatcher(ctx context.Context, matchId string, writeChannel chan json.RawMessage) context.CancelFunc {
	watchChannel, cancelFunc := hub.GetStorageService().WatchGame(ctx, matchId)
	matchJSON, _, _ := hub.GetStorageService().GetMatchJsonAndRevision(ctx, matchId)
	go func() {
		log.Printf("matchmaking: setting game watcher for match ID: %s", matchId)
		for {
			update, ok := <-watchChannel
			if !ok {
				log.Printf("matchmaking: watch channel closed for match ID: %s", matchId)
				return
			}
			sendMatchUpdate(update, writeChannel)
		}
	}()
	sendMatchUpdate(matchJSON, writeChannel)
	return cancelFunc
}

func sendMatchUpdate(update []byte, writeChannel chan json.RawMessage) {
	message := MatchUpdateMessage{
		Type:  "match_update",
		Match: update,
	}
	payload, _ := json.Marshal(message)
	writeChannel <- payload
}

// Adds the player to the matchmaking queue, once a game is found, the write channel will be used to create
// a watcher for the game. The onGameID callback will be called with the gameID once assigned.
func (hub *MatchmakingHub) JoinQueue(ctx context.Context, playerName string, writeChannel chan json.RawMessage, onGameID OnGameIDCallback) context.CancelFunc {
	hub.GetStorageService().PutUserIntoQueue(context.Background(), playerName)
	lobbyChannel, cancelFunc := hub.GetStorageService().WatchUserLobby(ctx, playerName)
	go func() {
		log.Printf("matchmaking: started watching lobby for player %s", playerName)
		for lobbyUpdate := range lobbyChannel {
			cancelFunc = hub.SetGameWatcher(ctx, string(lobbyUpdate), writeChannel)
			if onGameID != nil {
				onGameID(string(lobbyUpdate))
				log.Printf("matchmaking: lobby update for player %s: %s, starting watcher and returning", playerName, string(lobbyUpdate))
				return
			}
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
