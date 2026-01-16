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

type categoryRepository struct {
	baseRepository
}

func NewCategoryRepository(db *pgxpool.Pool) repository.CategoryRepository {
	return &categoryRepository{
		baseRepository: baseRepository{
			db: db,
			q:  sqlc.New(db),
		},
	}
}

func (r *categoryRepository) Create(ctx context.Context, category *models.Category) error {
	now := time.Now()

	dbCategory, err := r.queries(ctx).CreateCategory(ctx, sqlc.CreateCategoryParams{
		ID:          uuid.New(),
		ParentID:    convert.PtrToUUID(category.ParentID),
		Name:        category.Name,
		Slug:        category.Slug,
		Description: pgtype.Text{String: category.Description, Valid: true},
		ImageUrl:    convert.PtrToText(category.ImageURL),
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	if err != nil {
		return err
	}

	category.ID = dbCategory.ID
	category.CreatedAt = dbCategory.CreatedAt
	category.UpdatedAt = dbCategory.UpdatedAt
	return nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	dbCategory, err := r.queries(ctx).GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return r.toModel(&dbCategory), nil
}

func (r *categoryRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	dbCategory, err := r.queries(ctx).GetCategoryByIDForUpdate(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return r.toModel(&dbCategory), nil
}

func (r *categoryRepository) GetByName(ctx context.Context, name string) (*models.Category, error) {
	dbCategory, err := r.queries(ctx).GetCategoryByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return r.toModel(&dbCategory), nil
}

func (r *categoryRepository) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	dbCategory, err := r.queries(ctx).GetCategoryBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return r.toModel(&dbCategory), nil
}

func (r *categoryRepository) List(ctx context.Context, parentID *uuid.UUID) ([]*models.Category, error) {
	dbCategories, err := r.queries(ctx).ListCategories(ctx, convert.PtrToUUID(parentID))
	if err != nil {
		return nil, err
	}

	categories := make([]*models.Category, len(dbCategories))
	for i, dbCategory := range dbCategories {
		categories[i] = r.toModel(&dbCategory)
	}

	return categories, nil
}

func (r *categoryRepository) ListRoots(ctx context.Context) ([]*models.Category, error) {
	dbCategories, err := r.queries(ctx).ListRootCategories(ctx)
	if err != nil {
		return nil, err
	}

	categories := make([]*models.Category, len(dbCategories))
	for i, dbCategory := range dbCategories {
		categories[i] = r.toModel(&dbCategory)
	}

	return categories, nil
}

func (r *categoryRepository) ListChildren(ctx context.Context, parentID uuid.UUID) ([]*models.Category, error) {
	dbCategories, err := r.queries(ctx).ListChildCategories(ctx, pgtype.UUID{Bytes: parentID, Valid: true})
	if err != nil {
		return nil, err
	}

	categories := make([]*models.Category, len(dbCategories))
	for i, dbCategory := range dbCategories {
		categories[i] = r.toModel(&dbCategory)
	}

	return categories, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *models.Category) error {
	now := time.Now()

	return r.queries(ctx).UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		Name:        category.Name,
		Slug:        category.Slug,
		Description: pgtype.Text{String: category.Description, Valid: true},
		ImageUrl:    convert.PtrToText(category.ImageURL),
		ParentID:    convert.PtrToUUID(category.ParentID),
		UpdatedAt:   now,
		ID:          category.ID,
	})
}

func (r *categoryRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()

	return r.queries(ctx).SoftDeleteCategory(ctx, sqlc.SoftDeleteCategoryParams{
		ID:        id,
		DeletedAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
}

func (r *categoryRepository) toModel(dbCategory *sqlc.Category) *models.Category {
	return &models.Category{
		ID:          dbCategory.ID,
		ParentID:    convert.PgUUIDToPtr(dbCategory.ParentID),
		Name:        dbCategory.Name,
		Slug:        dbCategory.Slug,
		Description: dbCategory.Description.String,
		ImageURL:    convert.PgTextToPtr(dbCategory.ImageUrl),
		CreatedAt:   dbCategory.CreatedAt,
		UpdatedAt:   dbCategory.UpdatedAt,
	}
}
