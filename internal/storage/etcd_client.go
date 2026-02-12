package storage

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type GameRepository struct {
	client *clientv3.Client
}

func NewGameRepository(endpoints []string) (*GameRepository, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	return &GameRepository{client: cli}, err
}
