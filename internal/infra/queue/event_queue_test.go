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

	writer, _ := NewMockQueue()

	producer := NewKafkaEventProducer(writer)

	testEvent, err := event.NewEvent("test_event", map[string]string{"key": "value"})
	require.NoError(t, err)

	err = producer.Produce(context.Background(), testEvent)
	require.NoError(t, err)

	require.NotNil(t, writer.queue)
	require.Len(t, writer.queue.messages, 1)

	msg := writer.queue.messages[0]
	require.Equal(t, "events", msg.Topic)
	require.Equal(t, testEvent.ID.String(), string(msg.Key))

	expectedJSON, _ := json.Marshal(testEvent)
	require.JSONEq(t, string(expectedJSON), string(msg.Value))
}

func TestKafkaEventConsumer_Consume_Success(t *testing.T) {

	writer, reader := NewMockQueue()

	testEvent, err := event.NewEvent("test_event", map[string]string{"key": "value"})
	require.NoError(t, err)

	jsonData, err := json.Marshal(testEvent)
	require.NoError(t, err)

	err = writer.WriteMessages(context.Background(), kafka.Message{
		Topic: "events",
		Key:   []byte(testEvent.ID.String()),
		Value: jsonData,
	})
	require.NoError(t, err)

	consumer := NewKafkaEventConsumer(reader)

	evt, err := consumer.Consume(context.Background())
	require.NoError(t, err)

	require.Equal(t, testEvent.ID, evt.ID)
	require.Equal(t, testEvent.Name, evt.Name)
}
