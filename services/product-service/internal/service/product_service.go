package service

import (
	"context"

	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/dto"
	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/models"
)

type ProductService interface {
	CreateProduct(ctx context.Context, input *dto.CreateProductDTO) (*models.Product, error)
	GetProductByID(ctx context.Context, productID string) (*models.Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*models.Product, error)
	GetProductBySKU(ctx context.Context, sku string) (*models.Product, error)
	ListProducts(ctx context.Context, input *dto.ListProductsDTO) ([]*models.Product, int64, error)
	SearchProducts(ctx context.Context, input *dto.SearchProductsDTO) ([]*models.Product, int64, error)
	UpdateProduct(ctx context.Context, input *dto.UpdateProductDTO) (*models.Product, error)
	DeleteProduct(ctx context.Context, productID string) error
	GetProductsByIDs(ctx context.Context, input *dto.GetProductsByIDsDTO) ([]*models.Product, error)
}
