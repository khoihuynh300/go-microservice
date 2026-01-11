package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/models"
)

type CategoryRepository interface {
	Repository

	Create(ctx context.Context, category *models.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error)
	GetBySlug(ctx context.Context, slug string) (*models.Category, error)
	List(ctx context.Context, parentID *uuid.UUID) ([]*models.Category, error)
	ListRoots(ctx context.Context) ([]*models.Category, error)
	ListChildren(ctx context.Context, parentID uuid.UUID) ([]*models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
}
