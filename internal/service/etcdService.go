package service

import (
	"MarafoNet/model"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	uuid "github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const MATCH_COUNTER_PATH = "global/match_counter"
const MATCH_PREFIX = "match/%d"
const MATCH_TIMEOUT_PREFIX = "match_timeout/"
const MATCH_TIMEOUT_PATH = MATCH_TIMEOUT_PREFIX + "%s/%s" // match_timeout/{matchId}/{playerName}
const USER_QUEUE_PATH = "user_queue/"
const USERS_NAME_PATH = "users/%s"
const USERS_PASSWORD_PATH = "users/%s/password"
const USERS_IS_CONNECTED_PATH = "users/%s/is_connected"
const USERS_CURRENT_MATCH_PATH = "users/%s/current_match"
const IS_ONLINE = "1"
const IS_OFFLINE = "0"
const TIMEOUT = "1"
const KEEP_ALIVE_TTL = 120

type EtcdService struct {
	client *clientv3.Client
	uuid   uuid.UUID
}

// MatchTimeoutEvent is emitted by the watcher when a user's match timeout lease expires.
type MatchTimeoutEvent struct {
	MatchID  string
	Username string
}

func NewEtcdService(endpoints []string, dialTimeout time.Duration) (*EtcdService, error) {
	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}

	return &EtcdService{
		client: client,
		uuid:   uuid.New(),
	}, nil
}

func (etcdService *EtcdService) GetNextMatchID(ctx context.Context) (matchId string, err error) {
	key := MATCH_COUNTER_PATH

	for {
		current, revision, err := etcdService.fetchCurrentAndRevision(ctx, key)
		if err != nil {
			return "", err
		}

		next := current + 1
		succeeded, err := etcdService.updateValueIfRevisionMatches(ctx, key, revision, next)
		if err != nil {
			return "", err
		}
		if succeeded {
			return fmt.Sprintf(MATCH_PREFIX, next), nil
		}
	}
}

