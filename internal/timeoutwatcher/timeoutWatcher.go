package timeoutwatcher

import (
	"MarafoNet/internal/service"
	"MarafoNet/internal/storage"
	"context"
	"log"
	"sync"
)

type TimeoutWatcher struct {
	ctx           context.Context
	cancel        context.CancelFunc
	storage       storage.MatchmakingStorage
	gameService   service.GameService
	watcherCancel context.CancelFunc
}

func NewTimeoutWatcher(storage storage.MatchmakingStorage, gameService service.GameService) *TimeoutWatcher {
	ctx, cancel := context.WithCancel(context.Background())
	log.Printf("- [timeout watcher]: Timeout watcher initialized")
	return &TimeoutWatcher{
		ctx:         ctx,
		cancel:      cancel,
		storage:     storage,
		gameService: gameService,
	}
}

func (tw *TimeoutWatcher) Start(wg *sync.WaitGroup) {
	timeoutWatcher, cancelTimeoutWatcher := tw.storage.WatchUserTimeoutLease(context.Background())
	tw.watcherCancel = cancelTimeoutWatcher

	go func() {
		log.Printf("- [timeout watcher]: started watching timeout lease")
		for {
			select {
			case <-tw.ctx.Done():
				cancelTimeoutWatcher()
				wg.Done()
				log.Printf("- [timeout watcher]: Stopping timeout watcher service")
				return
			default:
				timeoutEvent, ok := <-timeoutWatcher
				if !ok {
					log.Printf("- [timeout watcher]: timeout channel closed")
					return
				}
				log.Printf("- [timeout watcher]: received timeout event for user %s in game %s", timeoutEvent.Username, timeoutEvent.GameID)
				tw.handleTimeoutEvent(timeoutEvent)
			}
		}
	}()
}

func (tw *TimeoutWatcher) Stop() {
	tw.cancel()
}

func (tw *TimeoutWatcher) handleTimeoutEvent(timeoutEvent storage.GameTimeoutEvent) {
	err := tw.gameService.ForfeitGame(context.Background(), timeoutEvent.GameID, timeoutEvent.Username)
	if err != nil {
		log.Printf("- [timeout watcher]: error forfeiting game %s for user %s: %v", timeoutEvent.GameID, timeoutEvent.Username, err)
	} else {
		log.Printf("- [timeout watcher]: successfully forfeited game %s for user %s", timeoutEvent.GameID, timeoutEvent.Username)
	}

	err = tw.storage.RemoveUserCurrentGameId(context.Background(), timeoutEvent.Username)
	if err != nil {
		log.Printf("- [timeout watcher]: error removing current game ID for user %s: %v", timeoutEvent.Username, err)
	}
}
