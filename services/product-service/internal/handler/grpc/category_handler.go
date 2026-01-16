package grpchandler

import (
	"context"

	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/dto"
	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/product-service/internal/utils/convert"
	productpb "github.com/khoihuynh300/go-microservice/shared/proto/product"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *ProductHandler) CreateCategory(ctx context.Context, req *productpb.CreateCategoryRequest) (*productpb.CategoryResponse, error) {
	input := &dto.CreateCategoryDTO{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		ParentID:    convert.StringWrapperToPtr(req.ParentId),
	}

	category, err := h.categoryService.CreateCategory(ctx, input)
	if err != nil {
		return nil, err
	}

	return &productpb.CategoryResponse{
		Category: toCategoryResponse(category),
	}, nil
}

func (h *ProductHandler) GetCategoryByID(ctx context.Context, req *productpb.GetCategoryByIDRequest) (*productpb.CategoryResponse, error) {
	category, err := h.categoryService.GetCategoryByID(ctx, req.CategoryId)
	if err != nil {
		return nil, err
	}

	return &productpb.CategoryResponse{
		Category: toCategoryResponse(category),
	}, nil
}

func (h *ProductHandler) GetCategoryBySlug(ctx context.Context, req *productpb.GetCategoryBySlugRequest) (*productpb.CategoryResponse, error) {
	category, err := h.categoryService.GetCategoryBySlug(ctx, req.Slug)
	if err != nil {
		return nil, err
	}

	return &productpb.CategoryResponse{
		Category: toCategoryResponse(category),
	}, nil
}

func (h *ProductHandler) ListCategories(ctx context.Context, req *productpb.ListCategoriesRequest) (*productpb.ListCategoriesResponse, error) {
	categories, err := h.categoryService.ListCategories(ctx, convert.StringWrapperToPtr(req.ParentId))
	if err != nil {
		return nil, err
	}

	pbCategories := make([]*productpb.Category, len(categories))
	for i, c := range categories {
		pbCategories[i] = toCategoryResponse(c)
	}

	return &productpb.ListCategoriesResponse{
		Categories: pbCategories,
		Total:      int64(len(categories)),
	}, nil
}

func (h *ProductHandler) ListRootCategories(ctx context.Context, req *emptypb.Empty) (*productpb.ListCategoriesResponse, error) {
	categories, err := h.categoryService.ListRootCategories(ctx)
	if err != nil {
		return nil, err
	}

	pbCategories := make([]*productpb.Category, len(categories))
	for i, c := range categories {
		pbCategories[i] = toCategoryResponse(c)
	}

	return &productpb.ListCategoriesResponse{
		Categories: pbCategories,
		Total:      int64(len(categories)),
	}, nil
}

func (h *ProductHandler) ListChildCategories(ctx context.Context, req *productpb.ListChildCategoriesRequest) (*productpb.ListCategoriesResponse, error) {
	categories, err := h.categoryService.ListChildCategories(ctx, req.ParentId)
	if err != nil {
		return nil, err
	}

	pbCategories := make([]*productpb.Category, len(categories))
	for i, c := range categories {
		pbCategories[i] = toCategoryResponse(c)
	}

	return &productpb.ListCategoriesResponse{
		Categories: pbCategories,
		Total:      int64(len(categories)),
	}, nil
}

func (h *ProductHandler) UpdateCategory(ctx context.Context, req *productpb.UpdateCategoryRequest) (*productpb.CategoryResponse, error) {
	input := &dto.UpdateCategoryDTO{
		ID:          req.CategoryId,
		ParentID:    convert.StringWrapperToPtr(req.ParentId),
		Name:        convert.StringWrapperToPtr(req.Name),
		Slug:        convert.StringWrapperToPtr(req.Slug),
		Description: convert.StringWrapperToPtr(req.Description),
		ImageURL:    convert.StringWrapperToPtr(req.ImageUrl),
	}

	category, err := h.categoryService.UpdateCategory(ctx, input)
	if err != nil {
		return nil, err
	}

	return &productpb.CategoryResponse{
		Category: toCategoryResponse(category),
	}, nil
}

func (h *ProductHandler) DeleteCategory(ctx context.Context, req *productpb.DeleteCategoryRequest) (*emptypb.Empty, error) {
	if err := h.categoryService.DeleteCategory(ctx, req.CategoryId); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func toCategoryResponse(category *models.Category) *productpb.Category {
	return &productpb.Category{
		Id:          category.ID.String(),
		Name:        category.Name,
		ParentId:    convert.PtrUUIDToStringWrapper(category.ParentID),
		Slug:        category.Slug,
		Description: category.Description,
		ImageUrl:    convert.PtrToStringWrapper(category.ImageURL),
		UpdatedAt:   convert.TimePtrToTimestamp(&category.UpdatedAt),
		CreatedAt:   convert.TimePtrToTimestamp(&category.CreatedAt),
	}
}
