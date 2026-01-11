package repository

import (
	"context"

	"github.com/google/uuid"
)

type ProductImageRepository interface {
	Repository

	Create(ctx context.Context, productID uuid.UUID, imageURL string, position int32) error
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]string, error)
	Delete(ctx context.Context, productID, imageID uuid.UUID) error
	DeleteAllByProductID(ctx context.Context, productID uuid.UUID) error
}
