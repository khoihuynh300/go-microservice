package grpchandler

import (
	"context"

	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/dto"
	"github.com/khoihuynh300/go-microservice/product-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/product-service/internal/utils/convert"
	productpb "github.com/khoihuynh300/go-microservice/shared/proto/product"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ProductHandler) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.ProductResponse, error) {
	input := &dto.CreateProductDTO{
		Name:        req.Name,
		SKU:         req.Sku,
		Slug:        req.Slug,
		Description: req.Description,
		CategoryID:  req.CategoryId,
		Price:       req.Price,
	}

	product, err := h.productService.CreateProduct(ctx, input)
	if err != nil {
		return nil, err
	}

	return &productpb.ProductResponse{
		Product: toProductResponse(product),
	}, nil
}

func (h *ProductHandler) GetProductByID(ctx context.Context, req *productpb.GetProductByIDRequest) (*productpb.ProductResponse, error) {
	product, err := h.productService.GetProductByID(ctx, req.ProductId)
	if err != nil {
		return nil, err
	}

	return &productpb.ProductResponse{
		Product: toProductResponse(product),
	}, nil
}

func (h *ProductHandler) GetProductBySlug(ctx context.Context, req *productpb.GetProductBySlugRequest) (*productpb.ProductResponse, error) {
	product, err := h.productService.GetProductBySlug(ctx, req.Slug)
	if err != nil {
		return nil, err
	}

	return &productpb.ProductResponse{
		Product: toProductResponse(product),
	}, nil
}

func (h *ProductHandler) GetProductBySKU(ctx context.Context, req *productpb.GetProductBySKURequest) (*productpb.ProductResponse, error) {
	product, err := h.productService.GetProductBySKU(ctx, req.Sku)
	if err != nil {
		return nil, err
	}

	return &productpb.ProductResponse{
		Product: toProductResponse(product),
	}, nil
}

func (h *ProductHandler) ListProducts(ctx context.Context, req *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error) {
	input := &dto.ListProductsDTO{
		Page:       req.Page,
		PageSize:   req.PageSize,
		CategoryID: convert.StringWrapperToPtr(req.CategoryId),
	}

	products, total, err := h.productService.ListProducts(ctx, input)
	if err != nil {
		return nil, err
	}

	pbProducts := make([]*productpb.ProductSummary, len(products))
	for i, p := range products {
		pbProducts[i] = toProductSummaryResponse(p)
	}

	totalPages := int32(total) / req.PageSize
	if int32(total)%req.PageSize > 0 {
		totalPages++
	}

	return &productpb.ListProductsResponse{
		Products:   pbProducts,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (h *ProductHandler) SearchProducts(ctx context.Context, req *productpb.SearchProductsRequest) (*productpb.ListProductsResponse, error) {
	input := &dto.SearchProductsDTO{
		SearchQuery: req.Search,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}

	products, total, err := h.productService.SearchProducts(ctx, input)
	if err != nil {
		return nil, err
	}

	pbProducts := make([]*productpb.ProductSummary, len(products))
	for i, p := range products {
		pbProducts[i] = toProductSummaryResponse(p)
	}

	totalPages := int32(total) / req.PageSize
	if int32(total)%req.PageSize > 0 {
		totalPages++
	}

	return &productpb.ListProductsResponse{
		Products:   pbProducts,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.ProductResponse, error) {
	var images *[]string
	if req.Images != nil {
		images = &req.Images.Images
	}

	input := &dto.UpdateProductDTO{
		ID:          req.ProductId,
		Name:        convert.StringWrapperToPtr(req.Name),
		SKU:         convert.StringWrapperToPtr(req.Sku),
		Slug:        convert.StringWrapperToPtr(req.Slug),
		Description: convert.StringWrapperToPtr(req.Description),
		CategoryID:  convert.StringWrapperToPtr(req.CategoryId),
		Price:       convert.DoubleWrapperToPtr(req.Price),
		Thumbnail:   convert.StringWrapperToPtr(req.Thumbnail),
		Images:      images,
	}

	product, err := h.productService.UpdateProduct(ctx, input)
	if err != nil {
		return nil, err
	}

	return &productpb.ProductResponse{
		Product: toProductResponse(product),
	}, nil
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*emptypb.Empty, error) {
	if err := h.productService.DeleteProduct(ctx, req.ProductId); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (h *ProductHandler) GetProductsByIDs(ctx context.Context, req *productpb.GetProductsByIDsRequest) (*productpb.ListProductsResponse, error) {
	input := &dto.GetProductsByIDsDTO{
		IDs: req.Ids,
	}

	products, err := h.productService.GetProductsByIDs(ctx, input)
	if err != nil {
		return nil, err
	}

	pbProducts := make([]*productpb.ProductSummary, len(products))
	for i, p := range products {
		pbProducts[i] = toProductSummaryResponse(p)
	}

	return &productpb.ListProductsResponse{
		Products: pbProducts,
		Total:    int64(len(products)),
	}, nil
}

func toProductResponse(product *models.Product) *productpb.Product {
	return &productpb.Product{
		Id:          product.ID.String(),
		Name:        product.Name,
		Sku:         product.SKU,
		Slug:        product.Slug,
		Description: product.Description,
		CategoryId:  product.CategoryID.String(),
		Price:       product.Price,
		Thumbnail:   convert.GenericStringPtrToWrapper(product.Thumbnail),
		CreatedAt:   timestamppb.New(product.CreatedAt),
		UpdatedAt:   timestamppb.New(product.UpdatedAt),
		Images:      product.Images,
	}
}

func toProductSummaryResponse(product *models.Product) *productpb.ProductSummary {
	return &productpb.ProductSummary{
		Id:         product.ID.String(),
		Name:       product.Name,
		Sku:        product.SKU,
		Slug:       product.Slug,
		CategoryId: product.CategoryID.String(),
		Price:      product.Price,
		Thumbnail:  convert.GenericStringPtrToWrapper(product.Thumbnail),
		CreatedAt:  timestamppb.New(product.CreatedAt),
		UpdatedAt:  timestamppb.New(product.UpdatedAt),
	}
}
