package storage

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const IS_ONLINE = "1"
const IS_OFFLINE = "0"
const TIMEOUT = "1"
const KEEP_ALIVE_TTL = 120

type EtcdService struct {
	*etcdGameRepositoryService
	*etcdUserSessionService
	*etcdWatcherService
	core *etcdCore
}

type etcdCore struct {
	client      *clientv3.Client
	pathBuilder PathBuilder
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
	core := &etcdCore{
		client:      client,
		pathBuilder: NewPathBuilder(),
	}
	userSession := &etcdUserSessionService{core: core}
	watcherFactory := NewWatcherFactory(client)
	return &EtcdService{
		etcdGameRepositoryService: &etcdGameRepositoryService{core: core},
		etcdUserSessionService:    userSession,
		etcdWatcherService:        &etcdWatcherService{core: core, session: userSession, watcherFactory: watcherFactory},
		core:                      core,
	}, nil
}

func (core *etcdCore) Close() error {
	return core.client.Close()
}

func (core *etcdCore) getValue(ctx context.Context, key string) (string, error) {
	value, _, err := core.getKeyValue(ctx, key)
	return value, err
}

func (core *etcdCore) putValue(ctx context.Context, key string, value string) error {
	_, err := core.client.Put(ctx, key, value)
	return err
}

func (core *etcdCore) deleteKey(ctx context.Context, key string) error {
	_, err := core.client.Delete(ctx, key)
	return err
}

func (core *etcdCore) getKeyValue(ctx context.Context, key string) (value string, revision int64, err error) {
	response, err := core.client.Get(ctx, key)
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

func (core *etcdCore) keyExists(ctx context.Context, key string) (bool, error) {
	response, err := core.client.Get(ctx, key)
	if err != nil {
		return false, err
	}

	return len(response.Kvs) > 0, nil
}

func (core *etcdCore) putIfComparison(ctx context.Context, key string, value string, compare clientv3.Cmp) (bool, error) {
	transaction := core.client.Txn(ctx).
		If(compare).
		Then(clientv3.OpPut(key, value))

	transactionResponse, err := transaction.Commit()
	if err != nil {
		return false, err
	}

	return transactionResponse.Succeeded, nil
}
