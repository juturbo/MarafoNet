package main

import (
	"MarafoNet/internal/matchmaking"
	"MarafoNet/internal/service"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const rootPath string = "/"

// FOR LOCAL DEVELOPMENT ONLY: CHANGE WHEN CREATING DOCKER IMAGE
const indexPath string = "./frontend/build/index.html"
const localFilePath string = "./frontend/build"

const webSocketPath = "/ws"

var etcdEndpoint = []string{"localhost:2379"}

func main() {
	printHeader()
	etcdService, err := service.NewEtcdService(etcdEndpoint, 5*time.Second)
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
	wg.Add(2)

	matchMakingService.StartMatchmaking(&wg)
	matchMakingService.StartTimeoutWatcher(&wg)

	wg.Wait()

}

func printHeader() {
	log.Println("====================================")
	log.Println("        Matchmaking Server          ")
	log.Println("====================================")
}
