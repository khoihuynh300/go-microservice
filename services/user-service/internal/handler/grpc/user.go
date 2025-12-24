package grpchandler

import (
	"context"

	userpb "github.com/khoihuynh300/go-microservice/shared/proto/user"
	domainerr "github.com/khoihuynh300/go-microservice/user-service/internal/domain/errors"
	"github.com/khoihuynh300/go-microservice/user-service/internal/dto/request"
	"github.com/khoihuynh300/go-microservice/user-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	userpb.UnimplementedUserServiceServer
	authService *service.AuthService
	userService *service.UserService
}

func NewUserHandler(authService *service.AuthService, userService *service.UserService) *UserHandler {
	return &UserHandler{
		authService: authService,
		userService: userService,
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

func (s *UserHandler) VerifyEmail(ctx context.Context, req *userpb.VerifyEmailRequest) (*emptypb.Empty, error) {
	err := s.authService.VerifyEmail(ctx, req.VerifyToken)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *UserHandler) ResendVerificationEmail(ctx context.Context, req *userpb.ResendVerificationEmailRequest) (*emptypb.Empty, error) {
	err := s.authService.ResendVerificationEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
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

func (s *UserHandler) GetMe(ctx context.Context, req *emptypb.Empty) (*userpb.GetUserResponse, error) {

	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	user, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domainerr.ErrUserNotFound
	}

	var dateOfBirth *timestamppb.Timestamp = nil
	if user.DateOfBirth != nil {
		dateOfBirth = timestamppb.New(*user.DateOfBirth)
	}

	return &userpb.GetUserResponse{
		User: &userpb.User{
			Id:          user.ID.String(),
			FullName:    user.FullName,
			Email:       user.Email,
			Phone:       user.Phone,
			DateOfBirth: dateOfBirth,
			AvatarUrl:   user.AvatarURL,
			Gender:      string(user.Gender),
			Status:      string(user.Status),
		},
	}, nil
}

// func (s *UserHandler) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
// }

// func (s *UserHandler) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
// }

// func (s *UserHandler) ChangePassword(ctx context.Context, req *userpb.ChangePasswordRequest) (*emptypb.Empty, error) {
// }

// func (s *UserHandler) ForgotPassword(ctx context.Context, req *userpb.ChangePasswordRequest) (*emptypb.Empty, error) {
// }

// func (s *UserHandler) ResetPassword(ctx context.Context, req *userpb.ChangePasswordRequest) (*emptypb.Empty, error) {
// }

// func (s *UserHandler) CreateUserAddress(ctx context.Context, req *userpb.ChangePasswordRequest) (*emptypb.Empty, error) {
// }

// func (s *UserHandler) GetUserAddresses(ctx context.Context, req *userpb.ChangePasswordRequest) (*emptypb.Empty, error) {
// }

// func (s *UserHandler) GetUserAddress(ctx context.Context, req *userpb.ChangePasswordRequest) (*emptypb.Empty, error) {
// }

// func (s *UserHandler) UpdateUserAddress(ctx context.Context, req *userpb.ChangePasswordRequest) (*emptypb.Empty, error) {
// }

// func (s *UserHandler) DeleteUserAddress(ctx context.Context, req *userpb.ChangePasswordRequest) (*emptypb.Empty, error) {
// }

// func (s *UserHandler) SetDefaultUserAddress(ctx context.Context, req *userpb.ChangePasswordRequest) (*emptypb.Empty, error) {
// }
