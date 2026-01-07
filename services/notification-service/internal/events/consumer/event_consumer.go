package consumer

import (
	"context"
	"log"

	"github.com/khoihuynh300/go-microservice/notification-service/internal/events/handlers"
	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/events"
	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/kafka"
)

type EventConsumer struct {
	consumer kafka.Consumer
	handlers []handlers.EventHandler
}

func NewEventConsumer(consumer kafka.Consumer) *EventConsumer {
	return &EventConsumer{
		consumer: consumer,
		handlers: make([]handlers.EventHandler, 0),
	}
}

func (c *EventConsumer) RegisterHandler(handler handlers.EventHandler) {
	c.handlers = append(c.handlers, handler)
	log.Println("Event handler registered")
}

func (c *EventConsumer) Start(ctx context.Context) error {
	log.Println("Starting event consumer...")

	messageHandler := func(ctx context.Context, event *events.Event) error {
		log.Printf("Received event: %s (type: %s)", event.EventID, event.EventType)

		var selectedHandler handlers.EventHandler
		for _, handler := range c.handlers {
			if handler.CanHandle(event.EventType) {
				selectedHandler = handler
				break
			}
		}

		if selectedHandler == nil {
			log.Printf("No handler registered for event type: %s", event.EventType)
			return nil
		}

		if err := selectedHandler.HandleEvent(ctx, event); err != nil {
			log.Printf("Error handling event %s: %v", event.EventID, err)
			return err
		}

		log.Printf("Successfully handled event: %s", event.EventID)
		return nil
	}

	return c.consumer.ConsumeWithHandler(ctx, messageHandler)
}

func (c *EventConsumer) Stop() error {
	log.Println("Stopping event consumer...")
	return c.consumer.Close()
}
