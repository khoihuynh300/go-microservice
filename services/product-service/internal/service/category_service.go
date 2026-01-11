package service

import (
	"context"

	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/dto"
	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/models"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, input *dto.CreateCategoryDTO) (*models.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*models.Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error)
	ListCategories(ctx context.Context, parentID *string) ([]*models.Category, error)
	ListRootCategories(ctx context.Context) ([]*models.Category, error)
	ListChildCategories(ctx context.Context, parentID string) ([]*models.Category, error)
	UpdateCategory(ctx context.Context, input *dto.UpdateCategoryDTO) (*models.Category, error)
	DeleteCategory(ctx context.Context, categoryID string) error
}
