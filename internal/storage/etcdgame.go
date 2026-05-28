package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	concurrency "go.etcd.io/etcd/client/v3/concurrency"
)

const DEFAULT_MAX_RETRIES = 3
const DEFAULT_INITIAL_BACKOFF = 100 * time.Millisecond
const DEFAULT_MAX_BACKOFF = 10 * time.Second
const DEFAULT_BACKOFF_MULTIPLIER = 2.0

type etcdGameRepositoryService struct {
	core *etcdCore
}

func (service *etcdGameRepositoryService) GetNextGameID() (gameId string, err error) {
	ctx := context.Background()

	return gameId, service.retryWithBackoff(ctx, "GetNextGameID", func() error {
		key := service.core.pathBuilder.GameCounterPath()

		var next int
		transactionResponse, err := concurrency.NewSTM(service.core.client, func(stm concurrency.STM) error {
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

		gameId = service.core.pathBuilder.GamePath(strconv.Itoa(next))
		return nil
	})
}

func (service *etcdGameRepositoryService) PutNewGame(ctx context.Context, key string, gameJson []byte) error {
	succeeded, err := service.core.putIfComparison(
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

func (service *etcdGameRepositoryService) GetGameJsonAndRevision(ctx context.Context, key string) (gameJson json.RawMessage, revision int64, err error) {
	value, revision, err := service.core.getKeyValue(ctx, key)
	if err != nil {
		return nil, 0, err
	}

	if value == "" {
		return nil, 0, fmt.Errorf("key not found: %s", key)
	}

	return json.RawMessage(value), revision, nil
}

func (service *etcdGameRepositoryService) PutUpdatedGameJsonIfRevisionMatch(ctx context.Context, gameId string, gameJson json.RawMessage, lastRevision int64) error {
	succeeded, err := service.core.putIfComparison(
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

func (service *etcdGameRepositoryService) Close() error {
	return service.core.Close()
}

func (service *etcdGameRepositoryService) retryWithBackoff(
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
