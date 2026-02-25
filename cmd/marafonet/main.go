package main

import (
	"net/http"

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
const indexPath string = "./frontend/build/index.html"
const localFilePath string = "./frontend/build"

const webSocketPath = "/ws"

func main() {
	// Set the router as the default one shipped with Gin
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
	// Start and run the server
	router.Run(":5000")
}
