package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/khoihuynh300/go-microservice/user-service/dto/request"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository"
	"github.com/khoihuynh300/go-microservice/user-service/internal/security/jwtprovider"
	passwordhasher "github.com/khoihuynh300/go-microservice/user-service/internal/security/password"
)

type AuthService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	passwordHasher   passwordhasher.PasswordHasher
	jwtService       *jwtprovider.JwtService
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	passwordHasher passwordhasher.PasswordHasher,
	jwtService *jwtprovider.JwtService,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		passwordHasher:   passwordHasher,
		jwtService:       jwtService,
	}
}

func (s *AuthService) Register(ctx context.Context, req *request.RegisterRequest) (*domain.User, error) {
	existedUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existedUser != nil {
		return nil, errors.New("email already exists")
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

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req *request.LoginRequest) (*domain.User, string, string, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", "", err
	}
	if user == nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	if !s.passwordHasher.Compare(user.HashedPassword, req.Password) {
		return nil, "", "", errors.New("invalid email or password")
	}

	if !user.IsActive() {
		return nil, "", "", errors.New("user is inactive or banned")
	}

	accessToken, refreshToken, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, error) {
	claims, err := s.jwtService.VerifyRefreshToken(refreshTokenStr)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	refreshTokenModel, err := s.refreshTokenRepo.FindByToken(ctx, hashRefreshToken(refreshTokenStr))
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	userID := claims.UserID
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", errors.New("user not found")
	}

	if !user.IsActive() {
		return "", "", errors.New("user is inactive or banned")
	}

	s.refreshTokenRepo.DeleteByID(ctx, refreshTokenModel.ID)

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
