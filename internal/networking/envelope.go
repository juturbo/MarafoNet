package networking

import (
	"MarafoNet/internal/model"
	userModel "MarafoNet/model"
	"encoding/json"
)

// A type of message sent from the client.
type MessageType string

const (
	// A message sent by the client when they register.
	RegisterType MessageType = "register_user"
	// A message sent by the client when they log in.
	LoginType MessageType = "login_user"
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
	GetUser() userModel.User
	GetPlayerName() string
	GetPassword() string
	EqualsType(otherType MessageType) bool
}

// A WebSocket message from the client.
type WSEnvelope struct {
	MessageType string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	User        userModel.User  `json:"user"`
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

func (e WSEnvelope) GetUser() userModel.User {
	return e.User
}

func (e WSEnvelope) GetPlayerName() string {
	return e.User.Name
}

func (e WSEnvelope) GetPassword() string {
	return e.User.Password
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
