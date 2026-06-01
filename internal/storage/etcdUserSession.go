package storage

import (
	"MarafoNet/internal/model"
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const KEEP_ALIVE_TTL = 120

type etcdUserSessionService struct {
	core *etcdCore
}

func (service *etcdUserSessionService) PutUserIntoQueue(ctx context.Context, playerName string) error {
	key := service.core.pathBuilder.UserQueuePath(playerName)
	return service.core.putValue(ctx, key, playerName)
}

func (service *etcdUserSessionService) GetUserQueue(ctx context.Context) (userQueue []string, err error) {
	key := service.core.pathBuilder.UserQueuePrefix()

	response, err := service.core.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	for _, kv := range response.Kvs {
		userQueue = append(userQueue, string(kv.Value))
	}

	return userQueue, nil
}

func (service *etcdUserSessionService) RemoveUserFromQueue(ctx context.Context, playerName string) error {
	key := service.core.pathBuilder.UserQueuePath(playerName)
	return service.core.deleteKey(ctx, key)
}

func (service *etcdUserSessionService) SetUserCurrentGameId(ctx context.Context, playerName string, gameId string) error {
	key := service.core.pathBuilder.UserCurrentGamePath(playerName)
	return service.core.putValue(ctx, key, gameId)
}

func (service *etcdUserSessionService) GetUserCurrentGameId(ctx context.Context, playerName string) (string, error) {
	key := service.core.pathBuilder.UserCurrentGamePath(playerName)
	return service.core.getValue(ctx, key)
}

func (service *etcdUserSessionService) RemoveUserCurrentGameId(ctx context.Context, playerName string) error {
	key := service.core.pathBuilder.UserCurrentGamePath(playerName)
	return service.core.deleteKey(ctx, key)
}

func (service *etcdUserSessionService) RegisterUser(ctx context.Context, user model.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	userKey := service.core.pathBuilder.UserPath(user.Name)
	passwordKey := service.core.pathBuilder.UserPasswordPath(user.Name)
	isConnectedKey := service.core.pathBuilder.UserConnectionPath(user.Name)
	hashedPassword, err := user.GeneratePasswordHash()
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	transaction := service.core.client.Txn(ctx).
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

func (service *etcdUserSessionService) LoginUser(ctx context.Context, user model.User) error {
	isValid, err := service.verifyUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}
	if !isValid {
		return fmt.Errorf("invalid password for user: %s", user.Name)
	}

	if err = service.setUserOnlineStatus(ctx, user.Name); err != nil {
		return err
	}

	if service.isUserInAGame(ctx, user.Name) {
		return service.removeUserTimeout(ctx, user.Name)
	}

	return nil
}

func (service *etcdUserSessionService) OnUserDisconnect(ctx context.Context, playerName string) error {
	err := service.setUserOfflineStatus(ctx, playerName)
	if err != nil {
		return err
	}

	if service.isUserInAGame(ctx, playerName) {
		if err := service.setUserGameTimeout(ctx, playerName); err != nil {
			return err
		}
	}

	return nil
}

func (service *etcdUserSessionService) verifyUser(ctx context.Context, user model.User) (bool, error) {
	passwordKey := service.core.pathBuilder.UserPasswordPath(user.Name)

	hashedPassword, err := service.core.getValue(ctx, passwordKey)
	if err != nil {
		return false, err
	}
	if hashedPassword == "" {
		return false, nil
	}
	return user.CheckPassword(hashedPassword), nil
}

func (service *etcdUserSessionService) updateUserConnectionStatus(ctx context.Context, playerName string, expect string, newState string) error {
	isConnectedKey := service.core.pathBuilder.UserConnectionPath(playerName)

	txn := service.core.client.Txn(ctx).
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

func (service *etcdUserSessionService) setUserOnlineStatus(ctx context.Context, playerName string) error {
	return service.updateUserConnectionStatus(ctx, playerName, IS_OFFLINE, IS_ONLINE)
}

func (service *etcdUserSessionService) setUserOfflineStatus(ctx context.Context, playerName string) error {
	return service.updateUserConnectionStatus(ctx, playerName, IS_ONLINE, IS_OFFLINE)
}

func (service *etcdUserSessionService) setUserGameTimeout(ctx context.Context, playerName string) error {
	gameId, err := service.GetUserCurrentGameId(ctx, playerName)
	if err != nil {
		return fmt.Errorf("failed to get user's current game ID: %w", err)
	}

	lease, err := service.core.client.Grant(ctx, KEEP_ALIVE_TTL)
	if err != nil {
		return fmt.Errorf("failed to grant lease: %w", err)
	}

	gameTimeoutKey := service.core.pathBuilder.GameTimeoutPath(gameId, playerName)
	_, err = service.core.client.Put(ctx, gameTimeoutKey, TIMEOUT, clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("failed to set game timeout: %w", err)
	}
	return nil
}

func (service *etcdUserSessionService) removeUserTimeout(ctx context.Context, playerName string) error {
	gameId, err := service.GetUserCurrentGameId(ctx, playerName)
	if err != nil {
		return fmt.Errorf("failed to get user's current game ID: %w", err)
	}

	gameTimeoutKey := service.core.pathBuilder.GameTimeoutPath(gameId, playerName)

	return service.core.deleteKey(ctx, gameTimeoutKey)
}

func (service *etcdUserSessionService) isUserInAGame(ctx context.Context, playerName string) bool {
	gameId, err := service.GetUserCurrentGameId(ctx, playerName)
	if err != nil {
		return false
	}
	return gameId != ""
}

func (service *etcdUserSessionService) isUserConnected(ctx context.Context, playerName string) (bool, error) {
	key := service.core.pathBuilder.UserConnectionPath(playerName)
	value, err := service.core.getValue(ctx, key)
	if err != nil {
		return false, err
	}
	return value == IS_ONLINE, nil
}
