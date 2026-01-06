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

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperr.ErrUserNotFound
	}

	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, userID string, updateData *request.UpdateUserRequest) (*models.User, error) {
	logger, _ := ctx.Value(contextkeys.LoggerKey).(*zap.Logger)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperr.ErrUserNotFound
	}

	if updateData.FullName != nil {
		user.FullName = *updateData.FullName
	}
	if updateData.DateOfBirth != nil {
		user.DateOfBirth = updateData.DateOfBirth
	}
	if updateData.Gender != nil {
		updateGender := models.Gender(*updateData.Gender)
		user.Gender = &updateGender
	}

	rowEffected, err := s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}
	if rowEffected == 0 {
		return nil, apperr.ErrUserNotFound
	}

	logger.Info("Updated user profile", zap.String("userID", userID))
	return user, nil
}
