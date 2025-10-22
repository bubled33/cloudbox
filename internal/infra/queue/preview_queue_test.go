package queue

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
)

func TestKafkaPreviewProducer_Produce(t *testing.T) {
	mockWriter := &MockWriter{}
	producer := NewKafkaPreviewProducer(mockWriter)

	versionID := uuid.New()
	err := producer.Produce(context.Background(), versionID)
	require.NoError(t, err)

	require.Len(t, mockWriter.messages, 1)
	msg := mockWriter.messages[0]

	require.Equal(t, "preview", msg.Topic)

	var payload map[string]string
	err = json.Unmarshal(msg.Value, &payload)
	require.NoError(t, err)
	require.Equal(t, versionID.String(), payload["version_id"])
}

func TestKafkaPreviewConsumer_Consume(t *testing.T) {
	versionID := uuid.New()
	payload := map[string]string{"version_id": versionID.String()}
	data, _ := json.Marshal(payload)

	mockReader := &MockReader{
		messages: []kafka.Message{
			{Topic: "preview", Key: []byte(versionID.String()), Value: data},
		},
	}

	consumer := NewKafkaPreviewConsumer(mockReader)

	id, err := consumer.Consume(context.Background())
	require.NoError(t, err)
	require.Equal(t, versionID, id)

	err = consumer.Remove(context.Background(), id)
	require.NoError(t, err)
}
