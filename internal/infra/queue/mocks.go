package queue

import (
	"context"
	"fmt"
	"sync"

	"github.com/segmentio/kafka-go"
)

// MockQueue представляет общее хранилище для сообщений
type MockQueue struct {
	messages []kafka.Message
	mu       sync.Mutex
}

type MockWriter struct {
	queue *MockQueue
}

func (w *MockWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	w.queue.mu.Lock()
	defer w.queue.mu.Unlock()

	w.queue.messages = append(w.queue.messages, msgs...)
	fmt.Printf("MockWriter: Added %d messages, total in queue: %d\n", len(msgs), len(w.queue.messages))
	return nil
}

type MockReader struct {
	queue     *MockQueue
	callCount int
	mu        sync.Mutex
}

func (r *MockReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.queue.mu.Lock()
	defer r.queue.mu.Unlock()

	if r.callCount >= len(r.queue.messages) {
		fmt.Println("MockReader: No messages available, waiting...")
		return kafka.Message{}, fmt.Errorf("no messages available")
	}

	m := r.queue.messages[r.callCount]
	fmt.Printf("MockReader: Reading message %d of %d\n", r.callCount+1, len(r.queue.messages))
	r.callCount++

	return m, nil
}

func (r *MockReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	fmt.Printf("MockReader: Committed %d messages\n", len(msgs))
	return nil
}

// NewMockQueue создает новую очередь с Reader и Writer
func NewMockQueue() (*MockWriter, *MockReader) {
	mq := &MockQueue{
		messages: make([]kafka.Message, 0),
	}

	return &MockWriter{queue: mq}, &MockReader{queue: mq}
}
