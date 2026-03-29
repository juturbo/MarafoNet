package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	uuid "github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const MATCH_COUNTER_PATH = "global/match_counter"
const KEEP_ALIVE_TTL = 300

type EtcdService struct {
	client *clientv3.Client
	uuid   uuid.UUID
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
			return fmt.Sprintf("match/%d", next), nil
		}
	}
}

func (etcdService *EtcdService) PutNewGame(ctx context.Context, key string, matchJson json.RawMessage) error {
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

func (etcdService *EtcdService) GetValueAndRevision(ctx context.Context, key string) (matchJson json.RawMessage, revision int64, err error) {
	value, rev, err := etcdService.getKeyValue(ctx, key)
	if err != nil {
		return nil, 0, err
	}

	return json.RawMessage(value), rev, nil
}

func (etcdService *EtcdService) PutUpdatedGameIfRevisionMatch(ctx context.Context, key string, matchJson json.RawMessage, lastRevision int64) error {
	succeeded, err := etcdService.putIfComparison(
		ctx,
		key,
		string(matchJson),
		clientv3.Compare(clientv3.ModRevision(key), "=", lastRevision),
	)
	if err != nil {
		return err
	}
	if !succeeded {
		return fmt.Errorf("failed to update match: revision mismatch")
	}
	return nil
}

func (etcdService *EtcdService) SetUserCurrentMatchId(ctx context.Context, playerName string, matchId string) error {
	key := fmt.Sprintf("user/%s/current_match", playerName)
	return etcdService.putValue(ctx, key, matchId)
}

func (etcdService *EtcdService) GetUserCurrentMatchId(ctx context.Context, playerName string) (string, error) {
	key := fmt.Sprintf("user/%s/current_match", playerName)
	return etcdService.getValue(ctx, key)
}

func (etcdService *EtcdService) RegisterUser(ctx context.Context, playerName string) (string, error) {
	uuidKey := fmt.Sprintf("user/%s/uuid", playerName)
	uuidValue := etcdService.uuid.String()

	err := etcdService.putValue(ctx, uuidKey, uuidValue)
	if err != nil {
		return "", err
	}
	return uuidValue, nil
}

func (etcdService *EtcdService) VerifyUser(ctx context.Context, playerName string, uuid string) error {
	uuidKey := fmt.Sprintf("user/%s/uuid", playerName)
	value, err := etcdService.getValue(ctx, uuidKey)
	if err != nil || value != uuid {
		return fmt.Errorf("authentication failed for user: %s", playerName)
	}

	return nil
}

func (etcdService *EtcdService) OnUserDisconnect(ctx context.Context, playerName string) error {
	uuidKey := fmt.Sprintf("user/%s/uuid", playerName)

	value, err := etcdService.getValue(ctx, uuidKey)
	if err != nil {
		return err
	}

	lease, err := etcdService.client.Grant(ctx, KEEP_ALIVE_TTL)
	if err != nil {
		return err
	}

	_, err = etcdService.client.Put(ctx, uuidKey, value, clientv3.WithLease(lease.ID))
	return err
}

func (etcdService *EtcdService) OnUserReconnect(ctx context.Context, playerName string, uuid string) error {
	uuidKey := fmt.Sprintf("user/%s/uuid", playerName)

	if err := etcdService.VerifyUser(ctx, playerName, uuid); err != nil {
		return err
	}

	return etcdService.putValue(ctx, uuidKey, uuid)
}

func (etcdService *EtcdService) WatchGame(ctx context.Context, matchId string) (<-chan json.RawMessage, context.CancelFunc) {
	return etcdService.watchKey(ctx, matchId)
}

func (etcdService *EtcdService) WatchUserLobby(ctx context.Context, playerName string) (<-chan json.RawMessage, context.CancelFunc) {
	key := fmt.Sprintf("user/%s/current_match", playerName)
	return etcdService.watchKey(ctx, key)
}

func (etcdService *EtcdService) watchKey(ctx context.Context, key string) (<-chan json.RawMessage, context.CancelFunc) {
	channel := make(chan json.RawMessage)
	watchCtx, cancel := context.WithCancel(ctx)
	watchChannel := etcdService.client.Watch(watchCtx, key)
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

func (etcdService *EtcdService) getKeyValue(ctx context.Context, key string) (value string, revision int64, err error) {
	response, err := etcdService.client.Get(ctx, key)
	if err != nil {
		return "", 0, err
	}
	if len(response.Kvs) == 0 {
		return "", 0, fmt.Errorf("key not found: %s", key)
	}

	value = string(response.Kvs[0].Value)
	revision = response.Kvs[0].ModRevision
	return value, revision, nil
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
