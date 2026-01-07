package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	mock_cache "github.com/khoihuynh300/go-microservice/shared/mocks/cache"
	"github.com/khoihuynh300/go-microservice/shared/pkg/const/contextkeys"
	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
	"github.com/khoihuynh300/go-microservice/user-service/internal/caching"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/dto/request"
	"github.com/khoihuynh300/go-microservice/user-service/internal/security/jwtprovider"
	"github.com/khoihuynh300/go-microservice/user-service/internal/service"
	mock_jwt "github.com/khoihuynh300/go-microservice/user-service/mocks/jwt"
	mock_password_hasher "github.com/khoihuynh300/go-microservice/user-service/mocks/passwordhasher"
	mock_repository "github.com/khoihuynh300/go-microservice/user-service/mocks/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type AuthServiceTestSuite struct {
	ctrl  *gomock.Controller
	cache *mock_cache.MockCache

	userRepo         *mock_repository.MockUserRepository
	refreshTokenRepo *mock_repository.MockRefreshTokenRepository
	passwordHasher   *mock_password_hasher.MockPasswordHasher
	jwtService       *mock_jwt.MockJwtProvider

	tokenCache *caching.TokenCache

	authService service.AuthService
}

func setupAuthServiceTestSuite(t *testing.T) *AuthServiceTestSuite {
	ctrl := gomock.NewController(t)
	cache := mock_cache.NewMockCache(ctrl)

	userRepo := mock_repository.NewMockUserRepository(ctrl)
	refreshTokenRepo := mock_repository.NewMockRefreshTokenRepository(ctrl)
	passwordHasher := mock_password_hasher.NewMockPasswordHasher(ctrl)
	jwtService := mock_jwt.NewMockJwtProvider(ctrl)

	tokenCache := caching.NewTokenCache(cache)

	authService := service.NewAuthService(userRepo, refreshTokenRepo, tokenCache, passwordHasher, jwtService)
	return &AuthServiceTestSuite{
		ctrl:             ctrl,
		cache:            cache,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		passwordHasher:   passwordHasher,
		jwtService:       jwtService,
		tokenCache:       tokenCache,
		authService:      authService,
	}
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name          string
		req           *request.RegisterRequest
		setupMock     func(suite *AuthServiceTestSuite)
		expectedError error
		checkFunc     func(t *testing.T, user *models.User, err error)
	}{
		{
			name: "Register Success",
			req: &request.RegisterRequest{
				Email:    "test@gmail.com",
				Password: "passwrod123",
				FullName: "testuser",
			},
			setupMock: func(s *AuthServiceTestSuite) {
				s.userRepo.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})
				s.userRepo.EXPECT().GetByEmail(gomock.Any(), "test@gmail.com").Return(nil, nil)
				s.passwordHasher.EXPECT().Hash("passwrod123").Return("hashedpassword", nil)
				s.userRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *models.User) error {
						user.ID = uuid.New()
						return nil
					})
				s.cache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedError: nil,
			checkFunc: func(t *testing.T, user *models.User, err error) {
				assert.NotNil(t, user)
				assert.Equal(t, "test@gmail.com", user.Email)
				assert.Equal(t, "testuser", user.FullName)
				assert.NotEmpty(t, user.ID)
			},
		},
		{
			name: "User Already Exists",
			req: &request.RegisterRequest{
				Email:    "existing@gmail.com",
				Password: "passwrod123",
				FullName: "testuser",
			},
			setupMock: func(s *AuthServiceTestSuite) {
				existingUser := &models.User{
					ID:    uuid.New(),
					Email: "existing@gmail.com",
				}
				s.userRepo.EXPECT().
					GetByEmail(gomock.Any(), "existing@gmail.com").
					Return(existingUser, nil)
			},
			expectedError: apperr.ErrEmailAlreadyExists,
			checkFunc: func(t *testing.T, user *models.User, err error) {
				assert.Nil(t, user)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := setupAuthServiceTestSuite(t)
			defer suite.ctrl.Finish()

			ctx := context.WithValue(context.Background(), contextkeys.LoggerKey, zap.NewNop())

			tt.setupMock(suite)

			user, err := suite.authService.Register(ctx, tt.req)

			assert.True(t, errors.Is(err, tt.expectedError))

			if tt.checkFunc != nil {
				tt.checkFunc(t, user, err)
			}
		})
	}
}

