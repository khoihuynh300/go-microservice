package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/dto"
	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/product-service/internal/repository"
	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
	zaplogger "github.com/khoihuynh300/go-microservice/shared/pkg/logger"
	"go.uber.org/zap"
)

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(
	categoryRepo repository.CategoryRepository,
) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *categoryService) CreateCategory(ctx context.Context, input *dto.CreateCategoryDTO) (*models.Category, error) {
	logger := zaplogger.FromContext(ctx)

	category := &models.Category{
		ID:          uuid.New(),
		Name:        input.Name,
		Slug:        input.Slug,
		Description: input.Description,
	}

	if input.ParentID != nil {
		parentID, err := uuid.Parse(*input.ParentID)
		if err != nil {
			return nil, err
		}

		_, err = s.categoryRepo.GetByID(ctx, parentID)
		if err != nil {
			return nil, apperr.ErrCategoryNotFound
		}

		category.ParentID = &parentID
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	logger.Info("Category created",
		zap.String("category_id", category.ID.String()),
		zap.String("name", category.Name),
	)

	return category, nil
}

func (s *categoryService) GetCategoryByID(ctx context.Context, categoryID string) (*models.Category, error) {
	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return nil, err
	}

	category, err := s.categoryRepo.GetByID(ctx, categoryUUID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, apperr.ErrCategoryNotFound
	}

	return category, nil
}

func (s *categoryService) GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error) {
	category, err := s.categoryRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, apperr.ErrCategoryNotFound
	}

	return category, nil
}

func (s *categoryService) ListCategories(ctx context.Context, parentID *string) ([]*models.Category, error) {
	var parentUUID *uuid.UUID
	if parentID != nil {
		id, err := uuid.Parse(*parentID)
		if err != nil {
			return nil, err
		}
		parentUUID = &id
	}

	return s.categoryRepo.List(ctx, parentUUID)
}

func (s *categoryService) ListRootCategories(ctx context.Context) ([]*models.Category, error) {
	return s.categoryRepo.ListRoots(ctx)
}

func (s *categoryService) ListChildCategories(ctx context.Context, parentID string) ([]*models.Category, error) {
	parentUUID, err := uuid.Parse(parentID)
	if err != nil {
		return nil, err
	}

	return s.categoryRepo.ListChildren(ctx, parentUUID)
}

func (s *categoryService) UpdateCategory(ctx context.Context, dto *dto.UpdateCategoryDTO) (*models.Category, error) {
	logger := zaplogger.FromContext(ctx)

	categoryUUID, err := uuid.Parse(dto.ID)
	if err != nil {
		return nil, err
	}

	category, err := s.categoryRepo.GetByID(ctx, categoryUUID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, apperr.ErrCategoryNotFound
	}

	if dto.Name != nil {
		category.Name = *dto.Name
	}
	if dto.Slug != nil {
		category.Slug = *dto.Slug
	}
	if dto.Description != nil {
		category.Description = *dto.Description
	}
	if dto.ParentID != nil {
		parentID, err := uuid.Parse(*dto.ParentID)
		if err != nil {
			return nil, err
		}

		if parentID == categoryUUID {
			return nil, fmt.Errorf("category cannot be its own parent")
		}

		category, err = s.categoryRepo.GetByID(ctx, parentID)
		if err != nil {
			return nil, err
		}
		if category == nil {
			return nil, apperr.ErrCategoryNotFound
		}

		category.ParentID = &parentID
	}

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	logger.Info("Category updated",
		zap.String("category_id", category.ID.String()),
		zap.String("name", category.Name),
		zap.Any("data", dto),
	)

	return category, nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, categoryID string) error {
	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return err
	}

	children, err := s.categoryRepo.ListChildren(ctx, categoryUUID)
	if err != nil {
		return err
	}

	if len(children) > 0 {
		return fmt.Errorf("cannot delete category with children")
	}

	return s.categoryRepo.Delete(ctx, categoryUUID)
}
