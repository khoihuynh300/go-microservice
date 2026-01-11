package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/dto"
	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/product-service/internal/repository"
	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
)

type productService struct {
	productRepo      repository.ProductRepository
	productImageRepo repository.ProductImageRepository
	categoryRepo     repository.CategoryRepository
}

func NewProductService(
	productRepo repository.ProductRepository,
	productImageRepo repository.ProductImageRepository,
	categoryRepo repository.CategoryRepository,
) ProductService {
	return &productService{
		productRepo:      productRepo,
		productImageRepo: productImageRepo,
		categoryRepo:     categoryRepo,
	}
}

func (s *productService) CreateProduct(ctx context.Context, dto *dto.CreateProductDTO) (*models.Product, error) {
	categoryID, err := uuid.Parse(dto.CategoryID)
	if err != nil {
		return nil, err
	}

	_, err = s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, apperr.ErrCategoryNotFound
	}

	existingProduct, err := s.productRepo.GetBySKU(ctx, dto.SKU)
	if err == nil && existingProduct != nil {
		return nil, apperr.ErrProductSKUExists
	}

	existingProduct, err = s.productRepo.GetBySlug(ctx, dto.Slug)
	if err == nil && existingProduct != nil {
		return nil, apperr.ErrProductSlugExists
	}

	product := &models.Product{
		SKU:         dto.SKU,
		Name:        dto.Name,
		Slug:        dto.Slug,
		Description: dto.Description,
		CategoryID:  categoryID,
		Price:       dto.Price,
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	// TODO: add image, add thumnail

	return product, nil
}

func (s *productService) GetProductByID(ctx context.Context, productID string) (*models.Product, error) {
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, err
	}

	product, err := s.productRepo.GetByID(ctx, productUUID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, apperr.ErrProductNotFound
	}

	// TODO: load images

	return product, nil
}

func (s *productService) GetProductBySKU(ctx context.Context, sku string) (*models.Product, error) {
	product, err := s.productRepo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, apperr.ErrProductNotFound
	}

	// TODO: load images

	return product, nil
}

func (s *productService) GetProductBySlug(ctx context.Context, slug string) (*models.Product, error) {
	product, err := s.productRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, apperr.ErrProductNotFound
	}

	// TODO: load images

	return product, nil
}

func (s *productService) ListProducts(ctx context.Context, dto *dto.ListProductsDTO) ([]*models.Product, int64, error) {
	var categoryID *uuid.UUID
	if dto.CategoryID != nil {
		id, err := uuid.Parse(*dto.CategoryID)
		if err != nil {
			return nil, 0, err
		}
		categoryID = &id
	}

	return s.productRepo.List(ctx, categoryID, dto.Page, dto.PageSize)
}

func (s *productService) SearchProducts(ctx context.Context, dto *dto.SearchProductsDTO) ([]*models.Product, int64, error) {
	return s.productRepo.Search(ctx, dto.SearchQuery, dto.Page, dto.PageSize)
}

func (s *productService) UpdateProduct(ctx context.Context, dto *dto.UpdateProductDTO) (*models.Product, error) {
	productUUID, err := uuid.Parse(dto.ID)
	if err != nil {
		return nil, err
	}

	product, err := s.productRepo.GetByID(ctx, productUUID)
	if err != nil {
		return nil, err
	}

	if dto.Name != nil {
		product.Name = *dto.Name
	}
	if dto.SKU != nil {
		product.SKU = *dto.SKU
	}
	if dto.Slug != nil {
		product.Slug = *dto.Slug
	}
	if dto.Description != nil {
		product.Description = *dto.Description
	}
	if dto.CategoryID != nil {
		categoryID, err := uuid.Parse(*dto.CategoryID)
		if err != nil {
			return nil, err
		}
		product.CategoryID = categoryID
	}
	if dto.Price != nil {
		product.Price = *dto.Price
	}

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *productService) DeleteProduct(ctx context.Context, productID string) error {
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return err
	}

	return s.productRepo.Delete(ctx, productUUID)
}

func (s *productService) GetProductsByIDs(ctx context.Context, dto *dto.GetProductsByIDsDTO) ([]*models.Product, error) {
	ids := make([]uuid.UUID, len(dto.IDs))
	for i, idStr := range dto.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		ids[i] = id
	}

	return s.productRepo.GetByIDs(ctx, ids)
}
