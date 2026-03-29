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
	GetUUID() string
}

// A WebSocket message from the client.
type WSEnvelope struct {
	MessageType string          `json:"type"`
	PlayerName  string          `json:"playerName"`
	Payload     json.RawMessage `json:"payload"`
	UUID        string          `json:"uuid"`
}

type playCardPayLoad struct {
	MatchID string     `json:"matchId"`
	Rank    model.Rank `json:"rank"`
	Suit    model.Suit `json:"suit"`
}

type SetTrumpPayLoad struct {
	MatchID string     `json:"matchId"`
	Suit    model.Suit `json:"suit"`
}

type ReplyMessageBuilder struct {
	Type    string `json:"type"`
	UUID    string `json:"uuid,omitempty"`
	Message string `json:"message,omitempty"`
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

func (e WSEnvelope) GetUUID() string {
	return e.UUID
}

// Returns true if the message type of the envelope is equal to the given type.
func (e WSEnvelope) EqualsType(otherType MessageType) bool {
	return e.GetMessageType() == otherType
}

func PayloadFromJSON(data json.RawMessage) (string, model.Card, error) {
	var p playCardPayLoad
	err := json.Unmarshal(data, &p)
	return p.MatchID,
		model.Card{
			Rank: p.Rank,
			Suit: p.Suit,
		}, err
}

func NewReplyMessageBuilder() *ReplyMessageBuilder {
	return &ReplyMessageBuilder{}
}

func (b *ReplyMessageBuilder) SetUUID(uuid string) {
	b.UUID = uuid
}

func (b *ReplyMessageBuilder) SetMessage(message string) {
	b.Message = message
}

func (b *ReplyMessageBuilder) SetType(messageType string) {
	b.Type = messageType
}

func (b *ReplyMessageBuilder) Build() json.RawMessage {
	messageJson, _ := json.Marshal(b)
	return messageJson
}

func BuildJSONErrorResponse(errorMessage string) json.RawMessage {
	errorResponse := NewReplyMessageBuilder()
	errorResponse.SetType("error")
	errorResponse.SetMessage(errorMessage)
	return errorResponse.Build()
}
