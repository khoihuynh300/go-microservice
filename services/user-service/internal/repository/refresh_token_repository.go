package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
)

type RefreshTokenRepository interface {
	Save(ctx context.Context, refreshToken *models.RefreshToken) error
	FindByToken(ctx context.Context, refreshTokenStr string) (*models.RefreshToken, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
}
