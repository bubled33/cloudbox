package queue

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
)

func TestKafkaExpirerProducer_Produce(t *testing.T) {
	mockWriter := &MockWriter{}
	producer := NewKafkaExpirerProducer(mockWriter)

	linkID := uuid.New()
	duration := 10 * time.Second

	err := producer.Produce(context.Background(), linkID, duration)
	require.NoError(t, err)

	require.Len(t, mockWriter.messages, 1)
	msg := mockWriter.messages[0]

	require.Equal(t, "expirer", msg.Topic)

	var payload map[string]interface{}
	err = json.Unmarshal(msg.Value, &payload)
	require.NoError(t, err)
	require.Equal(t, linkID.String(), payload["link_id"])
	require.Equal(t, duration.Seconds(), payload["duration"])
}

func TestKafkaExpirerConsumer_Consume(t *testing.T) {
	linkID := uuid.New()
	duration := 15 * time.Second

	payload := map[string]interface{}{
		"link_id":  linkID.String(),
		"duration": duration.Seconds(),
	}
	data, _ := json.Marshal(payload)

	mockReader := &MockReader{
		messages: []kafka.Message{
			{Topic: "expirer", Key: []byte(linkID.String()), Value: data},
		},
	}

	consumer := NewKafkaExpirerConsumer(mockReader)

	id, dur, err := consumer.Consume(context.Background())
	require.NoError(t, err)
	require.Equal(t, linkID, id)
	require.Equal(t, duration, dur)

	// Remove (Commit) просто проверяем, что не падает
	err = consumer.Remove(context.Background(), id)
	require.NoError(t, err)
}
