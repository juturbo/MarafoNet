package networking

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type WebSocketHub struct {
	WriteChannel chan json.RawMessage
}

// Calls goroutines to serve read and write channels for one WebSocket connection.
func ServeWS(Conn *websocket.Conn) {
	var hub WebSocketHub
	hub.WriteChannel = make(chan json.RawMessage)
	go ServeWrite(Conn, hub.WriteChannel)
	go ServeRead(Conn, hub.WriteChannel)

}

func ServeWrite(Conn *websocket.Conn, writeChannel chan json.RawMessage) {
	defer Conn.Close()

	for message := range writeChannel {
		err := Conn.WriteJSON(message)
		if err != nil {
			return
		}
	}
}

func ServeRead(Conn *websocket.Conn, writeChannel chan json.RawMessage) {
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
	case envelope.EqualsType(FirstJoinType):
		panic("unimplemented")
	case envelope.EqualsType(PlayCardType):
		panic("unimplemented")
	default:
		panic("unknown envelope type")
	}
}
