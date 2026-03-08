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
		var response, payload = HandleWSEnvelope(envelope)
		if response {
			writeChannel <- payload
		}
	}
}

func HandleWSEnvelope(envelope Envelope) (bool, json.RawMessage) {
	switch {
	case envelope.EqualsType(JoinType):
		panic("unimplemented")
	case envelope.EqualsType(PlayCardType):
		panic("unimplemented")
	case envelope.EqualsType(SetTrumpType):
		panic("unimplemented")
	default:
		panic("unknown envelope type")
	}
}
