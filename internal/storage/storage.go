package storage

import (
	"MarafoNet/internal/model"
	"context"
	"encoding/json"
)

type WebSocketRepository interface {
	UserSessionStorage
	QueueStorage
	GameStorage
}

type StorageService interface {
	GameStorage
	UserSessionStorage
	QueueStorage
	WatchStorage
	Close() error
}

type GameStorage interface {
	GetNextGameID() (gameId string, err error)
	PutNewGame(ctx context.Context, key string, gameJson []byte) error
	GetGameJsonAndRevision(ctx context.Context, key string) (gameJson json.RawMessage, revision int64, err error)
	PutUpdatedGameJsonIfRevisionMatch(ctx context.Context, gameId string, gameJson json.RawMessage, lastRevision int64) error
}

type UserSessionStorage interface {
	RegisterUser(ctx context.Context, user model.User) error
	LoginUser(ctx context.Context, user model.User) error
	GetUserCurrentGameId(ctx context.Context, playerName string) (string, error)
	SetUserCurrentGameId(ctx context.Context, playerName string, gameId string) error
	RemoveUserCurrentGameId(ctx context.Context, playerName string) error
	OnUserDisconnect(ctx context.Context, playerName string) error
}

type QueueStorage interface {
	PutUserIntoQueue(ctx context.Context, playerName string) error
	GetUserQueue(ctx context.Context) (userQueue []string, err error)
	RemoveUserFromQueue(ctx context.Context, playerName string) error
}

type WatchStorage interface {
	WatchGame(ctx context.Context, gameId string) (<-chan []byte, context.CancelFunc)
	WatchUserLobby(ctx context.Context, username string) (<-chan []byte, context.CancelFunc)
	WatchUserTimeoutLease(ctx context.Context) (<-chan GameTimeoutEvent, context.CancelFunc)
	WatchUserQueue(ctx context.Context) (<-chan []string, context.CancelFunc)
}
