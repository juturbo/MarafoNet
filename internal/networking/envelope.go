package networking

import "encoding/json"

// A type of message sent from the client.
type MessageType string

const (
	// A message sent by the client when they first connect to the server.
	FirstJoinType MessageType = "first_join"
	// A message send by the client to play a card.
	PlayCardType MessageType = "play_card"
)

type Envelope interface {
	getMessageType() MessageType
	getPayload() json.RawMessage
	EqualsType(otherType MessageType) bool
}

// A WebSocket message from the client.
type WSEnvelope struct {
	MessageType string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
}

func (e *WSEnvelope) getMessageType() MessageType {
	return MessageType(e.MessageType)
}

func (e *WSEnvelope) getPayload() json.RawMessage {
	return e.Payload
}

func (e *WSEnvelope) equalsType(otherType MessageType) bool {
	return e.getMessageType() == otherType
}
