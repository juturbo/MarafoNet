package storage

import (
	"context"
	"fmt"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdWatcherService struct {
	core           *etcdCore
	userSession    *etcdUserSessionService
	watcherFactory WatcherFactory
}

func (service *etcdWatcherService) WatchGame(ctx context.Context, gameId string) (<-chan []byte, context.CancelFunc) {
	response, err := service.core.client.Get(ctx, gameId)
	startRevision := int64(0)
	if err == nil {
		startRevision = response.Header.Revision
	}
	return service.watcherFactory.WatchBytes(ctx, WatcherConfig{
		Key:           gameId,
		EventType:     clientv3.EventTypePut,
		StartRevision: startRevision,
	})
}

func (service *etcdWatcherService) WatchUserLobby(ctx context.Context, username string) (<-chan []byte, context.CancelFunc) {
	key := service.core.pathBuilder.UserCurrentGamePath(username)
	response, err := service.core.client.Get(ctx, key)
	startRevision := int64(0)
	if err == nil {
		startRevision = response.Header.Revision
	}
	return service.watcherFactory.WatchBytes(ctx, WatcherConfig{
		Key:           key,
		EventType:     clientv3.EventTypePut,
		StartRevision: startRevision,
	})
}

func (service *etcdWatcherService) WatchUserQueue(ctx context.Context) (<-chan []string, context.CancelFunc) {
	key := service.core.pathBuilder.UserQueuePrefix()
	response, err := service.core.client.Get(ctx, key, clientv3.WithPrefix())
	startRevision := int64(0)
	if err == nil {
		startRevision = response.Header.Revision
	}
	return service.watcherFactory.WatchStrings(ctx, WatcherConfig{
		Key:           key,
		Prefix:        true,
		EventType:     clientv3.EventTypePut,
		StartRevision: startRevision,
		Transform: func(eventValue []byte) (interface{}, error) {
			return service.userSession.GetUserQueue(ctx)
		},
	})
}

func (service *etcdWatcherService) WatchUserTimeoutLease(ctx context.Context) (<-chan GameTimeoutEvent, context.CancelFunc) {
	key := service.core.pathBuilder.GameTimeoutPrefix()
	response, err := service.core.client.Get(ctx, key, clientv3.WithPrefix())
	startRevision := int64(0)
	if err == nil {
		startRevision = response.Header.Revision
	}
	return service.watcherFactory.WatchTimeoutEvents(ctx, WatcherConfig{
		Key:           key,
		Prefix:        true,
		EventType:     clientv3.EventTypeDelete,
		StartRevision: startRevision,
		Transform: func(keyPath []byte) (interface{}, error) {
			return service.parseGameTimeoutEvent(keyPath)
		},
		Validate: func(event interface{}) bool {
			timeoutEvent := event.(GameTimeoutEvent)
			isOnline, err := service.userSession.isUserConnected(ctx, timeoutEvent.Username)
			return err == nil && !isOnline
		},
	})
}

func (service *etcdWatcherService) parseGameTimeoutEvent(keyPath []byte) (GameTimeoutEvent, error) {
	timeoutKey := string(keyPath)
	prefix := service.core.pathBuilder.GameTimeoutPrefix()
	if !strings.HasPrefix(timeoutKey, prefix) {
		return GameTimeoutEvent{}, fmt.Errorf("invalid timeout key: %s", timeoutKey)
	}

	trimmed := strings.TrimPrefix(timeoutKey, prefix)
	last := strings.LastIndex(trimmed, "/")
	if last <= 0 {
		return GameTimeoutEvent{}, fmt.Errorf("invalid timeout key format: %s", trimmed)
	}

	gameId := trimmed[:last]
	username := trimmed[last+1:]

	return GameTimeoutEvent{GameID: gameId, Username: username}, nil
}
