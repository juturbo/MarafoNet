package repository

import (
	"MarafoNet/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	concurrency "go.etcd.io/etcd/client/v3/concurrency"
)

const IS_ONLINE = "1"
const IS_OFFLINE = "0"
const TIMEOUT = "1"
const KEEP_ALIVE_TTL = 120
const DEFAULT_MAX_RETRIES = 3
const DEFAULT_INITIAL_BACKOFF = 100 * time.Millisecond
const DEFAULT_MAX_BACKOFF = 10 * time.Second
const DEFAULT_BACKOFF_MULTIPLIER = 2.0

type EtcdService struct {
	client         *clientv3.Client
	pathBuilder    PathBuilder
	watcherFactory WatcherFactory
}

// GameTimeoutEvent is emitted by the watcher when a user's game timeout lease expires.
type GameTimeoutEvent struct {
	GameID   string
	Username string
}

func NewEtcdService(endpoints []string, dialTimeout time.Duration) (*EtcdService, error) {
	cfg := clientv3.Config{
		Endpoints:            endpoints,
		DialTimeout:          dialTimeout,
		DialKeepAliveTime:    5 * time.Second,
		DialKeepAliveTimeout: 3 * time.Hour,
	}
	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	return &EtcdService{
		client:         client,
		pathBuilder:    NewPathBuilder(),
		watcherFactory: NewWatcherFactory(client),
	}, nil
}

func (etcdService *EtcdService) GetNextGameID() (gameId string, err error) {
	ctx := context.Background()

	return gameId, etcdService.retryWithBackoff(ctx, "GetNextGameID", func() error {
		key := etcdService.pathBuilder.GameCounterPath()

		var next int
		transactionResponse, err := concurrency.NewSTM(etcdService.client, func(stm concurrency.STM) error {
			val := stm.Get(key)
			if val == "" {
				next = 1
			} else {
				current, convErr := strconv.Atoi(val)
				if convErr != nil {
					return convErr
				}
				next = current + 1
			}
			stm.Put(key, strconv.Itoa(next))
			return nil
		})
		if err != nil {
			return err
		}
		if !transactionResponse.Succeeded {
			return fmt.Errorf("failed to get next game ID")
		}

		gameId = etcdService.pathBuilder.GamePath(strconv.Itoa(next))
		return nil
	})
}

func (etcdService *EtcdService) PutNewGame(ctx context.Context, key string, gameJson []byte) error {
	succeeded, err := etcdService.putIfComparison(
		ctx,
		key,
		string(gameJson),
		clientv3.Compare(clientv3.CreateRevision(key), "=", 0),
	)
	if err != nil {
		return err
	}
	if !succeeded {
		return fmt.Errorf("failed to create new game: key already exists")
	}
	return nil
}

func (etcdService *EtcdService) GetGameJsonAndRevision(ctx context.Context, key string) (gameJson json.RawMessage, revision int64, err error) {
	value, revision, err := etcdService.getKeyValue(ctx, key)
	if err != nil {
		return nil, 0, err
	}

	if value == "" {
		return nil, 0, fmt.Errorf("key not found: %s", key)
	}

	return json.RawMessage(value), revision, nil
}

func (etcdService *EtcdService) PutUpdatedGameJsonIfRevisionMatch(ctx context.Context, gameId string, gameJson json.RawMessage, lastRevision int64) error {
	succeeded, err := etcdService.putIfComparison(
		ctx,
		gameId,
		string(gameJson),
		clientv3.Compare(clientv3.ModRevision(gameId), "=", lastRevision),
	)
	if err != nil {
		return err
	}
	if !succeeded {
		return fmt.Errorf("failed to update game: revision mismatch")
	}
	return nil
}

func (etcdService *EtcdService) PutUserIntoQueue(ctx context.Context, playerName string) (err error) {
	key := etcdService.pathBuilder.UserQueuePath(playerName)

	if err = etcdService.putValue(ctx, key, playerName); err != nil {
		return err
	}

	return nil
}

