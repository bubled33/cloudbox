package queue

import (
	"context"
	"testing"

	"github.com/segmentio/kafka-go"
)

type MockWriter struct {
	messages []kafka.Message
}

func (w *MockWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	w.messages = msgs
	return nil
}

func TestKafkaEventProducer_Produce_Success(t *testing.T) {
	mockWriter := MockWriter{}
	producer := NewKafkaEventProducer(mockWriter)
	err := producer.Produce(ctx, testEvent)
}
