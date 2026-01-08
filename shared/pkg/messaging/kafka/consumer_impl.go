package kafka

import (
	"context"
	"encoding/json"

	"github.com/khoihuynh300/go-microservice/shared/pkg/const/contextkeys"
	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/events"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type MessageHandler func(ctx context.Context, event *events.Event) error

type KafkaConsumer struct {
	reader   *kafka.Reader
	handlers map[string]MessageHandler
}

func NewConsumer(brokers, topics []string, groupID string) Consumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			GroupTopics: topics,
			GroupID:     groupID,
		}),
		handlers: make(map[string]MessageHandler),
	}
}

func (c *KafkaConsumer) RegisterHandler(topic string, handler MessageHandler) {
	c.handlers[topic] = handler
}

func (c *KafkaConsumer) Start(ctx context.Context, logger *zap.Logger) error {
	for {
		select {
		case <-ctx.Done():
			return c.reader.Close()

		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return nil
				}
				logger.Error("Error fetching message", zap.Error(err))
				continue
			}

			handler, exists := c.handlers[msg.Topic]
			if !exists {
				logger.Warn("No handler registered for topic", zap.String("topic", msg.Topic))
				c.reader.CommitMessages(ctx, msg)
				continue
			}

			var event = &events.Event{}
			err = json.Unmarshal(msg.Value, event)
			if err != nil {
				logger.Error("Failed to parse event", zap.Error(err))
				if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
					logger.Error("Failed to commit invalid message", zap.Error(commitErr))
				}
				continue
			}

			logger := logger.With(zap.String("trace_id", event.TraceID))
			ctx = context.WithValue(ctx, contextkeys.LoggerKey, logger)

			if err := handler(ctx, event); err != nil {
				logger.Error("Error handling message", zap.String("topic", msg.Topic), zap.Error(err))
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				logger.Error("Error committing message", zap.Error(err))
			}
		}
	}

}

func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
