package service

import (
	"context"

	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/dto/request"
)

type UserService interface {
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	UpdateUser(ctx context.Context, userID string, updateData *request.UpdateUserRequest) (*models.User, error)
}
