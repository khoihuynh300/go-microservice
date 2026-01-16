package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/models"
)

type ProductRepository interface {
	Repository

	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetBySlug(ctx context.Context, slug string) (*models.Product, error)
	GetBySKU(ctx context.Context, sku string) (*models.Product, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Product, error)
	List(ctx context.Context, categoryID *uuid.UUID, page, pageSize int32) ([]*models.Product, int64, error)
	Search(ctx context.Context, query string, page, pageSize int32) ([]*models.Product, int64, error)
	Update(ctx context.Context, product *models.Product) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
