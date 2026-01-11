package impl

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/khoihuynh300/go-microservice/product-service/internal/db/generated"
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
		ProductID: productID,
		ImageUrl:  imageURL,
		Position:  position,
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *productImageRepository) GetByProductID(ctx context.Context, productID uuid.UUID) ([]string, error) {
	dbImages, err := r.queries(ctx).GetProductImages(ctx, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	imageURLs := make([]string, len(dbImages))
	for i, img := range dbImages {
		imageURLs[i] = img.ImageUrl
	}

	return imageURLs, nil
}

func (r *productImageRepository) Delete(ctx context.Context, productID, imageID uuid.UUID) error {
	_, err := r.queries(ctx).DeleteProductImage(ctx, imageID)
	if err != nil {
		return err
	}

	return nil
}

func (r *productImageRepository) DeleteAllByProductID(ctx context.Context, productID uuid.UUID) error {
	_, err := r.queries(ctx).DeleteAllProductImages(ctx, productID)
	if err != nil {
		return err
	}

	return nil
}
