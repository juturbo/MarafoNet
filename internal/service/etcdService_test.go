package service

import (
	"MarafoNet/model"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// helper: crea EtcdService per i test o skip se etcd non raggiungibile
func newTestEtcdService(t *testing.T) *EtcdService {
	t.Helper()
	endpoint := os.Getenv("ETCD_TEST_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:2379"
	}
	svc, err := NewEtcdService([]string{endpoint}, 3*time.Second)
	if err != nil {
		t.Skipf("etcd non disponibile su %s: %v", endpoint, err)
		return nil
	}
	// Pulizia leggera per servizio: rimuove chiavi residue usate dai test
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cleanupCancel()
	prefixes := []string{"test/", "users/", "user_queue/"}
	for _, p := range prefixes {
		_, _ = svc.client.Delete(cleanupCtx, p, clientv3.WithPrefix())
	}

	return svc
}

func TestPutNewGame_GetValueAndRevision_PutIfUnmodified(t *testing.T) {
	svc := newTestEtcdService(t)
	if svc == nil {
		return
	}
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("test/%s", t.Name())
	expectedValue := []byte(`{"game":"example"}`)
	err := svc.PutNewGame(ctx, key, expectedValue)
	assert.NoError(t, err)

	actualValue, rev, err := svc.GetGameJsonAndRevision(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, string(expectedValue), string(actualValue))

	updatedValue := []byte(`{"game":"exampleModified"}`)
	err = svc.PutUpdatedGameJsonIfRevisionMatch(ctx, key, updatedValue, rev)
	assert.NoError(t, err)

	actualUpdateValue, _, err := svc.GetGameJsonAndRevision(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, string(updatedValue), string(actualUpdateValue))

	// cleanup handled by t.Cleanup
}

func TestSetAndGetUserCurrentGame(t *testing.T) {
	svc := newTestEtcdService(t)
	if svc == nil {
		return
	}
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	player := fmt.Sprintf("player_%s", t.Name())

	gameId := fmt.Sprintf("test/game-%s", t.Name())

	assert.NoError(t, svc.SetUserCurrentGameId(ctx, player, gameId))

	got, err := svc.GetUserCurrentGameId(ctx, player)
	assert.NoError(t, err)
	assert.Equal(t, gameId, got)
}

func TestWatchGameReceivesUpdates(t *testing.T) {
	svc := newTestEtcdService(t)
	if svc == nil {
		return
	}
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	key := fmt.Sprintf("test/watch/%s", t.Name())
	t.Cleanup(func() { _, _ = svc.client.Delete(context.Background(), key) })
	channel, watchCancel := svc.WatchGame(ctx, key)
	defer watchCancel()

	expectedValue := []byte(`{"watch":"one"}`)
	err := svc.PutNewGame(ctx, key, expectedValue)
	assert.NoError(t, err)
	select {
	case actualValue := <-channel:
		assert.Equal(t, string(expectedValue), string(actualValue))
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for watch event")
	}

	_, rev, err := svc.GetGameJsonAndRevision(ctx, key)
	if err != nil {
		t.Fatalf("GetValueAndRevision error: %v", err)
	}
	expectedValue = []byte(`{"watch":"two"}`)
	err = svc.PutUpdatedGameJsonIfRevisionMatch(ctx, key, expectedValue, rev)
	assert.NoError(t, err)
	select {
	case actualValue := <-channel:
		assert.Equal(t, string(expectedValue), string(actualValue))
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for watch update event")
	}
	// cleanup handled by t.Cleanup
}

func TestUserQueue(t *testing.T) {
	svc := newTestEtcdService(t)
	if svc == nil {
		return
	}
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	player := fmt.Sprintf("player_%s", t.Name())

	err := svc.PutUserIntoQueue(ctx, player)
	assert.NoError(t, err)

	queue, err := svc.GetUserQueue(ctx)
	assert.NoError(t, err)
	assert.Contains(t, queue, player)

	err = svc.RemoveUserFromQueue(ctx, player)
	assert.NoError(t, err)

	queue, err = svc.GetUserQueue(ctx)
	assert.NotContains(t, queue, player)
}

func TestUserRegistration(t *testing.T) {
	svc := newTestEtcdService(t)
	if svc == nil {
		return
	}
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name := fmt.Sprintf("Test_%s", t.Name())
	user := model.NewUser(name, "1234")

	isVerified, err := svc.verifyUser(ctx, user)
	assert.NoError(t, err)
	assert.False(t, isVerified)

	err = svc.RegisterUser(ctx, user)
	assert.NoError(t, err)

	isVerified, err = svc.verifyUser(ctx, user)
	assert.NoError(t, err)
	assert.True(t, isVerified)
}

func TestUserLogin(t *testing.T) {
	svc := newTestEtcdService(t)
	if svc == nil {
		return
	}
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	name := fmt.Sprintf("Test_%s", t.Name())
	user := model.NewUser(name, "1234")

	assert.NoError(t, svc.RegisterUser(ctx, user))

	isConnected, _ := svc.isUserConnected(ctx, user.Name)
	assert.False(t, isConnected)

	err := svc.LoginUser(ctx, user)
	assert.NoError(t, err)

	isConnected, err = svc.isUserConnected(ctx, user.Name)
	assert.True(t, isConnected)

	err = svc.LoginUser(ctx, user)
	assert.Error(t, err)
}
