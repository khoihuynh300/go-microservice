package grpchandler

import (
	"github.com/khoihuynh300/go-microservice/product-service/internal/service"
	productpb "github.com/khoihuynh300/go-microservice/shared/proto/product"
)

type ProductHandler struct {
	productpb.UnimplementedProductServiceServer
	productService  service.ProductService
	categoryService service.CategoryService
}

func NewProductHandler(
	productService service.ProductService,
	categoryService service.CategoryService,
) *ProductHandler {
	return &ProductHandler{
		productService:  productService,
		categoryService: categoryService,
	}
}
