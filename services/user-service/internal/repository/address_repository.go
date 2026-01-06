package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
)

type AddressRepository interface {
	Repository
	Create(ctx context.Context, address *models.Address) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Address, error)
	GetByIDAndUserID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Address, error)
	Update(ctx context.Context, address *models.Address) (int64, error)
	Delete(ctx context.Context, id uuid.UUID) (int64, error)
	SetDefaultAddress(ctx context.Context, userID, addressID uuid.UUID) (int64, error)
}