func (etcdService *EtcdService) PutNewGame(ctx context.Context, key string, matchJson []byte) error {
	succeeded, err := etcdService.putIfComparison(
		ctx,
		key,
		string(matchJson),
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

func (etcdService *EtcdService) GetMatchJsonAndRevision(ctx context.Context, key string) (matchJson json.RawMessage, revision int64, err error) {
	value, revision, err := etcdService.getKeyValue(ctx, key)
	if err != nil {
		return nil, 0, err
	}

	if value == "" {
		return nil, 0, fmt.Errorf("key not found: %s", key)
	}

	return json.RawMessage(value), revision, nil
}

func (etcdService *EtcdService) PutUpdatedGameJsonIfRevisionMatch(ctx context.Context, matchId string, matchJson json.RawMessage, lastRevision int64) error {
	succeeded, err := etcdService.putIfComparison(
		ctx,
		matchId,
		string(matchJson),
		clientv3.Compare(clientv3.ModRevision(matchId), "=", lastRevision),
	)
	if err != nil {
		return err
	}
	if !succeeded {
		return fmt.Errorf("failed to update match: revision mismatch")
	}
	return nil
}

func (etcdService *EtcdService) PutUserIntoQueue(ctx context.Context, playerName string) (err error) {
	key := USER_QUEUE_PATH + playerName

	if err = etcdService.putValue(ctx, key, playerName); err != nil {
		return err
	}

	return nil
}

func (etcdService *EtcdService) GetUserQueue(ctx context.Context) (userQueue []string, err error) {
	key := USER_QUEUE_PATH

	response, err := etcdService.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(response.Kvs) == 0 {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	for _, kv := range response.Kvs {
		userQueue = append(userQueue, string(kv.Value))
	}

	return userQueue, nil
}

func (etcdService *EtcdService) RemoveUserFromQueue(ctx context.Context, playerName string) error {
	key := USER_QUEUE_PATH + playerName
	return etcdService.deleteKey(ctx, key)
}

func (etcdService *EtcdService) SetUserCurrentMatchId(ctx context.Context, playerName string, matchId string) error {
	key := fmt.Sprintf(USERS_CURRENT_MATCH_PATH, playerName)
	return etcdService.putValue(ctx, key, matchId)
}

func (etcdService *EtcdService) GetUserCurrentMatchId(ctx context.Context, playerName string) (string, error) {
	key := fmt.Sprintf(USERS_CURRENT_MATCH_PATH, playerName)
	return etcdService.getValue(ctx, key)
}

func (etcdService *EtcdService) RemoveUserCurrentMatchId(ctx context.Context, playerName string) error {
	key := fmt.Sprintf(USERS_CURRENT_MATCH_PATH, playerName)
	return etcdService.deleteKey(ctx, key)
}

func (etcdService *EtcdService) RegisterUser(ctx context.Context, user model.User) error {
	userKey := fmt.Sprintf(USERS_NAME_PATH, user.Name)
	passwordKey := fmt.Sprintf(USERS_PASSWORD_PATH, user.Name)
	isConnectedKey := fmt.Sprintf(USERS_IS_CONNECTED_PATH, user.Name)
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

	if etcdService.isUserInAMatch(ctx, user.Name) {
		return etcdService.removeUserTimeout(ctx, user.Name)
	}

	return nil
}

func (etcdService *EtcdService) OnUserDisconnect(ctx context.Context, playerName string) error {
	err := etcdService.setUserOfflineStatus(ctx, playerName)
	if err != nil {
		return err
	}

	if etcdService.isUserInAMatch(ctx, playerName) {
		err := etcdService.setUserMatchTimeout(ctx, playerName)
		if err != nil {
			return err
		}
	}

	return err
}

func (etcdService *EtcdService) WatchGame(ctx context.Context, matchId string) (<-chan []byte, context.CancelFunc) {
	return etcdService.watchKey(ctx, matchId)
}

func (etcdService *EtcdService) WatchUserLobby(ctx context.Context, username string) (<-chan []byte, context.CancelFunc) {
	key := fmt.Sprintf(USERS_CURRENT_MATCH_PATH, username)
	return etcdService.watchKey(ctx, key)
}

func (etcdService *EtcdService) WatchUserQueue(ctx context.Context) (<-chan []string, context.CancelFunc) {
	key := USER_QUEUE_PATH

	channel := make(chan []string)
	watchCtx, cancel := context.WithCancel(ctx)
	watchChannel := etcdService.client.Watch(watchCtx, key, clientv3.WithPrefix())
	go func() {
		defer close(channel)
		for watchResponse := range watchChannel {
			if watchResponse.Err() != nil {
				return
			}
			for _, event := range watchResponse.Events {
				if event.Type == clientv3.EventTypePut {
					var userQueue, _ = etcdService.GetUserQueue(ctx)
					channel <- userQueue
				}
			}
		}
	}()

	return channel, cancel
}

func (etcdService *EtcdService) WatchUserTimeoutLease(ctx context.Context) (<-chan MatchTimeoutEvent, context.CancelFunc) {
	channel := make(chan MatchTimeoutEvent)
	watchCtx, cancel := context.WithCancel(ctx)
	watchChannel := etcdService.client.Watch(watchCtx, MATCH_TIMEOUT_PREFIX, clientv3.WithPrefix())
	go func() {
		defer close(channel)
		for watchResponse := range watchChannel {
			if watchResponse.Err() != nil {
				return
			}
			for _, event := range watchResponse.Events {
				if event.Type == clientv3.EventTypeDelete {
					if event.Kv == nil {
						continue
					}
					timeoutKey := string(event.Kv.Key)
					if !strings.HasPrefix(timeoutKey, MATCH_TIMEOUT_PREFIX) {
						continue
					}
					trimmed := strings.TrimPrefix(timeoutKey, MATCH_TIMEOUT_PREFIX) // "{matchId}/{username}"
					last := strings.LastIndex(trimmed, "/")
					if last <= 0 {
						continue
					}
					matchId := trimmed[:last]
					username := trimmed[last+1:]
					isOnline, err := etcdService.isUserConnected(ctx, username)
					if err == nil && !isOnline { // If lease expired and user is still offline, notify timeout
						channel <- MatchTimeoutEvent{MatchID: matchId, Username: username}
					}
				}
			}
		}
	}()

	return channel, cancel
}

func (etcdService *EtcdService) watchKey(ctx context.Context, key string, opts ...clientv3.OpOption) (<-chan []byte, context.CancelFunc) {
	channel := make(chan []byte)
	watchCtx, cancel := context.WithCancel(ctx)
	watchChannel := etcdService.client.Watch(watchCtx, key, opts...)
	go func() {
		defer close(channel)
		for watchResponse := range watchChannel {
			if watchResponse.Err() != nil {
				return
			}
			for _, event := range watchResponse.Events {
				if event.Type == clientv3.EventTypePut {
					channel <- event.Kv.Value
				}
			}
		}
	}()
	return channel, cancel
}

func (etcdService *EtcdService) Close() error {
	return etcdService.client.Close()
}

func (etcdService *EtcdService) fetchCurrentAndRevision(ctx context.Context, key string) (current int, revision int64, err error) {
	response, err := etcdService.client.Get(ctx, key)
	if err != nil {
		return current, revision, err
	}
	if len(response.Kvs) == 0 {
		return current, revision, nil
	}

	current, err = strconv.Atoi(string(response.Kvs[0].Value))
	if err != nil {
		return current, revision, err
	}

	revision = response.Kvs[0].ModRevision
	return current, revision, nil
}

func (etcdService *EtcdService) updateValueIfRevisionMatches(ctx context.Context, key string, expectedRevision int64, value int) (bool, error) {
	var compare clientv3.Cmp
	if expectedRevision == 0 {
		compare = clientv3.Compare(clientv3.CreateRevision(key), "=", 0)
	} else {
		compare = clientv3.Compare(clientv3.ModRevision(key), "=", expectedRevision)
	}
	return etcdService.putIfComparison(ctx, key, strconv.Itoa(value), compare)
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
	passwordKey := fmt.Sprintf(USERS_PASSWORD_PATH, user.Name)

	hashedPassword, err := etcdService.getValue(ctx, passwordKey)
	if err != nil {
		return false, err
	}

	return user.CheckPassword(hashedPassword), nil
}

func (etcdService *EtcdService) updateUserConnectionStatus(ctx context.Context, playerName string, expect string, newState string) error {
	isConnectedKey := fmt.Sprintf(USERS_IS_CONNECTED_PATH, playerName)

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

func (etcdService *EtcdService) setUserMatchTimeout(ctx context.Context, playerName string) error {
	matchId, err := etcdService.GetUserCurrentMatchId(ctx, playerName)
	if err != nil {
		return fmt.Errorf("failed to get user's current match ID: %w", err)
	}

	lease, err := etcdService.client.Grant(ctx, KEEP_ALIVE_TTL)
	if err != nil {
		return err
	}

	matchTimeoutKey := fmt.Sprintf(MATCH_TIMEOUT_PATH, matchId, playerName)
	_, err = etcdService.client.Put(ctx, matchTimeoutKey, TIMEOUT, clientv3.WithLease(lease.ID))
	return nil
}

func (etcdService *EtcdService) removeUserTimeout(ctx context.Context, playerName string) error {
	matchId, err := etcdService.GetUserCurrentMatchId(ctx, playerName)
	if err != nil {
		return fmt.Errorf("failed to get user's current match ID: %w", err)
	}

	matchTimeoutKey := fmt.Sprintf(MATCH_TIMEOUT_PATH, matchId, playerName)

	return etcdService.deleteKey(ctx, matchTimeoutKey)
}

func (etcdService *EtcdService) isUserInAMatch(ctx context.Context, playerName string) bool {
	matchId, err := etcdService.GetUserCurrentMatchId(ctx, playerName)
	return matchId != "" && err == nil
}

func (etcdService *EtcdService) isUserConnected(ctx context.Context, playerName string) (bool, error) {
	key := fmt.Sprintf(USERS_IS_CONNECTED_PATH, playerName)
	value, err := etcdService.getValue(ctx, key)
	if err != nil {
		return false, err
	}
	return value == IS_ONLINE, nil
}
