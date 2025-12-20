package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
)

type RegistryTokenRepository interface {
	Repository
	Create(ctx context.Context, token_hash string, userID uuid.UUID, expiresAt time.Time) error
	GetByToken(ctx context.Context, token string) (*models.RegistryToken, error)
	InvalidateToken(ctx context.Context, token_hash string) error
	MarkTokenAsUsed(ctx context.Context, token_hash string) error
}
