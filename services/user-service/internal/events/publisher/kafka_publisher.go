package publisher

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/shared/pkg/const/contextkeys"
	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/events"
	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/kafka"
	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/topics"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
)

type kafkaEventPublisher struct {
	producer kafka.Producer
}

func NewKafkaEventPublisher(producer kafka.Producer) EventPublisher {
	return &kafkaEventPublisher{
		producer: producer,
	}
}

func (p *kafkaEventPublisher) PublishVerifyEmail(ctx context.Context, user *models.User, token string) error {
	traceID := ctx.Value(contextkeys.TraceIDKey).(string)

	payload := &events.Event{
		EventID:    uuid.NewString(),
		EventType:  events.TypeUserRegisteredEvent,
		OccurredAt: time.Now().UTC(),
		TraceID:    traceID,
		Data: &events.UserRegisteredEvent{
			Email:    user.Email,
			FullName: user.FullName,
			Token:    token,
		},
	}
	if err := p.producer.Publish(ctx, topics.UserEventsTopic, payload); err != nil {
		return fmt.Errorf("failed to publish user registered event: %w", err)
	}

	return nil
}

func (p *kafkaEventPublisher) PublishEmailVerifySuccess(ctx context.Context, email string) error {
	traceID := ctx.Value(contextkeys.TraceIDKey).(string)

	payload := &events.Event{
		EventID:    uuid.NewString(),
		EventType:  events.TypeEmailVerifySuccessEvent,
		OccurredAt: time.Now().UTC(),
		TraceID:    traceID,
		Data: &events.EmailVerifySuccessEvent{
			Email: email,
		},
	}
	if err := p.producer.Publish(ctx, topics.UserEventsTopic, payload); err != nil {
		return fmt.Errorf("failed to publish email verified event: %w", err)
	}

	return nil
}

func (p *kafkaEventPublisher) PublishForgotPassword(ctx context.Context, user *models.User, token string) error {
	traceID := ctx.Value(contextkeys.TraceIDKey).(string)

	payload := &events.Event{
		EventID:    uuid.NewString(),
		EventType:  events.TypeForgotPasswordEvent,
		OccurredAt: time.Now().UTC(),
		TraceID:    traceID,
		Data: &events.UserForgotPasswordEvent{
			Email:    user.Email,
			FullName: user.FullName,
			Token:    token,
		},
	}
	if err := p.producer.Publish(ctx, topics.UserEventsTopic, payload); err != nil {
		return fmt.Errorf("failed to publish user registered event: %w", err)
	}

	return nil
}

func (p *kafkaEventPublisher) PublishPasswordResetSuccess(ctx context.Context, email string) error {
	traceID := ctx.Value(contextkeys.TraceIDKey).(string)
	payload := &events.Event{
		EventID:    uuid.NewString(),
		EventType:  events.TypePasswordResetSuccessEvent,
		OccurredAt: time.Now().UTC(),
		TraceID:    traceID,
		Data: &events.UserPasswordResetSuccessEvent{
			Email: email,
		},
	}
	if err := p.producer.Publish(ctx, topics.UserEventsTopic, payload); err != nil {
		return fmt.Errorf("failed to publish password reset success event: %w", err)
	}

	return nil
}

func (p *kafkaEventPublisher) Close() error {
	return p.producer.Close()
}