func (etcdService *EtcdService) GetUserQueue(ctx context.Context) (userQueue []string, err error) {
	key := etcdService.pathBuilder.UserQueuePrefix()

	response, err := etcdService.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	for _, kv := range response.Kvs {
		userQueue = append(userQueue, string(kv.Value))
	}

	return userQueue, nil
}

func (etcdService *EtcdService) RemoveUserFromQueue(ctx context.Context, playerName string) error {
	key := etcdService.pathBuilder.UserQueuePath(playerName)
	return etcdService.deleteKey(ctx, key)
}

func (etcdService *EtcdService) SetUserCurrentGameId(ctx context.Context, playerName string, gameId string) error {
	key := etcdService.pathBuilder.UserCurrentGamePath(playerName)
	return etcdService.putValue(ctx, key, gameId)
}

func (etcdService *EtcdService) GetUserCurrentGameId(ctx context.Context, playerName string) (string, error) {
	key := etcdService.pathBuilder.UserCurrentGamePath(playerName)
	return etcdService.getValue(ctx, key)
}

func (etcdService *EtcdService) RemoveUserCurrentGameId(ctx context.Context, playerName string) error {
	key := etcdService.pathBuilder.UserCurrentGamePath(playerName)
	return etcdService.deleteKey(ctx, key)
}

func (etcdService *EtcdService) RegisterUser(ctx context.Context, user model.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	userKey := etcdService.pathBuilder.UserPath(user.Name)
	passwordKey := etcdService.pathBuilder.UserPasswordPath(user.Name)
	isConnectedKey := etcdService.pathBuilder.UserConnectionPath(user.Name)
	hashedPassword, err := user.GeneratePasswordHash()
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	transaction := etcdService.client.Txn(ctx).
		If(
			clientv3.Compare(clientv3.CreateRevision(userKey), "=", 0),
			clientv3.Compare(clientv3.CreateRevision(passwordKey), "=", 0),
		).
		Then(
			clientv3.OpPut(userKey, user.Name),
			clientv3.OpPut(passwordKey, hashedPassword),
			clientv3.OpPut(isConnectedKey, IS_OFFLINE),
		)

	transactionResponse, err := transaction.Commit()
	if err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}
	if !transactionResponse.Succeeded {
		return fmt.Errorf("username not available: %s", user.Name)
	}

	return nil
}

func (etcdService *EtcdService) LoginUser(ctx context.Context, user model.User) error {
	isValid, err := etcdService.verifyUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}
	if !isValid {
		return fmt.Errorf("invalid password for user: %s", user.Name)
	}

	if err = etcdService.setUserOnlineStatus(ctx, user.Name); err != nil {
		return err
	}

	if etcdService.isUserInAGame(ctx, user.Name) {
		err := etcdService.removeUserTimeout(ctx, user.Name)
		return err
	}

	return nil
}

func (etcdService *EtcdService) OnUserDisconnect(ctx context.Context, playerName string) error {
	err := etcdService.setUserOfflineStatus(ctx, playerName)
	if err != nil {
		return err
	}

	if etcdService.isUserInAGame(ctx, playerName) {
		err := etcdService.setUserGameTimeout(ctx, playerName)
		if err != nil {
			return err
		}
	}

	return err
}

func (etcdService *EtcdService) WatchGame(ctx context.Context, gameId string) (<-chan []byte, context.CancelFunc) {
	return etcdService.watcherFactory.WatchBytes(ctx, WatcherConfig{
		Key:       gameId,
		EventType: clientv3.EventTypePut,
	})
}

func (etcdService *EtcdService) WatchUserLobby(ctx context.Context, username string) (<-chan []byte, context.CancelFunc) {
	return etcdService.watcherFactory.WatchBytes(ctx, WatcherConfig{
		Key:       etcdService.pathBuilder.UserCurrentGamePath(username),
		EventType: clientv3.EventTypePut,
	})
}