func TestAuthService_VerifyEmail(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		setupMock     func(suite *AuthServiceTestSuite)
		expectedError error
	}{
		{
			name:  "Verify Email Success",
			token: "valid-token",
			setupMock: func(s *AuthServiceTestSuite) {
				s.cache.EXPECT().Get(gomock.Any(), gomock.Any()).Return("test@gmail.com", nil)
				s.cache.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
				user := &models.User{
					ID:     uuid.New(),
					Email:  "test@gmail.com",
					Status: models.UserStatusPending,
				}
				s.userRepo.EXPECT().GetByEmail(gomock.Any(), "test@gmail.com").Return(user, nil)
				s.userRepo.EXPECT().VerifyEmail(gomock.Any(), user.ID).Return(int64(1), nil)
				s.userRepo.EXPECT().UpdateStatus(gomock.Any(), user.ID, models.UserStatusActive).Return(int64(1), nil)
			},
			expectedError: nil,
		},
		{
			name:  "Token Invalid Or Expired",
			token: "invalid-token",
			setupMock: func(s *AuthServiceTestSuite) {
				s.cache.EXPECT().Get(gomock.Any(), gomock.Any()).Return("", errors.New("not found"))
			},
			expectedError: apperr.ErrTokenInvalidOrExpired,
		},
		{
			name:  "User Not Found",
			token: "valid-token",
			setupMock: func(s *AuthServiceTestSuite) {
				s.cache.EXPECT().Get(gomock.Any(), gomock.Any()).Return("notfound@gmail.com", nil)
				s.cache.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
				s.userRepo.EXPECT().GetByEmail(gomock.Any(), "notfound@gmail.com").Return(nil, nil)
			},
			expectedError: apperr.ErrUserNotFound,
		},
		{
			name:  "Email Already Verified",
			token: "valid-token",
			setupMock: func(s *AuthServiceTestSuite) {
				s.cache.EXPECT().Get(gomock.Any(), gomock.Any()).Return("verified@gmail.com", nil)
				s.cache.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
				verifiedAt := time.Now()
				user := &models.User{
					ID:              uuid.New(),
					Email:           "verified@gmail.com",
					Status:          models.UserStatusActive,
					EmailVerifiedAt: &verifiedAt,
				}
				s.userRepo.EXPECT().GetByEmail(gomock.Any(), "verified@gmail.com").Return(user, nil)
			},
			expectedError: apperr.ErrEmailAlreadyVerified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := setupAuthServiceTestSuite(t)
			defer suite.ctrl.Finish()

			ctx := context.WithValue(context.Background(), contextkeys.LoggerKey, zap.NewNop())
			tt.setupMock(suite)

			err := suite.authService.VerifyEmail(ctx, tt.token)

			assert.True(t, errors.Is(err, tt.expectedError))
		})
	}
}

