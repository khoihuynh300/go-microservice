package kafka

import (
	"context"

	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/events"
)

type Consumer interface {
	ConsumeWithHandler(ctx context.Context, handler func(context.Context, *events.Event) error) error
	Close() error
}
