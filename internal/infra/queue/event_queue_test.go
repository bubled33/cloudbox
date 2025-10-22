package queue

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/cloud-file-storage/internal/domain/event"
)

func TestKafkaEventProducer_Produce_Success(t *testing.T) {
	mockWriter := &MockWriter{}
	producer := NewKafkaEventProducer(mockWriter)

	testEvent, err := event.NewEvent("test_event", map[string]string{"key": "value"})
	require.NoError(t, err)

	err = producer.Produce(context.Background(), testEvent)
	require.NoError(t, err)

	require.Len(t, mockWriter.messages, 1)
	msg := mockWriter.messages[0]

	require.Equal(t, "events", msg.Topic)
	require.Equal(t, testEvent.ID.String(), string(msg.Key))

	expectedJSON, _ := json.Marshal(testEvent)
	require.JSONEq(t, string(expectedJSON), string(msg.Value))
}

func TestKafkaEventConsumer_Consume_Success(t *testing.T) {
	testEvent, err := event.NewEvent("test_event", map[string]string{"key": "value"})
	require.NoError(t, err)

	jsonData, err := json.Marshal(testEvent)
	require.NoError(t, err)

	msg := kafka.Message{
		Topic: "events",
		Key:   []byte(testEvent.ID.String()),
		Value: jsonData,
	}

	mockReader := &MockReader{
		messages:  []kafka.Message{msg},
		callCount: 0,
	}

	consumer := NewKafkaEventConsumer(mockReader)

	evt, err := consumer.Consume(context.Background())
	require.NoError(t, err)

	require.Equal(t, testEvent.ID, evt.ID)
	require.Equal(t, testEvent.Name, evt.Name)
}
