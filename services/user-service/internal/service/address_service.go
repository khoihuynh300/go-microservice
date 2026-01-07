package service

import (
	"context"

	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/dto/request"
)

type AddressService interface {
	CreateUserAddress(ctx context.Context, userID string, req *request.CreateUserAddressRequest) (*models.Address, error)
	ListUserAddresses(ctx context.Context, userID string) ([]*models.Address, error)
	GetUserAddress(ctx context.Context, userID string, addressID string) (*models.Address, error)
	UpdateUserAddress(ctx context.Context, userID string, addressID string, req *request.UpdateAddressRequest) (*models.Address, error)
	DeleteUserAddress(ctx context.Context, userID string, addressID string) error
}
