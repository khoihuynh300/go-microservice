package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
	"github.com/khoihuynh300/go-microservice/user-service/dto/request"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository"
	"github.com/khoihuynh300/go-microservice/user-service/internal/security/jwtprovider"
	passwordhasher "github.com/khoihuynh300/go-microservice/user-service/internal/security/password"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

type AuthService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	passwordHasher   passwordhasher.PasswordHasher
	jwtService       *jwtprovider.JwtService
	logger           *zap.Logger
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	passwordHasher passwordhasher.PasswordHasher,
	jwtService *jwtprovider.JwtService,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		passwordHasher:   passwordHasher,
		jwtService:       jwtService,
		logger:           logger,
	}
}

func (s *AuthService) Register(ctx context.Context, req *request.RegisterRequest) (*domain.User, error) {
	existedUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existedUser != nil {
		return nil, apperr.New(apperr.CodeAlreadyExists, "Email already exists", codes.AlreadyExists)
	}

	hashedPassword, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:          req.Email,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Phone:          req.Phone,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	s.logger.Info("Register success",
		zap.String("user_id", user.ID.String()),
	)
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req *request.LoginRequest) (*domain.User, string, string, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", "", err
	}
	if user == nil {
		s.logger.Warn("Login failed: invalid credentials")
		return nil, "", "", apperr.ErrInvalidCredentials
	}

	if !s.passwordHasher.Compare(user.HashedPassword, req.Password) {
		s.logger.Warn("Login failed: invalid credentials")
		return nil, "", "", apperr.ErrInvalidCredentials
	}

	if !user.IsActive() {
		s.logger.Warn("Login failed: account is inactive",
			zap.String("user_id", user.ID.String()),
		)
		return nil, "", "", apperr.New(apperr.CodeUnauthorized, "Account is inactive", codes.PermissionDenied)
	}

	accessToken, refreshToken, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	s.logger.Info("Login success", zap.String("user_id", user.ID.String()))
	return user, accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, error) {
	claims, err := s.jwtService.VerifyRefreshToken(refreshTokenStr)
	if err != nil {
		if errors.Is(err, jwtprovider.ErrTokenExpired) {
			return "", "", apperr.ErrTokenExpired
		}
		if errors.Is(err, jwtprovider.ErrTokenInvalid) {
			return "", "", apperr.ErrTokenInvalid
		}
		return "", "", err
	}

	refreshTokenModel, err := s.refreshTokenRepo.FindByToken(ctx, hashRefreshToken(refreshTokenStr))
	if err != nil {
		return "", "", err
	}
	if refreshTokenModel == nil {
		return "", "", apperr.ErrTokenInvalid
	}

	userID := claims.UserID
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", apperr.ErrTokenInvalid
	}

	if !user.IsActive() {
		return "", "", apperr.New(apperr.CodeUnauthorized, "Account is inactive", codes.PermissionDenied)
	}

	if err := s.refreshTokenRepo.DeleteByID(ctx, refreshTokenModel.ID); err != nil {
		return "", "", err
	}

	return s.generateTokenPair(ctx, user)
}

func (s *AuthService) generateTokenPair(ctx context.Context, user *domain.User) (string, string, error) {
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return "", "", err
	}

	refreshTokenModel := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashRefreshToken(refreshToken),
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshTTL()),
	}

	err = s.refreshTokenRepo.Save(ctx, refreshTokenModel)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}
