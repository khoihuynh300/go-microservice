package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/shared/pkg/contextkeys"
	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
	"github.com/khoihuynh300/go-microservice/user-service/internal/config"
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
	cfg               *config.Config
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	registryTokenRepo repository.RegistryTokenRepository,
	passwordHasher passwordhasher.PasswordHasher,
	jwtService *jwtprovider.JwtService,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
		refreshTokenRepo:  refreshTokenRepo,
		registryTokenRepo: registryTokenRepo,
		passwordHasher:    passwordHasher,
		jwtService:        jwtService,
		cfg:               cfg,
	}
}

func (s *AuthService) Register(ctx context.Context, req *request.RegisterRequest) (*models.User, error) {
	logger, _ := ctx.Value(contextkeys.LoggerKey).(*zap.Logger)

	existedUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existedUser != nil {
		return nil, apperr.ErrEmailAlreadyExists
	}

	hashedPassword, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:          req.Email,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Phone:          &req.Phone,
		Status:         models.UserStatusPending,
	}

	err = s.userRepo.WithinTransaction(ctx, func(ctx context.Context) error {

		if err := s.userRepo.Create(ctx, user); err != nil {
			return err
		}

		verifyToken := uuid.New().String()
		err = s.registryTokenRepo.Create(
			ctx,
			hashToken(verifyToken),
			user.ID,
			time.Now().Add(s.cfg.RegistryTokenExpiry),
		)
		if err != nil {
			return err
		}
		// TODO: publish event to send verification email with verifyToken

		logger.Info("Register success",
			zap.String("user_id", user.ID.String()),
		)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	logger, _ := ctx.Value(contextkeys.LoggerKey).(*zap.Logger)

	tokenHash := hashToken(token)

	return s.userRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		registryToken, err := s.registryTokenRepo.GetByToken(ctx, tokenHash)
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
			return apperr.ErrUserNotFound
		}
		if user.IsEmailVerified() {
			logger.Info("Email already verified",
				zap.String("user_id", user.ID.String()),
			)
			return apperr.ErrEmailAlreadyVerified
		}

		user.Status = models.UserStatusActive
		now := time.Now()
		user.EmailVerifiedAt = &now
		if err := s.userRepo.Update(ctx, user); err != nil {
			return err
		}

		err = s.registryTokenRepo.MarkTokenAsUsed(ctx, tokenHash)
		if err != nil {
			return err
		}

		logger.Info("Email verification success",
			zap.String("user_id", user.ID.String()),
		)

		return nil
	})
}

func (s *AuthService) ResendVerificationEmail(ctx context.Context, email string) error {
	logger, _ := ctx.Value(contextkeys.LoggerKey).(*zap.Logger)

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return apperr.ErrUserNotFound
	}

	if user.IsEmailVerified() {
		logger.Info("Resend verification email skipped: account already active",
			zap.String("user_id", user.ID.String()),
		)
		return apperr.ErrEmailAlreadyVerified
	}

	return s.registryTokenRepo.WithinTransaction(ctx, func(ctx context.Context) error {
		err = s.registryTokenRepo.InvalidateAllUserTokens(ctx, user.ID)
		if err != nil {
			return err
		}

		verifyToken := uuid.New().String()
		err = s.registryTokenRepo.Create(
			ctx,
			hashToken(verifyToken),
			user.ID,
			time.Now().Add(s.cfg.RegistryTokenExpiry),
		)

		if err != nil {
			return err
		}

		logger.Info("Resend verification email success",
			zap.String("user_id", user.ID.String()),
		)

		return nil
	})
}

func (s *AuthService) Login(ctx context.Context, req *request.LoginRequest) (*models.User, string, string, error) {
	logger, _ := ctx.Value(contextkeys.LoggerKey).(*zap.Logger)

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", "", err
	}
	if user == nil {
		logger.Warn("Login failed: invalid credentials")
		return nil, "", "", apperr.ErrInvalidCredentials
	}

	if !s.passwordHasher.Compare(user.HashedPassword, req.Password) {
		logger.Warn("Login failed: invalid credentials")
		return nil, "", "", apperr.ErrInvalidCredentials
	}

	if !user.IsActive() {
		logger.Warn("Login failed: account is inactive",
			zap.String("user_id", user.ID.String()),
		)
		return nil, "", "", apperr.ErrAccountInactive
	}

	accessToken, refreshToken, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, "", "", err
	}

	logger.Info("Login success", zap.String("user_id", user.ID.String()))
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
		return "", "", apperr.ErrAccountInactive
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
