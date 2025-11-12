package queue

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"github.com/yourusername/cloud-file-storage/internal/domain/event"
)

type KafkaEventProducer struct {
	writer KafkaWriter
}

func NewKafkaEventProducer(writer KafkaWriter) *KafkaEventProducer {
	return &KafkaEventProducer{writer: writer}
}

func (p *KafkaEventProducer) Produce(ctx context.Context, e *event.Event) error {
	jsonData, err := json.Marshal(e)
	if err != nil {
		return err
	}
	message := kafka.Message{Topic: "events", Key: []byte(e.ID.String()), Value: []byte(jsonData)}
	err = p.writer.WriteMessages(ctx, message)
	return err
}

type KafkaEventConsumer struct {
	reader KafkaReader
}

func NewKafkaEventConsumer(reader KafkaReader) *KafkaEventConsumer {
	return &KafkaEventConsumer{reader: reader}
}

func (p *KafkaEventConsumer) Consume(ctx context.Context) (*event.Event, error) {
	data, err := p.reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}
	e := event.Event{}
	err = json.Unmarshal(data.Value, &e)
	if err != nil {
		return nil, err
	}

	return &e, nil
}
