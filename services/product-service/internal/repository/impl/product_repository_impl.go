package impl

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/khoihuynh300/go-microservice/product-service/internal/db/generated"
	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/product-service/internal/repository"
	"github.com/khoihuynh300/go-microservice/product-service/internal/utils/convert"
)

type productRepository struct {
	baseRepository
}

func NewProductRepository(db *pgxpool.Pool) repository.ProductRepository {
	return &productRepository{
		baseRepository: baseRepository{
			db: db,
			q:  sqlc.New(db),
		},
	}
}

func (r *productRepository) Create(ctx context.Context, product *models.Product) error {
	numericPrice, err := convert.DoubleToNumeric(product.Price)
	if err != nil {
		return err
	}

	dbProduct, err := r.queries(ctx).CreateProduct(ctx, sqlc.CreateProductParams{
		ID:          uuid.New(),
		Name:        product.Name,
		Sku:         product.SKU,
		Slug:        product.Slug,
		Description: pgtype.Text{String: product.Description, Valid: product.Description != ""},
		CategoryID:  pgtype.UUID{Bytes: product.CategoryID, Valid: product.CategoryID != uuid.Nil},
		Price:       numericPrice,
		Thumbnail:   pgtype.Text{String: product.Thumbnail, Valid: product.Thumbnail != ""},
	})

	if err != nil {
		return err
	}

	product.ID = dbProduct.ID
	product.CreatedAt = dbProduct.CreatedAt.Time
	product.UpdatedAt = dbProduct.UpdatedAt.Time
	return nil
}

func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	dbProduct, err := r.queries(ctx).GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return r.toModel(&dbProduct), nil
}

func (r *productRepository) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
	dbProduct, err := r.queries(ctx).GetProductBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return r.toModel(&dbProduct), nil
}

func (r *productRepository) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	dbProduct, err := r.queries(ctx).GetProductBySKU(ctx, sku)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return r.toModel(&dbProduct), nil
}

func (r *productRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Product, error) {
	dbProducts, err := r.queries(ctx).ListProductsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	products := make([]*models.Product, len(dbProducts))
	for i, dbProduct := range dbProducts {
		products[i] = r.toModel(&dbProduct)
	}

	return products, nil
}

func (r *productRepository) List(ctx context.Context, categoryID *uuid.UUID, page, pageSize int32) ([]*models.Product, int64, error) {
	total, err := r.queries(ctx).CountProducts(ctx, convert.PtrToUUID(categoryID))
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	listParams := sqlc.ListProductsParams{
		Limit:      pageSize,
		Offset:     offset,
		CategoryID: convert.PtrToUUID(categoryID),
	}

	dbProducts, err := r.queries(ctx).ListProducts(ctx, listParams)
	if err != nil {
		return nil, 0, err
	}

	products := make([]*models.Product, len(dbProducts))
	for i, dbProduct := range dbProducts {
		products[i] = r.toModel(&dbProduct)
	}

	return products, total, nil
}

func (r *productRepository) Search(ctx context.Context, query string, page, pageSize int32) ([]*models.Product, int64, error) {
	total, err := r.queries(ctx).CountSearchProducts(ctx, pgtype.Text{String: query, Valid: true})
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	dbProducts, err := r.queries(ctx).SearchProducts(ctx, sqlc.SearchProductsParams{
		Column1: pgtype.Text{String: query, Valid: true},
		Limit:   pageSize,
		Offset:  offset,
	})
	if err != nil {
		return nil, 0, err
	}

	products := make([]*models.Product, len(dbProducts))
	for i, dbProduct := range dbProducts {
		products[i] = r.toModel(&dbProduct)
	}

	return products, total, nil
}

func (r *productRepository) Update(ctx context.Context, product *models.Product) error {
	numericPrice, err := convert.DoubleToNumeric(product.Price)
	if err != nil {
		return err
	}

	_, err = r.queries(ctx).UpdateProduct(ctx, sqlc.UpdateProductParams{
		ID:          product.ID,
		Name:        pgtype.Text{String: product.Name, Valid: true},
		Sku:         pgtype.Text{String: product.SKU, Valid: true},
		Slug:        pgtype.Text{String: product.Slug, Valid: true},
		Description: pgtype.Text{String: product.Description, Valid: product.Description != ""},
		CategoryID:  pgtype.UUID{Bytes: product.CategoryID, Valid: product.CategoryID != uuid.Nil},
		Price:       numericPrice,
		Thumbnail:   pgtype.Text{String: product.Thumbnail, Valid: product.Thumbnail != ""},
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.queries(ctx).SoftDeleteProduct(ctx, sqlc.SoftDeleteProductParams{
		ID:        id,
		DeletedAt: pgtype.Timestamptz{Time: now, Valid: true},
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *productRepository) toModel(dbProduct *sqlc.Product) *models.Product {
	return &models.Product{
		ID:          dbProduct.ID,
		SKU:         dbProduct.Sku,
		Name:        dbProduct.Name,
		Slug:        dbProduct.Slug,
		Description: dbProduct.Description.String,
		CategoryID:  dbProduct.CategoryID.Bytes,
		Price:       convert.NumericToDouble(dbProduct.Price),
		Thumbnail:   dbProduct.Thumbnail.String,
		CreatedAt:   dbProduct.CreatedAt.Time,
		UpdatedAt:   dbProduct.UpdatedAt.Time,
	}
}
