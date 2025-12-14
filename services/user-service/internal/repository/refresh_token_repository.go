package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain"
)

type RefreshTokenRepository interface {
	Save(ctx context.Context, refreshToken *domain.RefreshToken) error
	FindByToken(ctx context.Context, refreshTokenStr string) (*domain.RefreshToken, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
}