func (etcdService *EtcdService) WatchUserQueue(ctx context.Context) (<-chan []string, context.CancelFunc) {
	return etcdService.watcherFactory.WatchStrings(ctx, WatcherConfig{
		Key:       etcdService.pathBuilder.UserQueuePrefix(),
		Prefix:    true,
		EventType: clientv3.EventTypePut,
		Transform: func(eventValue []byte) (interface{}, error) {
			return etcdService.GetUserQueue(ctx)
		},
	})
}

func (etcdService *EtcdService) WatchUserTimeoutLease(ctx context.Context) (<-chan GameTimeoutEvent, context.CancelFunc) {
	return etcdService.watcherFactory.WatchTimeoutEvents(ctx, WatcherConfig{
		Key:       etcdService.pathBuilder.GameTimeoutPrefix(),
		Prefix:    true,
		EventType: clientv3.EventTypeDelete,
		Transform: func(keyPath []byte) (interface{}, error) {
			return etcdService.parseGameTimeoutEvent(keyPath)
		},
		Validate: func(event interface{}) bool {
			timeoutEvent := event.(GameTimeoutEvent)
			isOnline, err := etcdService.isUserConnected(ctx, timeoutEvent.Username)
			return err == nil && !isOnline
		},
	})
}

func (etcdService *EtcdService) Close() error {
	return etcdService.client.Close()
}

