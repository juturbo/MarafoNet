package networking

import "encoding/json"

// A type of message sent from the client.
type MessageType string

const (
	// A message sent by the client when they connect to the server.
	JoinType MessageType = "first_join"
	// A message send by the client to play a card.
	PlayCardType MessageType = "play_card"
	// A message sent by the client to set the trump suit.
	SetTrumpType MessageType = "set_trump"
)

type Envelope interface {
	GetMessageType() MessageType
	GetPayload() json.RawMessage
	EqualsType(otherType MessageType) bool
}

// A WebSocket message from the client.
type WSEnvelope struct {
	MessageType string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
}

// Returns the message type of the envelope.
func (e WSEnvelope) GetMessageType() MessageType {
	return MessageType(e.MessageType)
}

// Returns the payload of the envelope.
func (e WSEnvelope) GetPayload() json.RawMessage {
	return e.Payload
}

// Returns true if the message type of the envelope is equal to the given type.
func (e WSEnvelope) EqualsType(otherType MessageType) bool {
	return e.GetMessageType() == otherType
}
