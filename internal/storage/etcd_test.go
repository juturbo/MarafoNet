//go:build integration

package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEtcdConnection(t *testing.T) {
	cli, err := NewGameRepository([]string{"localhost:2379"})
	if err != nil {
		t.Fatalf("Failed to connect to Etcd: %v", err)
	}
	defer cli.client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := "/test/connection"
	value := "works!"

	_, err = cli.client.Put(ctx, key, value)
	assert.NoError(t, err, "Failed to write to Etcd")

	resp, err := cli.client.Get(ctx, key)
	assert.NoError(t, err, "Failed to read from Etcd")

	if len(resp.Kvs) == 0 {
		t.Errorf("Expected key %s to exist, but got 0 results", key)
	} else {
		actualValue := string(resp.Kvs[0].Value)
		assert.Equal(t, value, actualValue, "The value retrieved should match the value saved")
	}

	cli.client.Delete(ctx, key)
}
