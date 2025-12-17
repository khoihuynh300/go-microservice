package grpchandler

import (
	"context"

	userpb "github.com/khoihuynh300/go-microservice/shared/proto/user"
	"github.com/khoihuynh300/go-microservice/user-service/dto/request"
	"github.com/khoihuynh300/go-microservice/user-service/internal/service"
)

type UserHandler struct {
	userpb.UnimplementedUserServiceServer
	authService *service.AuthService
}

func NewUserHandler(authService *service.AuthService) *UserHandler {
	return &UserHandler{
		authService: authService,
	}
}

func (s *UserHandler) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	registerReq := &request.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Phone:    req.Phone,
	}

	user, err := s.authService.Register(ctx, registerReq)
	if err != nil {
		return nil, err
	}

	return &userpb.RegisterResponse{
		UserId: user.ID.String(),
	}, nil
}

func (s *UserHandler) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.TokenResponse, error) {
	loginReq := &request.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	_, accessToken, refreshToken, err := s.authService.Login(ctx, loginReq)
	if err != nil {
		return nil, err
	}

	return &userpb.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserHandler) Refresh(ctx context.Context, req *userpb.RefreshRequest) (*userpb.TokenResponse, error) {
	accessToken, refreshToken, err := s.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &userpb.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
