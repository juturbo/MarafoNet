package main

import (
	"MarafoNet/internal/matchmaking"
	"MarafoNet/internal/service"
	"MarafoNet/internal/storage"
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
	log.Printf("starting matchmaking services...")

	gameService := service.NewGameService(etcdService)
	matchMakingService := matchmaking.NewMatchmakingHub(etcdService, gameService)

	var wg sync.WaitGroup
	wg.Add(1)

	matchMakingService.StartMatchmaking(&wg)

	wg.Wait()

}

func getEtcdEndpoints() []string {
	env := os.Getenv("ETCD_ENDPOINTS")
	if env == "" {
		// Fallback per 'make test' e 'make dev' in locale
		return []string{"localhost:2379"}
	}
	return strings.Split(env, ",")
}

func printHeader() {
	log.Println("====================================")
	log.Println("        Matchmaking Server          ")
	log.Println("====================================")
}
