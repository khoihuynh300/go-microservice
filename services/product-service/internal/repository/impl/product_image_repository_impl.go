package impl

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/khoihuynh300/go-microservice/product-service/internal/db/generated"
	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/product-service/internal/repository"
)

type productImageRepository struct {
	baseRepository
}

func NewProductImageRepository(db *pgxpool.Pool) repository.ProductImageRepository {
	return &productImageRepository{
		baseRepository: baseRepository{
			db: db,
			q:  sqlc.New(db),
		},
	}
}

func (r *productImageRepository) Create(ctx context.Context, productID uuid.UUID, imageURL string, position int32) error {
	_, err := r.queries(ctx).CreateProductImage(ctx, sqlc.CreateProductImageParams{
		ID:        uuid.New(),
		ProductID: productID,
		ImageUrl:  imageURL,
		Position:  position,
	})

	return err
}

func (r *productImageRepository) GetByProductID(ctx context.Context, productID uuid.UUID) ([]*models.ProductImage, error) {
	dbImages, err := r.queries(ctx).GetProductImages(ctx, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return r.ToModels(ctx, dbImages), nil
}

func (r *productImageRepository) GetByProductIDForUpdate(ctx context.Context, productID uuid.UUID) ([]*models.ProductImage, error) {
	dbImages, err := r.queries(ctx).GetProductImagesForUpdate(ctx, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return r.ToModels(ctx, dbImages), nil
}

func (r *productImageRepository) UpdatePosition(ctx context.Context, imageID uuid.UUID, position int32) error {
	return r.queries(ctx).UpdateImagePosition(ctx, sqlc.UpdateImagePositionParams{
		ID:       imageID,
		Position: position,
	})
}

func (r *productImageRepository) Delete(ctx context.Context, productID, imageID uuid.UUID) error {
	return r.queries(ctx).DeleteProductImage(ctx, imageID)
}

func (r *productImageRepository) DeleteAllByProductID(ctx context.Context, productID uuid.UUID) error {
	return r.queries(ctx).DeleteAllProductImages(ctx, productID)
}

func (r *productImageRepository) ToModels(
	ctx context.Context,
	dbImages []sqlc.ProductImage,
) []*models.ProductImage {
	images := make([]*models.ProductImage, len(dbImages))
	for i, img := range dbImages {
		images[i] = &models.ProductImage{
			ID:        img.ID,
			ProductID: img.ProductID,
			ImageURL:  img.ImageUrl,
			Position:  img.Position,
			CreatedAt: img.CreatedAt.Time,
		}
	}

	return images
}
