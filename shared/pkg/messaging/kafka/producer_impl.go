package kafka

import (
	"context"
	"encoding/json"

	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/events"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) Producer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr: kafka.TCP(brokers...),
		},
	}
}

func (p *KafkaProducer) Publish(ctx context.Context, topic string, event *events.Event) error {
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: value,
	})
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
