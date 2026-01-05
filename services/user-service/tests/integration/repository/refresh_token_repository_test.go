package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository/impl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T, ctx context.Context) *models.User {
	userRepo := impl.NewUserRepository(testDB.Pool)
	user := &models.User{
		Email:          "user" + uuid.New().String()[:8] + "@gmail.com",
		HashedPassword: "hashedpassword123",
		FullName:       "Test User",
		Status:         models.UserStatusActive,
	}
	require.NoError(t, userRepo.Create(ctx, user))
	return user
}

func TestRefreshTokenRepository_Save(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewRefreshTokenRepository(testDB.Pool)

	var tokenToSave *models.RefreshToken

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		expectedError bool
		checkFunc     func(t *testing.T)
	}{
		{
			name: "Save refresh token success",
			setup: func(t *testing.T) {
				user := createTestUser(t, ctx)
				tokenToSave = &models.RefreshToken{
					UserID:    user.ID,
					TokenHash: "hashedtoken123",
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}
			},
			expectedError: false,
			checkFunc: func(t *testing.T) {
				found, err := repo.FindByToken(ctx, tokenToSave.TokenHash)
				require.NoError(t, err)
				assert.NotNil(t, found)
				assert.Equal(t, tokenToSave.TokenHash, found.TokenHash)
				assert.Equal(t, tokenToSave.UserID, found.UserID)
			},
		},
		{
			name: "Save refresh token with invalid user ID",
			setup: func(t *testing.T) {
				tokenToSave = &models.RefreshToken{
					UserID:    uuid.New(),
					TokenHash: "hashedtoken456",
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}
			},
			expectedError: true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testDB.CleanupTestData(ctx))

			if tt.setup != nil {
				tt.setup(t)
			}

			err := repo.Save(ctx, tokenToSave)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t)
			}
		})
	}
}

func TestRefreshTokenRepository_FindByToken(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewRefreshTokenRepository(testDB.Pool)

	var tokenHashToFind string

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		expectedError bool
		checkFunc     func(t *testing.T, token *models.RefreshToken)
	}{
		{
			name: "Find existing token",
			setup: func(t *testing.T) {
				user := createTestUser(t, ctx)
				tokenHashToFind = "existingtoken123"
				token := &models.RefreshToken{
					UserID:    user.ID,
					TokenHash: tokenHashToFind,
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}
				require.NoError(t, repo.Save(ctx, token))
			},
			expectedError: false,
			checkFunc: func(t *testing.T, token *models.RefreshToken) {
				assert.NotNil(t, token)
				assert.Equal(t, tokenHashToFind, token.TokenHash)
			},
		},
		{
			name: "Find non-existing token",
			setup: func(t *testing.T) {
				tokenHashToFind = "nonexistenttoken"
			},
			expectedError: false,
			checkFunc: func(t *testing.T, token *models.RefreshToken) {
				assert.Nil(t, token)
			},
		},
		{
			name: "Find with empty token hash",
			setup: func(t *testing.T) {
				tokenHashToFind = ""
			},
			expectedError: false,
			checkFunc: func(t *testing.T, token *models.RefreshToken) {
				assert.Nil(t, token)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testDB.CleanupTestData(ctx))

			if tt.setup != nil {
				tt.setup(t)
			}

			result, err := repo.FindByToken(ctx, tokenHashToFind)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestRefreshTokenRepository_DeleteByID(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewRefreshTokenRepository(testDB.Pool)

	var tokenIDToDelete uuid.UUID
	var tokenHash string

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		expectedError bool
		checkFunc     func(t *testing.T)
	}{
		{
			name: "Delete existing token",
			setup: func(t *testing.T) {
				user := createTestUser(t, ctx)
				tokenHash = "deletetokentest"
				token := &models.RefreshToken{
					UserID:    user.ID,
					TokenHash: tokenHash,
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}
				require.NoError(t, repo.Save(ctx, token))

				savedToken, err := repo.FindByToken(ctx, tokenHash)
				require.NoError(t, err)
				require.NotNil(t, savedToken)
				tokenIDToDelete = savedToken.ID
			},
			expectedError: false,
			checkFunc: func(t *testing.T) {
				deletedToken, err := repo.FindByToken(ctx, tokenHash)
				require.NoError(t, err)
				assert.Nil(t, deletedToken, "Token should be deleted")
			},
		},
		{
			name: "Delete non-existing token",
			setup: func(t *testing.T) {
				tokenIDToDelete = uuid.New()
			},
			expectedError: true,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testDB.CleanupTestData(ctx))

			if tt.setup != nil {
				tt.setup(t)
			}

			err := repo.DeleteByID(ctx, tokenIDToDelete)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t)
			}
		})
	}
}
