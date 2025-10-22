package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type KafkaExpirerProducer struct {
	writer KafkaWriter
}

func NewKafkaExpirerProducer(writer KafkaWriter) *KafkaExpirerProducer {
	return &KafkaExpirerProducer{writer: writer}
}

func (p *KafkaExpirerProducer) Produce(ctx context.Context, linkID uuid.UUID, duration time.Duration) error {
	payload := map[string]interface{}{
		"link_id":  linkID.String(),
		"duration": duration.Seconds(),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Topic: "expirer",
		Key:   []byte(linkID.String()),
		Value: data,
	}
	return p.writer.WriteMessages(ctx, msg)
}

type KafkaExpirerConsumer struct {
	reader KafkaReader
}

func NewKafkaExpirerConsumer(reader KafkaReader) *KafkaExpirerConsumer {
	return &KafkaExpirerConsumer{reader: reader}
}

func (c *KafkaExpirerConsumer) Consume(ctx context.Context) (uuid.UUID, time.Duration, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return uuid.Nil, 0, err
	}

	var payload struct {
		LinkID   string  `json:"link_id"`
		Duration float64 `json:"duration"`
	}
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		return uuid.Nil, 0, err
	}

	id, err := uuid.Parse(payload.LinkID)
	if err != nil {
		return uuid.Nil, 0, err
	}

	return id, time.Duration(payload.Duration * float64(time.Second)), nil
}

func (c *KafkaExpirerConsumer) Remove(ctx context.Context, linkID uuid.UUID) error {
	return c.reader.CommitMessages(ctx, kafka.Message{Key: []byte(linkID.String())})
}
