package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/events"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
}

func NewConsumer(broker []string) Consumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: broker,
			Topic:   "user-events",
			GroupID: "notification-service-group",
		}),
	}
}

func (c *KafkaConsumer) ConsumeWithHandler(ctx context.Context, handler func(context.Context, *events.Event) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					return err
				}
				log.Printf("error fetching message: %v", err)
				continue
			}

			var event = &events.Event{}
			err = json.Unmarshal(msg.Value, event)
			if err != nil {
				log.Printf("failed to parse event: %v", err)
				if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
					log.Printf("failed to commit invalid message: %v", commitErr)
				}
				continue
			}

			if err := handler(ctx, event); err != nil {
				log.Printf("handler error for event %s: %v", event.EventID, err)
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("failed to commit message: %v", err)
			}
		}
	}
}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
