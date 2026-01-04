package grpchandler

import (
	"context"

	"github.com/khoihuynh300/go-microservice/shared/pkg/const/contextkeys"
	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
	userpb "github.com/khoihuynh300/go-microservice/shared/proto/user"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/dto/request"
	"github.com/khoihuynh300/go-microservice/user-service/internal/service"
	"github.com/khoihuynh300/go-microservice/user-service/internal/utils/convert"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserHandler struct {
	userpb.UnimplementedUserServiceServer
	authService    *service.AuthService
	userService    *service.UserService
	addressService *service.AddressService
}

func NewUserHandler(
	authService *service.AuthService,
	userService *service.UserService,
	addressService *service.AddressService,
) *UserHandler {
	return &UserHandler{
		authService:    authService,
		userService:    userService,
		addressService: addressService,
	}
}

func (s *UserHandler) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	registerReq := &request.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
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
	userID, ok := ctx.Value(contextkeys.UserIDKey).(string)
	if !ok {
		return nil, apperr.ErrUnauthenticated
	}

	user, err := s.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperr.ErrUserNotFound
	}

	return &userpb.GetUserResponse{
		User: toUserResponse(user),
	}, nil
}

func (s *UserHandler) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetPublicUserResponse, error) {
	user, err := s.userService.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperr.ErrUserNotFound
	}

	return &userpb.GetPublicUserResponse{
		User: toUserPublicResponse(user),
	}, nil
}

func (s *UserHandler) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	userID, ok := ctx.Value(contextkeys.UserIDKey).(string)
	if !ok {
		return nil, apperr.ErrUnauthenticated
	}

	dob, err := convert.StringPtrToTimePtr(req.DateOfBirth)
	if err != nil {
		return nil, apperr.NewErrValidationFailedWithDetail(
			"date_of_birth",
			apperr.CodeInvalidDateFormat,
			"Date of birth must be in DD-MM-YYYY format",
		)
	}

	updateReq := &request.UpdateUserRequest{
		FullName:    req.FullName,
		DateOfBirth: dob,
		Gender:      req.Gender,
	}

	updatedUser, err := s.userService.UpdateUser(ctx, userID, updateReq)
	if err != nil {
		return nil, err
	}

	return &userpb.UpdateUserResponse{
		User: toUserResponse(updatedUser),
	}, nil
}

