package main

import (
	"encoding/json"
	"net/http"
	"sync"
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

type Payload struct {
	Seed  string `json:"seed"`
	Power int    `json:"power"`
	Table string `json:"table"`
}

const webPagePath string = "../../client/build/index.html"

func main() {
	// Set the router as the default one shipped with Gin
	router := gin.Default()

	// Serve frontend static files
	router.Use(static.Serve("/", static.LocalFile(webPagePath, true)))

	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.String(400, "websocket upgrade failed")
			return
		}
		defer conn.Close()

		done := make(chan struct{})
		defer close(done)

		var writeMu sync.Mutex
		writeText := func(message string) error {
			writeMu.Lock()
			defer writeMu.Unlock()
			return conn.WriteMessage(websocket.TextMessage, []byte(message))
		}
		writeJSON := func(payload Payload) error {
			writeMu.Lock()
			defer writeMu.Unlock()
			jsonBytes, err := json.Marshal(payload)
			if err != nil {
				return err
			}
			return conn.WriteMessage(websocket.TextMessage, jsonBytes)
		}

		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					if err := writeText("Hello, WebSocket!"); err != nil {
						return
					}
				}
			}
		}()

		for {
			conn.ReadMessage()
			if err := writeJSON(Payload{Seed: "Coppe", Power: 2, Table: "asso di denari"}); err != nil {
				return
			}
		}
	})
	// Start and run the server
	router.Run(":5000")
}
