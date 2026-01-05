package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
)

type UserRepository interface {
	Repository
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.UserStatus) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
