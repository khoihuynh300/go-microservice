package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/dto"
	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/product-service/internal/repository"
	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
	zaplogger "github.com/khoihuynh300/go-microservice/shared/pkg/logger"
	"github.com/khoihuynh300/go-microservice/shared/pkg/storage"
	"go.uber.org/zap"
)

type productService struct {
	productRepo      repository.ProductRepository
	productImageRepo repository.ProductImageRepository
	categoryRepo     repository.CategoryRepository
	imageStorage     storage.Storage
}

func NewProductService(
	productRepo repository.ProductRepository,
	productImageRepo repository.ProductImageRepository,
	categoryRepo repository.CategoryRepository,
	imageStorage storage.Storage,
) ProductService {
	return &productService{
		productRepo:      productRepo,
		productImageRepo: productImageRepo,
		categoryRepo:     categoryRepo,
		imageStorage:     imageStorage,
	}
}

func (s *productService) CreateProduct(ctx context.Context, dto *dto.CreateProductDTO) (*models.Product, error) {
	logger := zaplogger.FromContext(ctx)

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

	if err = s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	logger.Info("Product created",
		zap.String("product_id", product.ID.String()),
		zap.String("name", product.Name),
		zap.String("product_sku", product.SKU),
		zap.String("product_slug", product.Slug),
	)

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

	images, err := s.productImageRepo.GetByProductID(ctx, productUUID)
	if err != nil {
		return nil, err
	}

	product.Images = make([]string, len(images))
	for i, img := range images {
		product.Images[i] = img.ImageURL
	}

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

	images, err := s.productImageRepo.GetByProductID(ctx, product.ID)
	if err != nil {
		return nil, err
	}

	product.Images = make([]string, len(images))
	for i, img := range images {
		product.Images[i] = img.ImageURL
	}

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

	images, err := s.productImageRepo.GetByProductID(ctx, product.ID)
	if err != nil {
		return nil, err
	}

	product.Images = make([]string, len(images))
	for i, img := range images {
		product.Images[i] = img.ImageURL
	}

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
	logger := zaplogger.FromContext(ctx)

	productUUID, err := uuid.Parse(dto.ID)
	if err != nil {
		return nil, err
	}

	var product *models.Product

	err = s.productRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		product, err = s.productRepo.GetByIDForUpdate(ctx, productUUID)
		if err != nil {
			return err
		}

		if err = s.updateProductInfo(dto, product); err != nil {
			return err
		}

		if err = s.productRepo.Update(ctx, product); err != nil {
			return err
		}

		if dto.Images != nil {
			if err = s.handleImageUpdates(ctx, productUUID, *dto.Images); err != nil {
				return err
			}
		}

		logger.Info("Product updated",
			zap.String("product_id", product.ID.String()),
			zap.String("name", product.Name),
		)

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Fetch the updated images
	updatedImages, err := s.productImageRepo.GetByProductID(ctx, product.ID)
	if err == nil {
		product.Images = make([]string, len(updatedImages))
		for i, img := range updatedImages {
			product.Images[i] = img.ImageURL
		}
	}

	return product, nil
}

func (s *productService) updateProductInfo(dto *dto.UpdateProductDTO, product *models.Product) error {
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
			return err
		}
		product.CategoryID = categoryID
	}

	if dto.Price != nil {
		product.Price = *dto.Price
	}

	if dto.Thumbnail != nil {
		product.Thumbnail = dto.Thumbnail
	}

	return nil
}

func (s *productService) handleImageUpdates(
	ctx context.Context,
	productID uuid.UUID,
	newImageURLs []string,
) error {
	logger := zaplogger.FromContext(ctx)

	currentImages, err := s.productImageRepo.GetByProductIDForUpdate(ctx, productID)
	if err != nil {
		return err
	}

	// Map for lookup
	currentMap := make(map[string]*models.ProductImage)
	for _, img := range currentImages {
		currentMap[img.ImageURL] = img
	}

	keptImageIDs := make(map[uuid.UUID]bool)
	var imagesToDeleteFromS3 []string

	for i, url := range newImageURLs {
		// update position if exists else add new product image
		if img, exists := currentMap[url]; exists {
			keptImageIDs[img.ID] = true
			if img.Position != int32(i) {
				if err := s.productImageRepo.UpdatePosition(ctx, img.ID, int32(i)); err != nil {
					return err
				}
			}
		} else {
			if err := s.productImageRepo.Create(ctx, productID, url, int32(i)); err != nil {
				return err
			}
		}
	}

	// Remove images from DB and prepare list for S3 deletion
	for _, img := range currentImages {
		if !keptImageIDs[img.ID] {
			if err := s.productImageRepo.Delete(ctx, productID, img.ID); err != nil {
				return err
			}
			imagesToDeleteFromS3 = append(imagesToDeleteFromS3, img.ImageURL)
		}
	}

	// Delete images from S3
	if len(imagesToDeleteFromS3) > 0 {
		go func(urls []string) {
			bgCtx := context.Background()
			for _, url := range urls {
				if err := s.imageStorage.Delete(bgCtx, url); err != nil {
					logger.Error("Failed to delete image from MinIO storage",
						zap.String("image_url", url),
						zap.Error(err),
					)
				}
			}
		}(imagesToDeleteFromS3)
	}

	return nil
}

func (s *productService) DeleteProduct(ctx context.Context, productID string) error {
	logger := zaplogger.FromContext(ctx)

	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return err
	}

	return s.productRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		product, err := s.productRepo.GetByIDForUpdate(ctx, productUUID)
		if err != nil {
			return err
		}
		if product == nil {
			return apperr.ErrProductNotFound
		}

		if err := s.productRepo.SoftDelete(ctx, productUUID); err != nil {
			return err
		}

		logger.Info("Product deleted",
			zap.String("product_id", productUUID.String()),
		)

		return nil
	})
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
