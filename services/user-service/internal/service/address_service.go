package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/shared/pkg/const/contextkeys"
	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/dto/request"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository"
	"go.uber.org/zap"
)

type AddressService struct {
	userRepo    repository.UserRepository
	addressRepo repository.AddressRepository
}

func NewAddressService(
	userRepo repository.UserRepository,
	addressRepo repository.AddressRepository,
) *AddressService {
	return &AddressService{
		userRepo:    userRepo,
		addressRepo: addressRepo,
	}
}

func (s *AddressService) CreateUserAddress(ctx context.Context, userID string, req *request.CreateUserAddressRequest) (*models.Address, error) {
	logger, _ := ctx.Value(contextkeys.LoggerKey).(*zap.Logger)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperr.ErrUserNotFound
	}

	var address *models.Address
	err = s.addressRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		address = &models.Address{
			UserID:       userUUID,
			AddressType:  models.AddressType(req.AddressType),
			FullName:     req.FullName,
			Phone:        req.Phone,
			AddressLine1: req.AddressLine1,
			AddressLine2: req.AddressLine2,
			Ward:         req.Ward,
			City:         req.City,
			Country:      req.Country,
		}
		err = s.addressRepo.Create(ctx, address)
		if err != nil {
			return err
		}

		if req.IsDefault {
			err = s.addressRepo.SetDefaultAddress(ctx, userUUID, address.ID)
			if err != nil {
				return err
			}
			address.IsDefault = req.IsDefault
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	logger.Info("Created new address", zap.String("addressID", address.ID.String()), zap.String("userID", userID))
	return address, nil
}

func (s *AddressService) ListUserAddresses(ctx context.Context, userID string) ([]*models.Address, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	addresses, err := s.addressRepo.ListByUserID(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	return addresses, nil
}

func (s *AddressService) GetUserAddress(ctx context.Context, userID string, addressID string) (*models.Address, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	addressUUID, err := uuid.Parse(addressID)
	if err != nil {
		return nil, err
	}

	address, err := s.addressRepo.FindByIDAndUserID(ctx, addressUUID, userUUID)
	if err != nil {
		return nil, err
	}
	if address == nil {
		return nil, apperr.ErrAddressNotFound
	}

	return address, nil
}

func (s *AddressService) UpdateUserAddress(ctx context.Context, userID string, addressID string, req *request.UpdateAddressRequest) (*models.Address, error) {
	logger, _ := ctx.Value(contextkeys.LoggerKey).(*zap.Logger)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	addressUUID, err := uuid.Parse(addressID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperr.ErrUserNotFound
	}

	if !user.IsActive() {
		return nil, apperr.ErrAccountInactive
	}

	address, err := s.addressRepo.FindByIDAndUserID(ctx, addressUUID, userUUID)
	if err != nil {
		return nil, err
	}
	if address == nil {
		return nil, apperr.ErrAddressNotFound
	}

	if req.FullName != nil {
		address.FullName = *req.FullName
	}
	if req.Phone != nil {
		address.Phone = *req.Phone
	}
	if req.AddressLine1 != nil {
		address.AddressLine1 = *req.AddressLine1
	}
	if req.AddressLine2 != nil {
		address.AddressLine2 = *req.AddressLine2
	}
	if req.Ward != nil {
		address.Ward = *req.Ward
	}
	if req.City != nil {
		address.City = *req.City
	}
	if req.Country != nil {
		address.Country = *req.Country
	}

	err = s.addressRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		if req.IsDefault != nil && *req.IsDefault {
			err = s.addressRepo.SetDefaultAddress(ctx, userUUID, address.ID)
			if err != nil {
				return err
			}
			address.IsDefault = *req.IsDefault

		}

		err = s.addressRepo.Update(ctx, address)
		if err != nil {
			return err
		}

		return nil
	})

	logger.Info("Updated address", zap.String("addressID", address.ID.String()), zap.String("userID", userID))
	return address, nil
}

func (s *AddressService) DeleteUserAddress(ctx context.Context, userID string, addressID string) error {
	logger, _ := ctx.Value(contextkeys.LoggerKey).(*zap.Logger)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	addressUUID, err := uuid.Parse(addressID)
	if err != nil {
		return err
	}

	address, err := s.addressRepo.FindByIDAndUserID(ctx, addressUUID, userUUID)
	if err != nil {
		return err
	}
	if address == nil {
		return apperr.ErrAddressNotFound
	}

	err = s.addressRepo.Delete(ctx, addressUUID)
	if err != nil {
		return err
	}

	logger.Info("Deleted address", zap.String("addressID", addressID), zap.String("userID", userID))
	return nil
}
