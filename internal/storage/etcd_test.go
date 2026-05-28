package storage

import (
	"MarafoNet/internal/model"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func newTestEtcdService(t *testing.T) *EtcdService {
	t.Helper()
	endpoint := os.Getenv("ETCD_TEST_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:2379"
	}
	svc, err := NewEtcdService([]string{endpoint}, 3*time.Second)
	if err != nil {
		t.Skipf("etcd not available on %s: %v", endpoint, err)
		return nil
	}
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cleanupCancel()
	prefixes := []string{"test/", "users/", "user_queue/", "game_timeout/"}
	for _, p := range prefixes {
		_, _ = svc.client.Delete(cleanupCtx, p, clientv3.WithPrefix())
	}

	return svc
}

func TestGetNextGameID_Sequential(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = svc.client.Delete(ctx, svc.pathBuilder.GameCounterPath())

	id1, err := svc.GetNextGameID()
	require.NoError(t, err)
	assert.Equal(t, "game/1", id1)

	id2, err := svc.GetNextGameID()
	require.NoError(t, err)
	assert.Equal(t, "game/2", id2)

	id3, err := svc.GetNextGameID()
	require.NoError(t, err)
	assert.Equal(t, "game/3", id3)
}

func TestPutNewGame_GetValueAndRevision(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("test/game_%s", t.Name())
	expectedValue := []byte(`{"game":"example"}`)
	err := svc.PutNewGame(ctx, key, expectedValue)
	assert.NoError(t, err)

	actualValue, rev, err := svc.GetGameJsonAndRevision(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, string(expectedValue), string(actualValue))
	assert.Greater(t, rev, int64(0))
}

func TestPutNewGame_AlreadyExists(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("test/game_%s", t.Name())
	expectedValue := []byte(`{"game":"example"}`)

	err := svc.PutNewGame(ctx, key, expectedValue)
	require.NoError(t, err)

	// Try to create again - should fail
	err = svc.PutNewGame(ctx, key, expectedValue)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestPutUpdatedGameJsonIfRevisionMatch(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("test/game_%s", t.Name())
	initialValue := []byte(`{"game":"v1"}`)
	err := svc.PutNewGame(ctx, key, initialValue)
	require.NoError(t, err)

	actualValue, revision, err := svc.GetGameJsonAndRevision(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, string(initialValue), string(actualValue))

	updatedValue := []byte(`{"game":"v2"}`)
	err = svc.PutUpdatedGameJsonIfRevisionMatch(ctx, key, updatedValue, revision)
	assert.NoError(t, err)

	newValue, newRevision, err := svc.GetGameJsonAndRevision(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, string(updatedValue), string(newValue))
	assert.NotEqual(t, revision, newRevision)
}

func TestPutUpdatedGameJsonIfRevisionMatch_RevisionMismatch(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("test/game_%s", t.Name())
	initialValue := []byte(`{"game":"v1"}`)
	err := svc.PutNewGame(ctx, key, initialValue)
	require.NoError(t, err)

	updatedValue := []byte(`{"game":"v2"}`)
	err = svc.PutUpdatedGameJsonIfRevisionMatch(ctx, key, updatedValue, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "revision mismatch")
}

func TestGetGameJsonAndRevision_NotFound(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := svc.GetGameJsonAndRevision(ctx, "test/nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPutUserIntoQueue_GetUserQueue(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	player1 := fmt.Sprintf("player_%s_1", t.Name())
	player2 := fmt.Sprintf("player_%s_2", t.Name())

	err := svc.PutUserIntoQueue(ctx, player1)
	assert.NoError(t, err)

	queue, err := svc.GetUserQueue(ctx)
	assert.NoError(t, err)
	assert.Contains(t, queue, player1)

	err = svc.PutUserIntoQueue(ctx, player2)
	assert.NoError(t, err)

	queue, err = svc.GetUserQueue(ctx)
	assert.NoError(t, err)
	assert.Contains(t, queue, player1)
	assert.Contains(t, queue, player2)
}

func TestRemoveUserFromQueue(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	player := fmt.Sprintf("player_%s", t.Name())
	err := svc.PutUserIntoQueue(ctx, player)
	require.NoError(t, err)

	queue, err := svc.GetUserQueue(ctx)
	require.NoError(t, err)
	assert.Contains(t, queue, player)

	err = svc.RemoveUserFromQueue(ctx, player)
	assert.NoError(t, err)

	queue, err = svc.GetUserQueue(ctx)
	assert.NoError(t, err)
	assert.NotContains(t, queue, player)
}

func TestSetUserCurrentGameId_GetUserCurrentGameId(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	player := fmt.Sprintf("player_%s", t.Name())
	gameId := fmt.Sprintf("game/%s", t.Name())

	err := svc.SetUserCurrentGameId(ctx, player, gameId)
	assert.NoError(t, err)

	retrievedGameId, err := svc.GetUserCurrentGameId(ctx, player)
	assert.NoError(t, err)
	assert.Equal(t, gameId, retrievedGameId)
}

func TestGetUserCurrentGameId_NotSet(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	player := fmt.Sprintf("player_%s_nonexistent", t.Name())
	gameId, err := svc.GetUserCurrentGameId(ctx, player)
	assert.NoError(t, err)
	assert.Equal(t, "", gameId)
}

func TestRemoveUserCurrentGameId(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	player := fmt.Sprintf("player_%s", t.Name())
	gameId := "game/test"

	err := svc.SetUserCurrentGameId(ctx, player, gameId)
	require.NoError(t, err)

	err = svc.RemoveUserCurrentGameId(ctx, player)
	assert.NoError(t, err)

	retrievedGameId, err := svc.GetUserCurrentGameId(ctx, player)
	assert.NoError(t, err)
	assert.Equal(t, "", retrievedGameId)
}

func TestRegisterUser(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name := fmt.Sprintf("User_%s", t.Name())
	user := model.NewUser(name, "securePassword123")

	err := svc.RegisterUser(ctx, user)
	assert.NoError(t, err)

	isVerified, err := svc.verifyUser(ctx, user)
	assert.NoError(t, err)
	assert.True(t, isVerified)
}

func TestRegisterUser_Duplicate(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name := fmt.Sprintf("User_%s", t.Name())
	user := model.NewUser(name, "password123")

	err := svc.RegisterUser(ctx, user)
	require.NoError(t, err)

	// Try to register same user again
	err = svc.RegisterUser(ctx, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}

func TestVerifyUser_WrongPassword(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name := fmt.Sprintf("User_%s", t.Name())
	user := model.NewUser(name, "correctPassword")

	err := svc.RegisterUser(ctx, user)
	require.NoError(t, err)

	wrongUser := model.NewUser(name, "wrongPassword")
	isVerified, err := svc.verifyUser(ctx, wrongUser)
	assert.NoError(t, err)
	assert.False(t, isVerified)
}

func TestVerifyUser_NotRegistered(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name := fmt.Sprintf("User_%s_nonexistent", t.Name())
	user := model.NewUser(name, "password")

	isVerified, err := svc.verifyUser(ctx, user)
	assert.NoError(t, err)
	assert.False(t, isVerified)
}

func TestLoginUser_Success(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name := fmt.Sprintf("User_%s", t.Name())
	user := model.NewUser(name, "password123")

	err := svc.RegisterUser(ctx, user)
	require.NoError(t, err)

	isConnected, _ := svc.isUserConnected(ctx, user.Name)
	assert.False(t, isConnected)

	err = svc.LoginUser(ctx, user)
	assert.NoError(t, err)

	isConnected, err = svc.isUserConnected(ctx, user.Name)
	assert.NoError(t, err)
	assert.True(t, isConnected)
}

func TestLoginUser_WrongPassword(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name := fmt.Sprintf("User_%s", t.Name())
	user := model.NewUser(name, "correctPassword")

	err := svc.RegisterUser(ctx, user)
	require.NoError(t, err)

	wrongUser := model.NewUser(name, "wrongPassword")
	err = svc.LoginUser(ctx, wrongUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid password")
}

func TestLoginUser_AlreadyOnline(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name := fmt.Sprintf("User_%s", t.Name())
	user := model.NewUser(name, "password123")

	err := svc.RegisterUser(ctx, user)
	require.NoError(t, err)

	err = svc.LoginUser(ctx, user)
	require.NoError(t, err)

	// Try to login again while already online
	err = svc.LoginUser(ctx, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected state transition")
}

func TestSetUserOnlineOfflineStatus(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name := fmt.Sprintf("User_%s", t.Name())
	user := model.NewUser(name, "password")

	err := svc.RegisterUser(ctx, user)
	require.NoError(t, err)

	isOnline, _ := svc.isUserConnected(ctx, user.Name)
	assert.False(t, isOnline)

	err = svc.setUserOnlineStatus(ctx, user.Name)
	assert.NoError(t, err)

	isOnline, _ = svc.isUserConnected(ctx, user.Name)
	assert.True(t, isOnline)

	err = svc.setUserOfflineStatus(ctx, user.Name)
	assert.NoError(t, err)

	isOnline, _ = svc.isUserConnected(ctx, user.Name)
	assert.False(t, isOnline)
}

func TestOnUserDisconnect(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name := fmt.Sprintf("User_%s", t.Name())
	user := model.NewUser(name, "password")

	err := svc.RegisterUser(ctx, user)
	require.NoError(t, err)

	err = svc.LoginUser(ctx, user)
	require.NoError(t, err)

	err = svc.OnUserDisconnect(ctx, user.Name)
	assert.NoError(t, err)

	isOnline, _ := svc.isUserConnected(ctx, user.Name)
	assert.False(t, isOnline)
}

func TestWatchGame_ReceivesUpdates(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	key := fmt.Sprintf("test/watch_game_%s", t.Name())
	t.Cleanup(func() { _, _ = svc.client.Delete(context.Background(), key) })

	channel, watchCancel := svc.WatchGame(ctx, key)
	defer watchCancel()

	expectedValue := []byte(`{"watch":"update1"}`)
	err := svc.PutNewGame(ctx, key, expectedValue)
	assert.NoError(t, err)

	select {
	case actualValue := <-channel:
		assert.Equal(t, string(expectedValue), string(actualValue))
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for watch event")
	}
}

func TestWatchUserLobby(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	player := fmt.Sprintf("player_%s", t.Name())
	channel, watchCancel := svc.WatchUserLobby(ctx, player)
	defer watchCancel()

	gameId := "game/123"
	err := svc.SetUserCurrentGameId(ctx, player, gameId)
	assert.NoError(t, err)

	select {
	case update := <-channel:
		assert.Equal(t, gameId, string(update))
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for lobby update")
	}
}

func TestIsUserInAGame(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	player := fmt.Sprintf("player_%s", t.Name())

	inGame := svc.isUserInAGame(ctx, player)
	assert.False(t, inGame)

	gameId := "game/123"
	err := svc.SetUserCurrentGameId(ctx, player, gameId)
	require.NoError(t, err)

	inGame = svc.isUserInAGame(ctx, player)
	assert.True(t, inGame)
}

func TestIsUserConnected(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name := fmt.Sprintf("User_%s", t.Name())
	user := model.NewUser(name, "password")

	err := svc.RegisterUser(ctx, user)
	require.NoError(t, err)

	isConnected, err := svc.isUserConnected(ctx, user.Name)
	assert.NoError(t, err)
	assert.False(t, isConnected)

	err = svc.setUserOnlineStatus(ctx, user.Name)
	require.NoError(t, err)

	isConnected, err = svc.isUserConnected(ctx, user.Name)
	assert.NoError(t, err)
	assert.True(t, isConnected)
}

func TestContextCancellation(t *testing.T) {
	svc := newTestEtcdService(t)
	defer svc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.GetUserQueue(ctx)
	assert.Error(t, err)
}

func TestClose(t *testing.T) {
	svc := newTestEtcdService(t)
	err := svc.Close()
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Should fail after close
	_, err = svc.GetUserQueue(ctx)
	assert.Error(t, err)
}