func (s *UserHandler) ChangePassword(ctx context.Context, req *userpb.ChangePasswordRequest) (*emptypb.Empty, error) {
	userID, ok := ctx.Value(contextkeys.UserIDKey).(string)
	if !ok {
		return nil, apperr.ErrUnauthenticated
	}

	changePasswordReq := &request.ChangePasswordRequest{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	err := s.authService.ChangePassword(ctx, userID, changePasswordReq)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *UserHandler) ForgotPassword(ctx context.Context, req *userpb.ForgotPasswordRequest) (*emptypb.Empty, error) {
	err := s.authService.ForgotPassword(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *UserHandler) ResetPassword(ctx context.Context, req *userpb.ResetPasswordRequest) (*emptypb.Empty, error) {
	err := s.authService.ResetPassword(ctx, req.ResetToken, req.NewPassword)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *UserHandler) CreateUserAddress(ctx context.Context, req *userpb.CreateUserAddressRequest) (*userpb.CreateUserAddressResponse, error) {
	userID, ok := ctx.Value(contextkeys.UserIDKey).(string)
	if !ok {
		return nil, apperr.ErrUnauthenticated
	}

	createAddrReq := &request.CreateUserAddressRequest{
		AddressType:  req.AddressType,
		FullName:     req.FullName,
		Phone:        req.Phone,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		Ward:         req.Ward,
		City:         req.City,
		Country:      req.Country,
		IsDefault:    req.IsDefault,
	}
	address, err := s.addressService.CreateUserAddress(ctx, userID, createAddrReq)
	if err != nil {
		return nil, err
	}

	return &userpb.CreateUserAddressResponse{
		Address: toAddressResponse(address),
	}, nil
}

func (s *UserHandler) GetUserAddresses(ctx context.Context, req *emptypb.Empty) (*userpb.GetUserAddressesResponse, error) {
	userID, ok := ctx.Value(contextkeys.UserIDKey).(string)
	if !ok {
		return nil, apperr.ErrUnauthenticated
	}

	addresses, err := s.addressService.ListUserAddresses(ctx, userID)
	if err != nil {
		return nil, err
	}

	var addressResponses []*userpb.Address
	for _, addr := range addresses {
		addressResponses = append(addressResponses, toAddressResponse(addr))
	}

	return &userpb.GetUserAddressesResponse{
		Addresses: addressResponses,
	}, nil
}

func (s *UserHandler) GetUserAddress(ctx context.Context, req *userpb.GetUserAddressRequest) (*userpb.GetUserAddressResponse, error) {
	userID, ok := ctx.Value(contextkeys.UserIDKey).(string)
	if !ok {
		return nil, apperr.ErrUnauthenticated
	}

	address, err := s.addressService.GetUserAddress(ctx, userID, req.AddressId)
	if err != nil {
		return nil, err
	}

	return &userpb.GetUserAddressResponse{
		Address: toAddressResponse(address),
	}, nil
}

func (s *UserHandler) UpdateUserAddress(ctx context.Context, req *userpb.UpdateUserAddressRequest) (*userpb.UpdateUserAddressResponse, error) {
	userID, ok := ctx.Value(contextkeys.UserIDKey).(string)
	if !ok {
		return nil, apperr.ErrUnauthenticated
	}

	updateAddrReq := &request.UpdateAddressRequest{
		AddressType:  req.AddressType,
		FullName:     req.FullName,
		Phone:        req.Phone,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		Ward:         req.Ward,
		City:         req.City,
		Country:      req.Country,
		IsDefault:    req.IsDefault,
	}

	address, err := s.addressService.UpdateUserAddress(ctx, userID, req.AddressId, updateAddrReq)
	if err != nil {
		return nil, err
	}

	return &userpb.UpdateUserAddressResponse{
		Address: toAddressResponse(address),
	}, nil
}

func (s *UserHandler) DeleteUserAddress(ctx context.Context, req *userpb.DeleteUserAddressRequest) (*emptypb.Empty, error) {
	userID, ok := ctx.Value(contextkeys.UserIDKey).(string)
	if !ok {
		return nil, apperr.ErrUnauthenticated
	}

	err := s.addressService.DeleteUserAddress(ctx, userID, req.AddressId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func toUserResponse(user *models.User) *userpb.User {
	return &userpb.User{
		Id:          user.ID.String(),
		FullName:    user.FullName,
		Email:       user.Email,
		Phone:       convert.GenericStringPtrToWrapper(user.Phone),
		DateOfBirth: convert.TimePtrToDateStringWrapper(user.DateOfBirth),
		AvatarUrl:   convert.GenericStringPtrToWrapper(user.AvatarURL),
		Gender:      convert.GenericStringPtrToWrapper(user.Gender),
		Status:      string(user.Status),
	}
}

func toUserPublicResponse(user *models.User) *userpb.PublicUserProfile {
	return &userpb.PublicUserProfile{
		Id:        user.ID.String(),
		FullName:  user.FullName,
		AvatarUrl: convert.GenericStringPtrToWrapper(user.AvatarURL),
	}
}

func toAddressResponse(address *models.Address) *userpb.Address {
	return &userpb.Address{
		Id:           address.ID.String(),
		UserId:       address.UserID.String(),
		AddressType:  string(address.AddressType),
		FullName:     address.FullName,
		Phone:        address.Phone,
		AddressLine1: address.AddressLine1,
		AddressLine2: address.AddressLine2,
		Ward:         address.Ward,
		City:         address.City,
		Country:      address.Country,
		IsDefault:    address.IsDefault,
	}
}
