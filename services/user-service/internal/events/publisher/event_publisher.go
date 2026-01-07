package publisher

import (
	"context"

	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
)

type EventPublisher interface {
	PublishVerifyEmail(ctx context.Context, user *models.User, token string) error
	PublishEmailVerifySuccess(ctx context.Context, email string) error
	PublishForgotPassword(ctx context.Context, user *models.User, token string) error
	PublishPasswordResetSuccess(ctx context.Context, email string) error

	Close() error
}
