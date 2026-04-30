package main

import (
	"MarafoNet/internal/matchmaking"
	"MarafoNet/internal/networking"
	"MarafoNet/internal/service"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
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
const indexPath string = "./frontend/index.html"
const localFilePath string = "./frontend"

const webSocketPath = "/ws"

var etcdEndpoint = []string{"local-etcd-service:2379"}

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
	log.Printf("starting game services...")

	gameService := service.NewGameService(etcdService)
	matchMakingService := matchmaking.NewMatchmakingHub(etcdService, gameService)

	log.Printf("configuring routing...")
	// Set the router as the default one shipped with Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Serve frontend static files
	router.Use(static.Serve(rootPath, static.LocalFile(localFilePath, true)))

	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == webSocketPath {
			c.String(404, "Not Found")
		} else {
			c.File(indexPath)
		}
	})

	router.GET(webSocketPath, func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.String(400, "websocket upgrade failed")
			return
		}
		networking.ServeWS(conn, gameService, etcdService, matchMakingService)
	})

	log.Printf("starting server on port 5000")
	if err := router.Run(":5000"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func printHeader() {
	log.Println("====================================")
	log.Println("          Marafonet Server          ")
	log.Println("====================================")
}
