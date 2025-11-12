package queue

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type KafkaPreviewProducer struct {
	writer KafkaWriter
}

func NewKafkaPreviewProducer(writer KafkaWriter) *KafkaPreviewProducer {
	return &KafkaPreviewProducer{writer: writer}
}

func (p *KafkaPreviewProducer) Produce(ctx context.Context, versionID uuid.UUID) error {
	payload := map[string]string{"version_id": versionID.String()}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Topic: "preview",
		Key:   []byte(versionID.String()),
		Value: data,
	}
	return p.writer.WriteMessages(ctx, msg)
}

type KafkaPreviewConsumer struct {
	reader KafkaReader
}

func NewKafkaPreviewConsumer(reader KafkaReader) *KafkaPreviewConsumer {
	return &KafkaPreviewConsumer{reader: reader}
}

func (c *KafkaPreviewConsumer) Consume(ctx context.Context) (uuid.UUID, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	var payload struct {
		VersionID string `json:"version_id"`
	}
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.Parse(payload.VersionID)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (c *KafkaPreviewConsumer) Remove(ctx context.Context, versionID uuid.UUID) error {
	return c.reader.CommitMessages(ctx, kafka.Message{Key: []byte(versionID.String())})
}
