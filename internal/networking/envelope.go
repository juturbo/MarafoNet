package envelope

import "encoding/json"

// MessageType is the type of message being sent by the client.
const (
	FirstJoinType MessageType = "first_join"
	PlayCardType  MessageType = "play_card"
)

type Envelope interface {
	getMessageType() (MessageType, error)
	getPayload() (json.RawMessage, error)
}

// A WebSocket message from the client. 
type WSEnvelope struct {
    MessageType    	string          `json:"type"`
    Payload 		json.RawMessage `json:"payload"`
}

func (e *WSEnvelope) getMessageType() (MessageType, error) {
	return MessageType(e.MessageType), nil
}

func (e *WSEnvelope) getPayload() (json.RawMessage, error) {
	return e.Payload, nil
}
