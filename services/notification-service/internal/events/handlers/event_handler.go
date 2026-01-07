package handlers

import (
	"context"

	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/events"
)

type EventHandler interface {
	HandleEvent(ctx context.Context, event *events.Event) error
	CanHandle(eventType string) bool
}
