package networking

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func initEnvelope(messageType MessageType) *WSEnvelope {
	return &WSEnvelope{
		MessageType: string(messageType),
		Payload:     []byte("this is a test payload"),
	}
}

func TestEqualsMessageType(t *testing.T) {
	envelope := initEnvelope(FirstJoinType)
	assert.Equal(t, FirstJoinType, envelope.GetMessageType(), "Expected getMessageType() to equal FirstJoinType but got %s", envelope.GetMessageType())

	envelope = initEnvelope(PlayCardType)
	assert.Equal(t, PlayCardType, envelope.GetMessageType(), "Expected getMessageType() to equal PlayCardType but got %s", envelope.GetMessageType())
}
