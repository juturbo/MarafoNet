package main

import (
	"MarafoNet/internal/service"
	"MarafoNet/internal/storage"
	"MarafoNet/internal/timeoutwatcher"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var etcdEndpoint = getEtcdEndpoints()

func main() {
	printHeader()
	etcdService, err := storage.NewEtcdService(etcdEndpoint, 5*time.Second)
	if err != nil {
		log.Fatalf("failed to connect to etcd: %v", err)
	}
	defer func() {
		closeErr := etcdService.Close()
		if closeErr != nil {
			log.Printf("failed to close etcd client: %v", closeErr)
		}
	}()

	log.Printf("connected to etcd at %v", etcdEndpoint)
	log.Printf("starting timeout watcher service...")

	gameService := service.NewGameService(etcdService)
	timeoutWatcherService := timeoutwatcher.NewTimeoutWatcher(etcdService, gameService)

	var wg sync.WaitGroup
	wg.Add(1)

	timeoutWatcherService.Start(&wg)

	log.Printf("timeout watcher service is running. Press Ctrl+C to stop.")
	wg.Wait()
}

func getEtcdEndpoints() []string {
	env := os.Getenv("ETCD_ENDPOINTS")
	if env == "" {
		// Fallback for local development
		return []string{"localhost:2379"}
	}
	return strings.Split(env, ",")
}

func printHeader() {
	log.Println("====================================")
	log.Println("      Timeout Watcher Service       ")
	log.Println("====================================")
}
