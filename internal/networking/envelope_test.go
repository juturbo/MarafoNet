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
	envelope := initEnvelope(JoinType)
	assert.Equal(t, JoinType, envelope.GetMessageType(), "Expected getMessageType() to equal JoinType but got %s", envelope.GetMessageType())

	envelope = initEnvelope(PlayCardType)
	assert.Equal(t, PlayCardType, envelope.GetMessageType(), "Expected getMessageType() to equal PlayCardType but got %s", envelope.GetMessageType())
}
