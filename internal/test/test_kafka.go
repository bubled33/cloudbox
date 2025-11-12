package test

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/testcontainers/testcontainers-go/modules/redpanda"
)

type TestKafka struct {
	container *redpanda.Container
	brokers   []string
}

func SetupTestKafka(ctx context.Context) (*TestKafka, error) {
	redpandaContainer, err := redpanda.Run(ctx,
		"docker.redpanda.com/redpandadata/redpanda:v23.3.3",
		redpanda.WithAutoCreateTopics(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start redpanda: %w", err)
	}

	brokers, err := redpandaContainer.KafkaSeedBroker(ctx)
	if err != nil {
		return nil, err
	}

	return &TestKafka{
		container: redpandaContainer,
		brokers:   []string{brokers},
	}, nil
}

func (tk *TestKafka) GetBrokers() []string {
	return tk.brokers
}

func (tk *TestKafka) CreateTopic(ctx context.Context, topic string) error {
	conn, err := kafka.Dial("tcp", tk.brokers[0])
	if err != nil {
		return fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return fmt.Errorf("failed to dial controller: %w", err)
	}
	defer controllerConn.Close()

	err = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}

func (tk *TestKafka) NewWriter(topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(tk.brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func (tk *TestKafka) NewReader(topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     tk.brokers,
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    1, // Читать даже маленькие сообщения
		MaxBytes:    10e6,
		MaxWait:     500 * time.Millisecond, // Не ждать долго
		StartOffset: kafka.FirstOffset,      // Читать с начала
	})
}

func (tk *TestKafka) Terminate(ctx context.Context) error {
	if tk.container != nil {
		return tk.container.Terminate(ctx)
	}
	return nil
}