func (etcdService *EtcdService) retryWithBackoff(
	ctx context.Context,
	operation string,
	fn func() error,
) error {
	var lastErr error
	backoff := DEFAULT_INITIAL_BACKOFF

	for attempt := 0; attempt <= DEFAULT_MAX_RETRIES; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt < DEFAULT_MAX_RETRIES {
			select {
			case <-time.After(backoff):
				if backoff < DEFAULT_MAX_BACKOFF {
					backoff = time.Duration(float64(backoff) * DEFAULT_BACKOFF_MULTIPLIER)
					if backoff > DEFAULT_MAX_BACKOFF {
						backoff = DEFAULT_MAX_BACKOFF
					}
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("operation %q failed after %d retries: %w", operation, DEFAULT_MAX_RETRIES, lastErr)
}

func (etcdService *EtcdService) getValue(ctx context.Context, key string) (string, error) {
	value, _, err := etcdService.getKeyValue(ctx, key)
	return value, err
}

func (etcdService *EtcdService) putValue(ctx context.Context, key string, value string) error {
	_, err := etcdService.client.Put(ctx, key, value)
	return err
}

func (etcdService *EtcdService) deleteKey(ctx context.Context, key string) error {
	_, err := etcdService.client.Delete(ctx, key)
	return err
}

func (etcdService *EtcdService) getKeyValue(ctx context.Context, key string) (value string, revision int64, err error) {
	response, err := etcdService.client.Get(ctx, key)
	if err != nil {
		return "", 0, err
	}
	if len(response.Kvs) == 0 {
		return "", 0, nil
	}

	value = string(response.Kvs[0].Value)
	revision = response.Kvs[0].ModRevision
	return value, revision, nil
}

func (etcdService *EtcdService) keyExists(ctx context.Context, key string) (bool, error) {
	response, err := etcdService.client.Get(ctx, key)
	if err != nil {
		return false, err
	}

	return len(response.Kvs) > 0, nil
}

func (etcdService *EtcdService) putIfComparison(ctx context.Context, key string, value string, compare clientv3.Cmp) (bool, error) {
	transaction := etcdService.client.Txn(ctx).
		If(compare).
		Then(clientv3.OpPut(key, value))

	transactionResponse, err := transaction.Commit()
	if err != nil {
		return false, err
	}

	return transactionResponse.Succeeded, nil
}

func (etcdService *EtcdService) verifyUser(ctx context.Context, user model.User) (bool, error) {
	passwordKey := etcdService.pathBuilder.UserPasswordPath(user.Name)

	hashedPassword, err := etcdService.getValue(ctx, passwordKey)
	if err != nil {
		return false, err
	}
	if hashedPassword == "" {
		return false, nil
	}
	return user.CheckPassword(hashedPassword), nil
}

func (etcdService *EtcdService) updateUserConnectionStatus(ctx context.Context, playerName string, expect string, newState string) error {
	isConnectedKey := etcdService.pathBuilder.UserConnectionPath(playerName)

	txn := etcdService.client.Txn(ctx).
		If(
			clientv3.Compare(clientv3.Value(isConnectedKey), "=", expect),
		).
		Then(
			clientv3.OpPut(isConnectedKey, newState),
		)

	transactionResponse, err := txn.Commit()
	if err != nil {
		return fmt.Errorf("failed to login user: %w", err)
	}
	if !transactionResponse.Succeeded {
		return fmt.Errorf("unexpected state transition for user %s", playerName)
	}
	return nil
}

func (etcdService *EtcdService) setUserOnlineStatus(ctx context.Context, playerName string) error {
	return etcdService.updateUserConnectionStatus(ctx, playerName, IS_OFFLINE, IS_ONLINE)
}

func (etcdService *EtcdService) setUserOfflineStatus(ctx context.Context, playerName string) error {
	return etcdService.updateUserConnectionStatus(ctx, playerName, IS_ONLINE, IS_OFFLINE)
}

func (etcdService *EtcdService) setUserGameTimeout(ctx context.Context, playerName string) error {
	gameId, err := etcdService.GetUserCurrentGameId(ctx, playerName)
	if err != nil {
		return fmt.Errorf("failed to get user's current game ID: %w", err)
	}

	lease, err := etcdService.client.Grant(ctx, KEEP_ALIVE_TTL)
	if err != nil {
		return fmt.Errorf("failed to grant lease: %w", err)
	}

	gameTimeoutKey := etcdService.pathBuilder.GameTimeoutPath(gameId, playerName)
	_, err = etcdService.client.Put(ctx, gameTimeoutKey, TIMEOUT, clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("failed to set game timeout: %w", err)
	}
	return nil
}

func (etcdService *EtcdService) removeUserTimeout(ctx context.Context, playerName string) error {
	gameId, err := etcdService.GetUserCurrentGameId(ctx, playerName)
	if err != nil {
		return fmt.Errorf("failed to get user's current game ID: %w", err)
	}

	gameTimeoutKey := etcdService.pathBuilder.GameTimeoutPath(gameId, playerName)

	return etcdService.deleteKey(ctx, gameTimeoutKey)
}

func (etcdService *EtcdService) parseGameTimeoutEvent(keyPath []byte) (GameTimeoutEvent, error) {
	timeoutKey := string(keyPath)
	prefix := etcdService.pathBuilder.GameTimeoutPrefix()
	if !strings.HasPrefix(timeoutKey, prefix) {
		return GameTimeoutEvent{}, fmt.Errorf("invalid timeout key: %s", timeoutKey)
	}

	trimmed := strings.TrimPrefix(timeoutKey, prefix) // "{gameId}/{username}"
	last := strings.LastIndex(trimmed, "/")
	if last <= 0 {
		return GameTimeoutEvent{}, fmt.Errorf("invalid timeout key format: %s", trimmed)
	}

	gameId := trimmed[:last]
	username := trimmed[last+1:]

	return GameTimeoutEvent{GameID: gameId, Username: username}, nil
}

func (etcdService *EtcdService) isUserInAGame(ctx context.Context, playerName string) bool {
	gameId, err := etcdService.GetUserCurrentGameId(ctx, playerName)
	if err != nil {
		return false
	}
	return gameId != ""
}

func (etcdService *EtcdService) isUserConnected(ctx context.Context, playerName string) (bool, error) {
	key := etcdService.pathBuilder.UserConnectionPath(playerName)
	value, err := etcdService.getValue(ctx, key)
	if err != nil {
		return false, err
	}
	return value == IS_ONLINE, nil
}
