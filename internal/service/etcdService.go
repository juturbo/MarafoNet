package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const MATCH_COUNTER_PATH = "global/match_counter"

type EtcdService struct {
	client *clientv3.Client
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
	return &EtcdService{client: client}, nil
}

func (etcdService *EtcdService) GetNextMatchID(ctx context.Context) (matchId string, err error) {
	key := MATCH_COUNTER_PATH
	for {
		current, revision, err := etcdService.fetchCurrentAndRevision(ctx, key)
		if err != nil {
			return "", err
		}
		next := current + 1
		isUpdateSuccessful, err := etcdService.updateValueIfRevisionMatches(ctx, key, revision, next)
		if err != nil {
			return "", err
		}
		if isUpdateSuccessful {
			matchId := fmt.Sprintf("match/%d", next)
			return matchId, nil
		}
	}
}

func (etcdService *EtcdService) PutNewGame(ctx context.Context, key string, matchJson []byte) (bool, error) {
	transaction := etcdService.client.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, string(matchJson)))

	transactionResponse, err := transaction.Commit()
	if err != nil {
		return false, err
	}
	return transactionResponse.Succeeded, nil
}

func (etcdService *EtcdService) GetValueAndRevision(ctx context.Context, key string) (matchJson json.RawMessage, revision int64, err error) {
	response, err := etcdService.client.Get(ctx, key)
	if err != nil {
		return matchJson, revision, err
	}
	if len(response.Kvs) == 0 {
		err = fmt.Errorf("key not found: %s", key)
		return matchJson, revision, err
	}
	matchJson = response.Kvs[0].Value
	revision = response.Kvs[0].ModRevision
	return matchJson, revision, nil
}

func (etcdService *EtcdService) PutUpdatedGameIfRevisionMatch(ctx context.Context, key string, matchJson []byte, lastRevision int64) (bool, error) {
	transaction := etcdService.client.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(key), "=", lastRevision)).
		Then(clientv3.OpPut(key, string(matchJson)))
	transactionResponse, err := transaction.Commit()
	if err != nil {
		return false, err
	}
	return transactionResponse.Succeeded, nil
}

func (etcdService *EtcdService) SetUserCurrentMatchId(ctx context.Context, playerName string, matchId string) error {
	key := fmt.Sprintf("user/%s/current_match", playerName)
	_, err := etcdService.client.Put(ctx, key, matchId)
	return err
}

func (etcdService *EtcdService) GetUserCurrentMatchId(ctx context.Context, playerName string) (string, error) {
	key := fmt.Sprintf("user/%s/current_match", playerName)
	resp, err := etcdService.client.Get(ctx, key)
	if err != nil || len(resp.Kvs) == 0 {
		return "", fmt.Errorf("nessuna partita attiva trovata per l'utente")
	}
	return string(resp.Kvs[0].Value), nil
}

func (etcdService *EtcdService) WatchMatch(ctx context.Context, matchId string) (<-chan []byte, context.CancelFunc) {
	channel := make(chan []byte)
	watchCtx, cancel := context.WithCancel(ctx)
	watchChannel := etcdService.client.Watch(watchCtx, matchId)
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
	transaction := etcdService.client.Txn(ctx).
		If(compare).
		Then(clientv3.OpPut(key, strconv.Itoa(value)))
	transactionResponse, err := transaction.Commit()
	if err != nil {
		return false, err
	}
	return transactionResponse.Succeeded, nil
}
