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

type categoryService struct {
	categoryRepo repository.CategoryRepository
	imageStorage storage.Storage
}

func NewCategoryService(
	categoryRepo repository.CategoryRepository,
	imageStorage storage.Storage,
) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
		imageStorage: imageStorage,
	}
}

func (s *categoryService) CreateCategory(ctx context.Context, input *dto.CreateCategoryDTO) (*models.Category, error) {
	logger := zaplogger.FromContext(ctx)

	existCategory, err := s.categoryRepo.GetByName(ctx, input.Name)
	if err != nil {
		return nil, err
	}
	if existCategory != nil {
		return nil, apperr.ErrCategoryAlreadyExists
	}

	existCategory, err = s.categoryRepo.GetBySlug(ctx, input.Slug)
	if err != nil {
		return nil, err
	}
	if existCategory != nil {
		return nil, apperr.ErrCategorySlugExists
	}

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

	if err = s.categoryRepo.Create(ctx, category); err != nil {
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

	var category *models.Category
	var oldImageToDelete string

	err = s.categoryRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		category, err = s.categoryRepo.GetByIDForUpdate(ctx, categoryUUID)
		if err != nil {
			return err
		}
		if category == nil {
			return apperr.ErrCategoryNotFound
		}

		if dto.ImageURL != nil && category.ImageURL != nil && *category.ImageURL != "" && *dto.ImageURL != *category.ImageURL {
			oldImageToDelete = *category.ImageURL
		}

		if err = s.updateCategoryInfo(ctx, dto, category); err != nil {
			return err
		}

		if err = s.categoryRepo.Update(ctx, category); err != nil {
			return err
		}

		logger.Info("Category updated",
			zap.String("category_id", category.ID.String()),
			zap.String("name", category.Name),
		)

		return nil
	})

	if err != nil {
		return nil, err
	}

	if oldImageToDelete != "" {
		go func(oldImageURL string) {
			if err := s.imageStorage.Delete(context.Background(), oldImageURL); err != nil {
				logger.Error("Failed to delete old category image", zap.String("url", oldImageURL), zap.Error(err))
			}
		}(oldImageToDelete)
	}

	return category, nil
}

func (s *categoryService) updateCategoryInfo(
	ctx context.Context,
	dto *dto.UpdateCategoryDTO,
	category *models.Category,
) error {
	if dto.Name != nil {
		category.Name = *dto.Name
	}

	if dto.Slug != nil {
		category.Slug = *dto.Slug
	}

	if dto.Description != nil {
		category.Description = *dto.Description
	}

	if dto.ImageURL != nil {
		category.ImageURL = dto.ImageURL

	}

	if dto.ParentID != nil {
		if *dto.ParentID == "" {
			// Unset parent category
			category.ParentID = nil
		} else {
			// Set parent category
			parentID, err := uuid.Parse(*dto.ParentID)
			if err != nil {
				return err
			}

			if parentID == category.ID {
				return apperr.ErrCategoryCannotBeOwnParent
			}

			parentCategory, err := s.categoryRepo.GetByIDForUpdate(ctx, parentID)
			if err != nil {
				return err
			}
			if parentCategory == nil {
				return apperr.ErrParentCategoryNotFound
			}

			category.ParentID = &parentID
		}
	}

	return nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, categoryID string) error {
	logger := zaplogger.FromContext(ctx)

	categoryUUID, err := uuid.Parse(categoryID)
	if err != nil {
		return err
	}

	return s.categoryRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		category, err := s.categoryRepo.GetByIDForUpdate(ctx, categoryUUID)
		if err != nil {
			return err
		}
		if category == nil {
			return apperr.ErrCategoryNotFound
		}

		children, err := s.categoryRepo.ListChildren(ctx, categoryUUID)
		if err != nil {
			return err
		}
		if len(children) > 0 {
			return apperr.ErrCategoryHasChildren
		}

		err = s.categoryRepo.SoftDelete(ctx, categoryUUID)
		if err != nil {
			return err
		}

		logger.Info("Category deleted",
			zap.String("category_id", categoryUUID.String()),
		)

		return nil
	})
}
