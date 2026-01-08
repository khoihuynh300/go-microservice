package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
	zaplogger "github.com/khoihuynh300/go-microservice/shared/pkg/logger"
	"github.com/khoihuynh300/go-microservice/user-service/internal/caching"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/dto/request"
	"github.com/khoihuynh300/go-microservice/user-service/internal/events/publisher"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository"
	"github.com/khoihuynh300/go-microservice/user-service/internal/security/jwtprovider"
	passwordhasher "github.com/khoihuynh300/go-microservice/user-service/internal/security/password"
	"github.com/khoihuynh300/go-microservice/user-service/internal/utils"
	"go.uber.org/zap"
)

type authService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	tokenCache       *caching.TokenCache
	passwordHasher   passwordhasher.PasswordHasher
	jwtService       jwtprovider.JwtProvider
	eventPublisher   publisher.EventPublisher
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	tokenCache *caching.TokenCache,
	passwordHasher passwordhasher.PasswordHasher,
	jwtService jwtprovider.JwtProvider,
	eventPublisher publisher.EventPublisher,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenCache:       tokenCache,
		passwordHasher:   passwordHasher,
		jwtService:       jwtService,
		eventPublisher:   eventPublisher,
	}
}

func (s *authService) Register(ctx context.Context, req *request.RegisterRequest) (*models.User, error) {
	logger := zaplogger.FromContext(ctx)

	existedUser, err := s.userRepo.GetByEmail(ctx, req.Email)
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
		Status:         models.UserStatusPending,
	}

	err = s.userRepo.WithinTransaction(ctx, func(ctx context.Context) error {

		if err := s.userRepo.Create(ctx, user); err != nil {
			return err
		}

		verifyToken, err := s.tokenCache.SetEmailVerifyToken(ctx, user.Email)
		if err != nil {
			return err
		}

		err = s.eventPublisher.PublishVerifyEmail(ctx, user, verifyToken)
		if err != nil {
			return err
		}

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

func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	logger := zaplogger.FromContext(ctx)

	email, err := s.tokenCache.VerifyEmailToken(ctx, token)
	if err != nil {
		if errors.Is(err, caching.ErrTokenInvalidOrExpired) {
			return apperr.ErrTokenInvalidOrExpired
		}
		return err
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
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

	rowEffected, err := s.userRepo.VerifyEmail(ctx, user.ID)
	if err != nil {
		return err
	}
	if rowEffected == 0 {
		return apperr.ErrUserNotFound
	}

	rowEffected, err = s.userRepo.UpdateStatus(ctx, user.ID, models.UserStatusActive)
	if err != nil {
		return err
	}
	if rowEffected == 0 {
		return apperr.ErrUserNotFound
	}

	err = s.eventPublisher.PublishEmailVerifySuccess(ctx, user.Email)
	if err != nil {
		return err
	}

	logger.Info("Email verification success",
		zap.String("user_id", user.ID.String()),
	)

	return nil
}

func (s *authService) ResendVerificationEmail(ctx context.Context, email string) error {
	logger := zaplogger.FromContext(ctx)

	user, err := s.userRepo.GetByEmail(ctx, email)
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

	verifyToken, err := s.tokenCache.SetEmailVerifyToken(ctx, user.Email)
	if err != nil {
		return err
	}
	err = s.eventPublisher.PublishVerifyEmail(ctx, user, verifyToken)
	if err != nil {
		return err
	}

	logger.Info("Resend verification email success",
		zap.String("user_id", user.ID.String()),
	)

	return nil
}

func (s *authService) Login(ctx context.Context, req *request.LoginRequest) (*models.User, string, string, error) {
	logger := zaplogger.FromContext(ctx)

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
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

func (s *authService) RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, error) {
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

	refreshTokenModel, err := s.refreshTokenRepo.GetByToken(ctx, utils.HashToken(refreshTokenStr))
	if err != nil {
		return "", "", err
	}
	if refreshTokenModel == nil {
		return "", "", apperr.ErrTokenInvalid
	}

	userID := uuid.MustParse(claims.Subject)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", apperr.ErrTokenInvalid
	}

	if !user.IsActive() {
		return "", "", apperr.ErrAccountInactive
	}

	rowEffected, err := s.refreshTokenRepo.DeleteByID(ctx, refreshTokenModel.ID)
	if err != nil {
		return "", "", err
	}
	if rowEffected == 0 {
		return "", "", apperr.ErrTokenInvalid
	}

	return s.generateTokenPair(ctx, user)
}

func (s *authService) generateTokenPair(ctx context.Context, user *models.User) (string, string, error) {
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
		TokenHash: utils.HashToken(refreshToken),
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshTTL()),
	}

	err = s.refreshTokenRepo.Create(ctx, refreshTokenModel)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *authService) ChangePassword(ctx context.Context, userID string, req *request.ChangePasswordRequest) error {
	logger := zaplogger.FromContext(ctx)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(ctx, userUUID)
	if err != nil {
		return err
	}
	if user == nil {
		return apperr.ErrUserNotFound
	}

	if !s.passwordHasher.Compare(user.HashedPassword, req.CurrentPassword) {
		logger.Warn("Change password failed: invalid current password",
			zap.String("user_id", user.ID.String()),
		)
		return apperr.ErrInvalidCurrentPassword
	}

	newHashedPassword, err := s.passwordHasher.Hash(req.NewPassword)
	if err != nil {
		return err
	}

	rowEffected, err := s.userRepo.UpdatePassword(ctx, user.ID, newHashedPassword)
	if err != nil {
		return err
	}
	if rowEffected == 0 {
		return apperr.ErrUserNotFound
	}

	logger.Info("Change password success",
		zap.String("user_id", user.ID.String()),
	)
	return nil
}

func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	logger := zaplogger.FromContext(ctx)
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return apperr.ErrUserNotFound
	}

	token, err := s.tokenCache.SetPasswordResetToken(ctx, user.Email)
	if err != nil {
		return err
	}

	err = s.eventPublisher.PublishForgotPassword(ctx, user, token)
	if err != nil {
		return err
	}

	logger.Info("Forgot password email sent",
		zap.String("user_id", user.ID.String()),
	)

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, token string, newPassword string) error {
	logger := zaplogger.FromContext(ctx)

	email, err := s.tokenCache.VerifyPasswordResetToken(ctx, token)
	if err != nil {
		if errors.Is(err, caching.ErrTokenInvalidOrExpired) {
			return apperr.ErrTokenInvalidOrExpired
		}
		return err
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return apperr.ErrUserNotFound
	}

	newHashedPassword, err := s.passwordHasher.Hash(newPassword)
	if err != nil {
		return err
	}

	rowEffected, err := s.userRepo.UpdatePassword(ctx, user.ID, newHashedPassword)
	if err != nil {
		return err
	}
	if rowEffected == 0 {
		return apperr.ErrUserNotFound
	}

	err = s.eventPublisher.PublishPasswordResetSuccess(ctx, user.Email)
	if err != nil {
		return err
	}

	logger.Info("Reset password success",
		zap.String("user_id", user.ID.String()),
	)
	return nil
}