func TestAuthService_ResendVerificationEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		setupMock     func(suite *AuthServiceTestSuite)
		expectedError error
	}{
		{
			name:  "Resend Success",
			email: "pending@gmail.com",
			setupMock: func(s *AuthServiceTestSuite) {
				user := &models.User{
					ID:     uuid.New(),
					Email:  "pending@gmail.com",
					Status: models.UserStatusPending,
				}
				s.userRepo.EXPECT().GetByEmail(gomock.Any(), "pending@gmail.com").Return(user, nil)
				s.cache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "User Not Found",
			email: "notfound@gmail.com",
			setupMock: func(s *AuthServiceTestSuite) {
				s.userRepo.EXPECT().GetByEmail(gomock.Any(), "notfound@gmail.com").Return(nil, nil)
			},
			expectedError: apperr.ErrUserNotFound,
		},
		{
			name:  "Email Already Verified",
			email: "verified@gmail.com",
			setupMock: func(s *AuthServiceTestSuite) {
				verifiedAt := time.Now()
				user := &models.User{
					ID:              uuid.New(),
					Email:           "verified@gmail.com",
					Status:          models.UserStatusActive,
					EmailVerifiedAt: &verifiedAt,
				}
				s.userRepo.EXPECT().GetByEmail(gomock.Any(), "verified@gmail.com").Return(user, nil)
			},
			expectedError: apperr.ErrEmailAlreadyVerified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := setupAuthServiceTestSuite(t)
			defer suite.ctrl.Finish()

			ctx := context.WithValue(context.Background(), contextkeys.LoggerKey, zap.NewNop())
			tt.setupMock(suite)

			err := suite.authService.ResendVerificationEmail(ctx, tt.email)

			assert.True(t, errors.Is(err, tt.expectedError))
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	testUserID := uuid.New()
	verifiedAt := time.Now()

	tests := []struct {
		name          string
		req           *request.LoginRequest
		setupMock     func(suite *AuthServiceTestSuite)
		expectedError error
		checkFunc     func(t *testing.T, user *models.User, accessToken, refreshToken string, err error)
	}{
		{
			name: "Login Success",
			req: &request.LoginRequest{
				Email:    "active@gmail.com",
				Password: "password123",
			},
			setupMock: func(s *AuthServiceTestSuite) {
				user := &models.User{
					ID:              testUserID,
					Email:           "active@gmail.com",
					HashedPassword:  "hashedpassword",
					Status:          models.UserStatusActive,
					EmailVerifiedAt: &verifiedAt,
				}
				s.userRepo.EXPECT().GetByEmail(gomock.Any(), "active@gmail.com").Return(user, nil)
				s.passwordHasher.EXPECT().Compare("hashedpassword", "password123").Return(true)
				s.jwtService.EXPECT().GenerateAccessToken(gomock.Any()).Return("access-token", nil)
				s.jwtService.EXPECT().GenerateRefreshToken(testUserID.String()).Return("refresh-token", nil)
				s.jwtService.EXPECT().GetRefreshTTL().Return(7 * 24 * time.Hour)
				s.refreshTokenRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedError: nil,
			checkFunc: func(t *testing.T, user *models.User, accessToken, refreshToken string, err error) {
				assert.NotNil(t, user)
				assert.Equal(t, "access-token", accessToken)
				assert.Equal(t, "refresh-token", refreshToken)
			},
		},
		{
			name: "User Not Found",
			req: &request.LoginRequest{
				Email:    "notfound@gmail.com",
				Password: "password123",
			},
			setupMock: func(s *AuthServiceTestSuite) {
				s.userRepo.EXPECT().GetByEmail(gomock.Any(), "notfound@gmail.com").Return(nil, nil)
			},
			expectedError: apperr.ErrInvalidCredentials,
			checkFunc:     nil,
		},
		{
			name: "Wrong Password",
			req: &request.LoginRequest{
				Email:    "active@gmail.com",
				Password: "wrongpassword",
			},
			setupMock: func(s *AuthServiceTestSuite) {
				user := &models.User{
					ID:             testUserID,
					Email:          "active@gmail.com",
					HashedPassword: "hashedpassword",
					Status:         models.UserStatusActive,
				}
				s.userRepo.EXPECT().GetByEmail(gomock.Any(), "active@gmail.com").Return(user, nil)
				s.passwordHasher.EXPECT().Compare("hashedpassword", "wrongpassword").Return(false)
			},
			expectedError: apperr.ErrInvalidCredentials,
			checkFunc:     nil,
		},
		{
			name: "Account Inactive",
			req: &request.LoginRequest{
				Email:    "inactive@gmail.com",
				Password: "password123",
			},
			setupMock: func(s *AuthServiceTestSuite) {
				user := &models.User{
					ID:             testUserID,
					Email:          "inactive@gmail.com",
					HashedPassword: "hashedpassword",
					Status:         models.UserStatusPending,
				}
				s.userRepo.EXPECT().GetByEmail(gomock.Any(), "inactive@gmail.com").Return(user, nil)
				s.passwordHasher.EXPECT().Compare("hashedpassword", "password123").Return(true)
			},
			expectedError: apperr.ErrAccountInactive,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := setupAuthServiceTestSuite(t)
			defer suite.ctrl.Finish()

			ctx := context.WithValue(context.Background(), contextkeys.LoggerKey, zap.NewNop())
			tt.setupMock(suite)

			user, accessToken, refreshToken, err := suite.authService.Login(ctx, tt.req)

			assert.True(t, errors.Is(err, tt.expectedError))

			if tt.checkFunc != nil {
				tt.checkFunc(t, user, accessToken, refreshToken, err)
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	testUserID := uuid.New()
	testTokenID := uuid.New()
	verifiedAt := time.Now()

	tests := []struct {
		name          string
		refreshToken  string
		setupMock     func(suite *AuthServiceTestSuite)
		expectedError error
		checkFunc     func(t *testing.T, accessToken, refreshToken string, err error)
	}{
		{
			name:         "Refresh Token Success",
			refreshToken: "valid-refresh-token",
			setupMock: func(s *AuthServiceTestSuite) {
				claims := &jwtprovider.RefreshTokenClaims{}
				claims.Subject = testUserID.String()

				s.jwtService.EXPECT().VerifyRefreshToken("valid-refresh-token").Return(claims, nil)
				s.refreshTokenRepo.EXPECT().GetByToken(gomock.Any(), gomock.Any()).Return(&models.RefreshToken{
					ID:        testTokenID,
					UserID:    testUserID,
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}, nil)
				s.userRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(&models.User{
					ID:              testUserID,
					Email:           "test@gmail.com",
					Status:          models.UserStatusActive,
					EmailVerifiedAt: &verifiedAt,
				}, nil)
				s.refreshTokenRepo.EXPECT().DeleteByID(gomock.Any(), testTokenID).Return(int64(1), nil)
				s.jwtService.EXPECT().GenerateAccessToken(gomock.Any()).Return("new-access-token", nil)
				s.jwtService.EXPECT().GenerateRefreshToken(testUserID.String()).Return("new-refresh-token", nil)
				s.jwtService.EXPECT().GetRefreshTTL().Return(7 * 24 * time.Hour)
				s.refreshTokenRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedError: nil,
			checkFunc: func(t *testing.T, accessToken, refreshToken string, err error) {
				assert.Equal(t, "new-access-token", accessToken)
				assert.Equal(t, "new-refresh-token", refreshToken)
			},
		},
		{
			name:         "Token Expired",
			refreshToken: "expired-token",
			setupMock: func(s *AuthServiceTestSuite) {
				s.jwtService.EXPECT().VerifyRefreshToken("expired-token").Return(nil, jwtprovider.ErrTokenExpired)
			},
			expectedError: apperr.ErrTokenExpired,
			checkFunc:     nil,
		},
		{
			name:         "Token Invalid",
			refreshToken: "invalid-token",
			setupMock: func(s *AuthServiceTestSuite) {
				s.jwtService.EXPECT().VerifyRefreshToken("invalid-token").Return(nil, jwtprovider.ErrTokenInvalid)
			},
			expectedError: apperr.ErrTokenInvalid,
			checkFunc:     nil,
		},
		{
			name:         "Token Not Found In DB",
			refreshToken: "valid-but-not-in-db",
			setupMock: func(s *AuthServiceTestSuite) {
				claims := &jwtprovider.RefreshTokenClaims{}
				claims.Subject = testUserID.String()

				s.jwtService.EXPECT().VerifyRefreshToken("valid-but-not-in-db").Return(claims, nil)
				s.refreshTokenRepo.EXPECT().GetByToken(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			expectedError: apperr.ErrTokenInvalid,
			checkFunc:     nil,
		},
		{
			name:         "User Not Found",
			refreshToken: "valid-refresh-token",
			setupMock: func(s *AuthServiceTestSuite) {
				claims := &jwtprovider.RefreshTokenClaims{}
				claims.Subject = testUserID.String()

				s.jwtService.EXPECT().VerifyRefreshToken("valid-refresh-token").Return(claims, nil)
				s.refreshTokenRepo.EXPECT().GetByToken(gomock.Any(), gomock.Any()).Return(&models.RefreshToken{
					ID:        testTokenID,
					UserID:    testUserID,
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}, nil)
				s.userRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(nil, nil)
			},
			expectedError: apperr.ErrTokenInvalid,
			checkFunc:     nil,
		},
		{
			name:         "Account Inactive",
			refreshToken: "valid-refresh-token",
			setupMock: func(s *AuthServiceTestSuite) {
				claims := &jwtprovider.RefreshTokenClaims{}
				claims.Subject = testUserID.String()

				s.jwtService.EXPECT().VerifyRefreshToken("valid-refresh-token").Return(claims, nil)
				s.refreshTokenRepo.EXPECT().GetByToken(gomock.Any(), gomock.Any()).Return(&models.RefreshToken{
					ID:        testTokenID,
					UserID:    testUserID,
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}, nil)
				s.userRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(&models.User{
					ID:     testUserID,
					Email:  "inactive@gmail.com",
					Status: models.UserStatusPending,
				}, nil)
			},
			expectedError: apperr.ErrAccountInactive,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := setupAuthServiceTestSuite(t)
			defer suite.ctrl.Finish()

			ctx := context.WithValue(context.Background(), contextkeys.LoggerKey, zap.NewNop())
			tt.setupMock(suite)

			accessToken, refreshToken, err := suite.authService.RefreshToken(ctx, tt.refreshToken)

			assert.True(t, errors.Is(err, tt.expectedError))

			if tt.checkFunc != nil {
				tt.checkFunc(t, accessToken, refreshToken, err)
			}
		})
	}
}

func TestAuthService_ChangePassword(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		name          string
		userID        string
		req           *request.ChangePasswordRequest
		setupMock     func(suite *AuthServiceTestSuite)
		expectedError error
	}{
		{
			name:   "Change Password Success",
			userID: testUserID.String(),
			req: &request.ChangePasswordRequest{
				CurrentPassword: "oldpassword",
				NewPassword:     "newpassword",
			},
			setupMock: func(s *AuthServiceTestSuite) {
				user := &models.User{
					ID:             testUserID,
					Email:          "test@gmail.com",
					HashedPassword: "hashedoldpassword",
					Status:         models.UserStatusActive,
				}
				s.userRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(user, nil)
				s.passwordHasher.EXPECT().Compare("hashedoldpassword", "oldpassword").Return(true)
				s.passwordHasher.EXPECT().Hash("newpassword").Return("hashednewpassword", nil)
				s.userRepo.EXPECT().UpdatePassword(gomock.Any(), testUserID, "hashednewpassword").Return(int64(1), nil)
			},
			expectedError: nil,
		},
		{
			name:   "User Not Found",
			userID: testUserID.String(),
			req: &request.ChangePasswordRequest{
				CurrentPassword: "oldpassword",
				NewPassword:     "newpassword",
			},
			setupMock: func(s *AuthServiceTestSuite) {
				s.userRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(nil, nil)
			},
			expectedError: apperr.ErrUserNotFound,
		},
		{
			name:   "Invalid Current Password",
			userID: testUserID.String(),
			req: &request.ChangePasswordRequest{
				CurrentPassword: "wrongpassword",
				NewPassword:     "newpassword",
			},
			setupMock: func(s *AuthServiceTestSuite) {
				user := &models.User{
					ID:             testUserID,
					Email:          "test@gmail.com",
					HashedPassword: "hashedoldpassword",
					Status:         models.UserStatusActive,
				}
				s.userRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(user, nil)
				s.passwordHasher.EXPECT().Compare("hashedoldpassword", "wrongpassword").Return(false)
			},
			expectedError: apperr.ErrInvalidCurrentPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := setupAuthServiceTestSuite(t)
			defer suite.ctrl.Finish()

			ctx := context.WithValue(context.Background(), contextkeys.LoggerKey, zap.NewNop())
			tt.setupMock(suite)

			err := suite.authService.ChangePassword(ctx, tt.userID, tt.req)

			assert.True(t, errors.Is(err, tt.expectedError))
		})
	}
}

func TestAuthService_ForgotPassword(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		name          string
		email         string
		setupMock     func(suite *AuthServiceTestSuite)
		expectedError error
	}{
		{
			name:  "Forgot Password Success",
			email: "test@gmail.com",
			setupMock: func(s *AuthServiceTestSuite) {
				user := &models.User{
					ID:     testUserID,
					Email:  "test@gmail.com",
					Status: models.UserStatusActive,
				}
				s.userRepo.EXPECT().GetByEmail(gomock.Any(), "test@gmail.com").Return(user, nil)
				s.cache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "User Not Found",
			email: "notfound@gmail.com",
			setupMock: func(s *AuthServiceTestSuite) {
				s.userRepo.EXPECT().GetByEmail(gomock.Any(), "notfound@gmail.com").Return(nil, nil)
			},
			expectedError: apperr.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := setupAuthServiceTestSuite(t)
			defer suite.ctrl.Finish()

			ctx := context.WithValue(context.Background(), contextkeys.LoggerKey, zap.NewNop())
			tt.setupMock(suite)

			err := suite.authService.ForgotPassword(ctx, tt.email)

			assert.True(t, errors.Is(err, tt.expectedError))
		})
	}
}

func TestAuthService_ResetPassword(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		name          string
		token         string
		newPassword   string
		setupMock     func(suite *AuthServiceTestSuite)
		expectedError error
	}{
		{
			name:        "Reset Password Success",
			token:       "valid-reset-token",
			newPassword: "newpassword123",
			setupMock: func(s *AuthServiceTestSuite) {
				s.cache.EXPECT().Get(gomock.Any(), gomock.Any()).Return("test@gmail.com", nil)
				s.cache.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
				user := &models.User{
					ID:     testUserID,
					Email:  "test@gmail.com",
					Status: models.UserStatusActive,
				}
				s.userRepo.EXPECT().GetByEmail(gomock.Any(), "test@gmail.com").Return(user, nil)
				s.passwordHasher.EXPECT().Hash("newpassword123").Return("hashednewpassword", nil)
				s.userRepo.EXPECT().UpdatePassword(gomock.Any(), testUserID, "hashednewpassword").Return(int64(1), nil)
			},
			expectedError: nil,
		},
		{
			name:        "Token Invalid Or Expired",
			token:       "invalid-token",
			newPassword: "newpassword123",
			setupMock: func(s *AuthServiceTestSuite) {
				s.cache.EXPECT().Get(gomock.Any(), gomock.Any()).Return("", errors.New("not found"))
			},
			expectedError: apperr.ErrTokenInvalidOrExpired,
		},
		{
			name:        "User Not Found",
			token:       "valid-token",
			newPassword: "newpassword123",
			setupMock: func(s *AuthServiceTestSuite) {
				s.cache.EXPECT().Get(gomock.Any(), gomock.Any()).Return("notfound@gmail.com", nil)
				s.cache.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
				s.userRepo.EXPECT().GetByEmail(gomock.Any(), "notfound@gmail.com").Return(nil, nil)
			},
			expectedError: apperr.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := setupAuthServiceTestSuite(t)
			defer suite.ctrl.Finish()

			ctx := context.WithValue(context.Background(), contextkeys.LoggerKey, zap.NewNop())
			tt.setupMock(suite)

			err := suite.authService.ResetPassword(ctx, tt.token, tt.newPassword)

			assert.True(t, errors.Is(err, tt.expectedError))
		})
	}
}
