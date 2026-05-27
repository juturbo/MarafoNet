package repository

import (
	"MarafoNet/internal/model"
	"context"
	"encoding/json"
)

type GameServicer interface {
	StartGame(ctx context.Context, playerNames []string) (gameId string, err error)
	IsGameEnded(gameJson []byte) (bool, error)
	GetGameView(gameJson []byte, playerName string) (gameViewJson []byte, err error)
	ForfeitGame(ctx context.Context, gameId string, playerName string) error
	SetTrumpSuit(ctx context.Context, gameId string, playerName string, suit model.Suit) error
	PlayCard(ctx context.Context, gameId string, playerName string, card model.Card) error
}

type GameRepository interface {
	GetNextGameID() (gameId string, err error)
	PutNewGame(ctx context.Context, key string, gameJson []byte) error
	GetGameJsonAndRevision(ctx context.Context, key string) (gameJson json.RawMessage, revision int64, err error)
	PutUpdatedGameJsonIfRevisionMatch(ctx context.Context, gameId string, gameJson json.RawMessage, lastRevision int64) error
}

type UserRepository interface {
	RegisterUser(ctx context.Context, user model.User) error
	LoginUser(ctx context.Context, user model.User) error
	GetUserCurrentGameId(ctx context.Context, playerName string) (string, error)
	SetUserCurrentGameId(ctx context.Context, playerName string, gameId string) error
	RemoveUserCurrentGameId(ctx context.Context, playerName string) error
	OnUserDisconnect(ctx context.Context, playerName string) error
}

type QueueRepository interface {
	PutUserIntoQueue(ctx context.Context, playerName string) error
	GetUserQueue(ctx context.Context) (userQueue []string, err error)
	RemoveUserFromQueue(ctx context.Context, playerName string) error
	WatchUserQueue(ctx context.Context) (<-chan []string, context.CancelFunc)
}

type GameWatcherRepository interface {
	WatchGame(ctx context.Context, gameId string) (<-chan []byte, context.CancelFunc)
	WatchUserLobby(ctx context.Context, username string) (<-chan []byte, context.CancelFunc)
	WatchUserTimeoutLease(ctx context.Context) (<-chan GameTimeoutEvent, context.CancelFunc)
}

type SessionRepository interface {
	RemoveUserFromQueue(ctx context.Context, playerName string) error
	OnUserDisconnect(ctx context.Context, playerName string) error
}

type WebSocketRepository interface {
	UserRepository
	QueueRepository
	GameRepository
}

type StorageServicer interface {
	GameRepository
	UserRepository
	QueueRepository
	GameWatcherRepository
	Close() error
}
