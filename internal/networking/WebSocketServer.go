package networking

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	writeChannel chan json.RawMessage
}

// Calls goroutines to serve read and write channels for one WebSocket connection.
func ServeWS(Conn *websocket.Conn) {
	go ServeWrite()
	go ServeRead(Conn)

}

func ServeWrite() {

}

func ServeRead(Conn *websocket.Conn) {
	defer Conn.Close()

	for {
		var envelope WSEnvelope
		err := Conn.ReadJSON(&envelope)
		if err != nil {
			break
		}
		HandleWSEnvelope(envelope)
	}
}

func HandleWSEnvelope(envelope Envelope) {
	switch {
	case envelope.equalsType(FirstJoinType):
		panic("unimplemented")
	case envelope.equalsType(PlayCardType):
		panic("unimplemented")
	default:
		panic("unknown envelope type")
	}
}
