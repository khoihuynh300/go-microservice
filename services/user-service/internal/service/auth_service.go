package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
	"github.com/khoihuynh300/go-microservice/user-service/internal/config"
	domainerr "github.com/khoihuynh300/go-microservice/user-service/internal/domain/errors"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/dto/request"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository"
	"github.com/khoihuynh300/go-microservice/user-service/internal/security/jwtprovider"
	passwordhasher "github.com/khoihuynh300/go-microservice/user-service/internal/security/password"
	"go.uber.org/zap"
)

type AuthService struct {
	userRepo          repository.UserRepository
	refreshTokenRepo  repository.RefreshTokenRepository
	registryTokenRepo repository.RegistryTokenRepository
	passwordHasher    passwordhasher.PasswordHasher
	jwtService        *jwtprovider.JwtService
	logger            *zap.Logger
	cfg               *config.Config
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	registryTokenRepo repository.RegistryTokenRepository,
	passwordHasher passwordhasher.PasswordHasher,
	jwtService *jwtprovider.JwtService,
	logger *zap.Logger,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
		refreshTokenRepo:  refreshTokenRepo,
		registryTokenRepo: registryTokenRepo,
		passwordHasher:    passwordHasher,
		jwtService:        jwtService,
		logger:            logger,
		cfg:               cfg,
	}
}

func (s *AuthService) Register(ctx context.Context, req *request.RegisterRequest) (*models.User, error) {
	existedUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existedUser != nil {
		return nil, domainerr.ErrEmailAlreadyExists
	}

	hashedPassword, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:          req.Email,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Phone:          req.Phone,
		Status:         models.UserStatusPending,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	verifyToken := uuid.New().String()
	err = s.registryTokenRepo.Create(
		ctx,
		hashToken(verifyToken),
		user.ID,
		time.Now().Add(s.cfg.RegistryTokenExpiry),
	)
	if err != nil {
		return nil, err
	}
	// TODO: publish event to send verification email with verifyToken

	s.logger.Info("Register success",
		zap.String("user_id", user.ID.String()),
	)
	return user, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	tokenHash := hashToken(token)

	registryToken, err := s.registryTokenRepo.GetUserIdByToken(ctx, tokenHash)
	if err != nil {
		return err
	}
	if registryToken == nil {
		return apperr.ErrTokenInvalid
	}

	if !registryToken.IsValid() {
		if registryToken.IsExpired() {
			return apperr.ErrTokenExpired
		}
		return apperr.ErrTokenInvalid
	}

	user, err := s.userRepo.FindByID(ctx, registryToken.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return apperr.ErrTokenInvalid
	}

	user.Status = models.UserStatusActive
	now := time.Now()
	user.EmailVerifiedAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	s.logger.Info("Email verification success",
		zap.String("user_id", user.ID.String()),
	)
	return s.registryTokenRepo.MarkTokenAsUsed(ctx, tokenHash)
}

func (s *AuthService) Login(ctx context.Context, req *request.LoginRequest) (*models.User, string, string, error) {
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
		return nil, "", "", domainerr.ErrAccountInactive
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

	refreshTokenModel, err := s.refreshTokenRepo.FindByToken(ctx, hashToken(refreshTokenStr))
	if err != nil {
		return "", "", err
	}
	if refreshTokenModel == nil {
		return "", "", apperr.ErrTokenInvalid
	}

	userID := uuid.MustParse(claims.Subject)

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", apperr.ErrTokenInvalid
	}

	if !user.IsActive() {
		return "", "", domainerr.ErrAccountInactive
	}

	if err := s.refreshTokenRepo.DeleteByID(ctx, refreshTokenModel.ID); err != nil {
		return "", "", err
	}

	return s.generateTokenPair(ctx, user)
}

func (s *AuthService) generateTokenPair(ctx context.Context, user *models.User) (string, string, error) {
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return "", "", err
	}

	refreshTokenModel := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshTTL()),
	}

	err = s.refreshTokenRepo.Save(ctx, refreshTokenModel)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}
