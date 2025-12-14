package repository

import (
	"context"

	"github.com/khoihuynh300/go-microservice/user-service/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context, status domain.UserStatus, limit, offset int) ([]*domain.User, error)
	Count(ctx context.Context, status domain.UserStatus) (int64, error)
	Update(ctx context.Context, user *domain.User) error
	UpdatePassword(ctx context.Context, id, hashedPassword string) error
	UpdateStatus(ctx context.Context, id string, status domain.UserStatus) error
	SoftDelete(ctx context.Context, id string) error
}
