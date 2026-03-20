package networking

import (
	"MarafoNet/internal/model"
	"encoding/json"
)

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
	GetPlayerName() string
	EqualsType(otherType MessageType) bool
}

// A WebSocket message from the client.
type WSEnvelope struct {
	MessageType string          `json:"type"`
	PlayerName  string          `json:"playerName"`
	Payload     json.RawMessage `json:"payload"`
}

type PlayCardPayLoad struct {
	MatchID string     `json:"matchId"`
	Rank    model.Rank `json:"rank"`
	Suit    model.Suit `json:"suit"`
}

type SetTrumpPayLoad struct {
	MatchID string     `json:"matchId"`
	Suit    model.Suit `json:"suit"`
}

// Returns the message type of the envelope.
func (e WSEnvelope) GetMessageType() MessageType {
	return MessageType(e.MessageType)
}

// Returns the payload of the envelope.
func (e WSEnvelope) GetPayload() json.RawMessage {
	return e.Payload
}

func (e WSEnvelope) GetPlayerName() string {
	return e.PlayerName
}

// Returns true if the message type of the envelope is equal to the given type.
func (e WSEnvelope) EqualsType(otherType MessageType) bool {
	return e.GetMessageType() == otherType
}

func BuildJSONErrorResponse(errorMessage string) json.RawMessage {
	errorResponse := map[string]string{
		"type":    "error",
		"message": errorMessage,
	}
	errorResponseJson, _ := json.Marshal(errorResponse)
	return errorResponseJson
}
