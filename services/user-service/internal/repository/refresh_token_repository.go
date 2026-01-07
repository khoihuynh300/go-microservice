package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
)

type RefreshTokenRepository interface {
	Repository
	Create(ctx context.Context, refreshToken *models.RefreshToken) error
	GetByToken(ctx context.Context, refreshTokenStr string) (*models.RefreshToken, error)
	DeleteByID(ctx context.Context, id uuid.UUID) (int64, error)
}
