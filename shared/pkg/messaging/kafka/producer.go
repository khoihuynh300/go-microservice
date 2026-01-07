package kafka

import (
	"context"

	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/events"
)

type Producer interface {
	Publish(ctx context.Context, topic string, event *events.Event) error
	Close() error
}
