package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
)

type UserRepository interface {
	Repository
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) (int64, error)
	UpdateAvatar(ctx context.Context, id uuid.UUID, avatarURL string) (int64, error)
	VerifyEmail(ctx context.Context, id uuid.UUID) (int64, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) (int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.UserStatus) (int64, error)
	SoftDelete(ctx context.Context, id uuid.UUID) (int64, error)
}
