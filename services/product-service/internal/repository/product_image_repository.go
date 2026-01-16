package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/models"
)

type ProductImageRepository interface {
	Repository

	Create(ctx context.Context, productID uuid.UUID, imageURL string, position int32) error
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]*models.ProductImage, error)
	GetByProductIDForUpdate(ctx context.Context, productID uuid.UUID) ([]*models.ProductImage, error)
	Delete(ctx context.Context, productID, imageID uuid.UUID) error
	DeleteAllByProductID(ctx context.Context, productID uuid.UUID) error
	UpdatePosition(ctx context.Context, imageID uuid.UUID, position int32) error
}
