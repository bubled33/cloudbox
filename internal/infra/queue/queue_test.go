package queue

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type MockWriter struct {
	messages []kafka.Message
}

func (w *MockWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	w.messages = append(w.messages, msgs...)
	return nil
}

type MockReader struct {
	messages  []kafka.Message
	callCount int
}

func (r *MockReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	m := r.messages[r.callCount]
	r.callCount += 1

	return m, nil
}

func (r *MockReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return nil
}
