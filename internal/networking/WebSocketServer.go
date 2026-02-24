package WebSocketServer

import (
	"envelope"
)

// Calls goroutines to serve read and write channels for one WebSocket connection.
func ServeWS() {
	go ServeWrite()
	go ServeRead()

}

func ServeWrite() {

}

func ServeRead() {

}