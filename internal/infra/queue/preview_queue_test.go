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

	writer, _ := NewMockQueue()

	producer := NewKafkaPreviewProducer(writer)

	versionID := uuid.New()

	err := producer.Produce(context.Background(), versionID)
	require.NoError(t, err)

	require.NotNil(t, writer.queue)
	require.Len(t, writer.queue.messages, 1)

	msg := writer.queue.messages[0]
	require.Equal(t, "preview", msg.Topic)

	var payload map[string]string
	err = json.Unmarshal(msg.Value, &payload)
	require.NoError(t, err)
	require.Equal(t, versionID.String(), payload["version_id"])
}

func TestKafkaPreviewConsumer_Consume(t *testing.T) {

	writer, reader := NewMockQueue()

	consumer := NewKafkaPreviewConsumer(reader)

	versionID := uuid.New()
	payload := map[string]string{"version_id": versionID.String()}
	data, _ := json.Marshal(payload)

	err := writer.WriteMessages(context.Background(), kafka.Message{
		Topic: "preview",
		Key:   []byte(versionID.String()),
		Value: data,
	})
	require.NoError(t, err)

	id, err := consumer.Consume(context.Background())
	require.NoError(t, err)
	require.Equal(t, versionID, id)

	err = consumer.Remove(context.Background(), id)
	require.NoError(t, err)
}
