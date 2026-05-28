package storage

import (
	"context"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// WatcherConfig defines how a watcher should behave
type WatcherConfig struct {
	Key           string
	Prefix        bool
	EventType     mvccpb.Event_EventType
	StartRevision int64
	Transform     func([]byte) (interface{}, error) // Convert event value → output type
	Validate      func(interface{}) bool            // Optional validation gate before sending
}

type WatcherFactory interface {
	WatchBytes(ctx context.Context, config WatcherConfig) (<-chan []byte, context.CancelFunc)
	WatchStrings(ctx context.Context, config WatcherConfig) (<-chan []string, context.CancelFunc)
	WatchTimeoutEvents(ctx context.Context, config WatcherConfig) (<-chan GameTimeoutEvent, context.CancelFunc)
}

type watcherFactoryImpl struct {
	etcdClient *clientv3.Client
}

func NewWatcherFactory(etcdClient *clientv3.Client) WatcherFactory {
	return &watcherFactoryImpl{etcdClient: etcdClient}
}

func (w *watcherFactoryImpl) WatchBytes(ctx context.Context, config WatcherConfig) (<-chan []byte, context.CancelFunc) {
	channel := make(chan []byte)
	watchCtx, cancel := context.WithCancel(ctx)

	go func() {
		defer close(channel)
		opts := []clientv3.OpOption{}
		if config.Prefix {
			opts = append(opts, clientv3.WithPrefix())
		}
		if config.StartRevision > 0 {
			opts = append(opts, clientv3.WithRev(config.StartRevision+1))
		}

		watchChannel := w.etcdClient.Watch(watchCtx, config.Key, opts...)

		for resp := range watchChannel {
			if resp.Err() != nil {
				return
			}
			for _, event := range resp.Events {
				if event.Type == config.EventType {
					channel <- event.Kv.Value
				}
			}
		}
	}()

	return channel, cancel
}

func (w *watcherFactoryImpl) WatchStrings(ctx context.Context, config WatcherConfig) (<-chan []string, context.CancelFunc) {
	channel := make(chan []string)
	watchCtx, cancel := context.WithCancel(ctx)

	go func() {
		defer close(channel)
		opts := []clientv3.OpOption{}
		if config.Prefix {
			opts = append(opts, clientv3.WithPrefix())
		}
		if config.StartRevision > 0 {
			opts = append(opts, clientv3.WithRev(config.StartRevision+1))
		}

		watchChannel := w.etcdClient.Watch(watchCtx, config.Key, opts...)

		for resp := range watchChannel {
			if resp.Err() != nil {
				return
			}
			for _, event := range resp.Events {
				if event.Type == config.EventType {
					// Transform event value to interface{}
					if config.Transform != nil {
						result, err := config.Transform(event.Kv.Value)
						if err != nil {
							continue
						}
						// Apply optional validation gate
						if config.Validate != nil && !config.Validate(result) {
							continue
						}
						// Type assert to []string
						if strSlice, ok := result.([]string); ok {
							channel <- strSlice
						}
					}
				}
			}
		}
	}()

	return channel, cancel
}

func (w *watcherFactoryImpl) WatchTimeoutEvents(ctx context.Context, config WatcherConfig) (<-chan GameTimeoutEvent, context.CancelFunc) {
	channel := make(chan GameTimeoutEvent)
	watchCtx, cancel := context.WithCancel(ctx)

	go func() {
		defer close(channel)
		opts := []clientv3.OpOption{}
		if config.Prefix {
			opts = append(opts, clientv3.WithPrefix())
		}
		if config.StartRevision > 0 {
			opts = append(opts, clientv3.WithRev(config.StartRevision+1))
		}

		watchChannel := w.etcdClient.Watch(watchCtx, config.Key, opts...)

		for resp := range watchChannel {
			if resp.Err() != nil {
				return
			}
			for _, event := range resp.Events {
				if event.Type == config.EventType {
					if config.Transform != nil {
						result, err := config.Transform(event.Kv.Key)
						if err != nil {
							continue
						}
						if config.Validate != nil && !config.Validate(result) {
							continue
						}
						// Type assert to GameTimeoutEvent
						if timeoutEvent, ok := result.(GameTimeoutEvent); ok {
							channel <- timeoutEvent
						}
					}
				}
			}
		}
	}()

	return channel, cancel
}
