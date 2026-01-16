package dto

import "github.com/khoihuynh300/go-microservice/product-service/internal/domain/models"

type CreateProductDTO struct {
	Name        string
	SKU         string
	Slug        string
	Description string
	CategoryID  string
	Price       float64
}

type UpdateProductDTO struct {
	ID          string
	Name        *string
	SKU         *string
	Slug        *string
	Description *string
	CategoryID  *string
	Price       *float64
	Thumbnail   *string
	Images      *[]string
}

type SearchProductsDTO struct {
	SearchQuery string
	Page        int32
	PageSize    int32
}

type ListProductsDTO struct {
	CategoryID *string
	Page       int32
	PageSize   int32
}

type GetProductsByIDsDTO struct {
	IDs []string
}

type AddProductImageDTO struct {
	ProductID string
	ImageURL  string
	Position  int32
}

type DeleteProductImageDTO struct {
	ProductID string
	ImageID   string
}

type ListProductsResult struct {
	Products   []*models.Product
	Total      int64
	Page       int32
	PageSize   int32
	TotalPages int32
}
