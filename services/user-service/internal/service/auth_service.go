package service

import (
	"context"

	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/dto/request"
)

type AuthService interface {
	Register(ctx context.Context, req *request.RegisterRequest) (*models.User, error)
	VerifyEmail(ctx context.Context, token string) error
	ResendVerificationEmail(ctx context.Context, email string) error
	Login(ctx context.Context, req *request.LoginRequest) (*models.User, string, string, error)
	RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, error)
	generateTokenPair(ctx context.Context, user *models.User) (string, string, error)
	ChangePassword(ctx context.Context, userID string, req *request.ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token string, newPassword string) error
}
