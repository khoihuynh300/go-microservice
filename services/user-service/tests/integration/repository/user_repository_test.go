package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository/impl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewUserRepository(testDB.Pool)

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		user          *models.User
		expectedError bool
		checkFunc     func(t *testing.T, user *models.User)
	}{
		{
			name:  "Create User Success",
			setup: nil,
			user: &models.User{
				Email:          "test@gmail.com",
				HashedPassword: "hashedpassword123",
				FullName:       "Test User",
				Status:         models.UserStatusPending,
			},
			expectedError: false,
			checkFunc: func(t *testing.T, user *models.User) {
				assert.NotEqual(t, uuid.Nil, user.ID)
				// assert.False(t, user.CreatedAt.IsZero())
				// assert.False(t, user.UpdatedAt.IsZero())
			},
		},
		{
			name: "Create User Duplicate Email",
			setup: func(t *testing.T) {
				existingUser := &models.User{
					Email:          "duplicate@gmail.com",
					HashedPassword: "hashedpassword456",
					FullName:       "User 2",
					Status:         models.UserStatusPending,
				}
				require.NoError(t, repo.Create(ctx, existingUser))
			},
			user: &models.User{
				Email:          "duplicate@gmail.com",
				HashedPassword: "hashedpassword123",
				FullName:       "User 1",
				Status:         models.UserStatusPending,
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

			err := repo.Create(ctx, tt.user)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, tt.user)
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewUserRepository(testDB.Pool)

	var findUserId uuid.UUID

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		expectedError bool
		checkFunc     func(t *testing.T, user *models.User)
	}{
		{
			name: "Get existing user by ID",
			setup: func(t *testing.T) {
				user := &models.User{
					Email:          "test@gmail.com",
					HashedPassword: "hashedpassword123",
					FullName:       "Test User",
					Status:         models.UserStatusActive,
				}
				require.NoError(t, repo.Create(ctx, user))
				findUserId = user.ID
			},
			expectedError: false,
			checkFunc: func(t *testing.T, user *models.User) {
				assert.NotNil(t, user)
				assert.Equal(t, findUserId, user.ID)
			},
		},
		{
			name: "Get non-existing user by ID",
			setup: func(t *testing.T) {
				findUserId = uuid.New()
			},
			expectedError: false,
			checkFunc: func(t *testing.T, user *models.User) {
				assert.Nil(t, user)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testDB.CleanupTestData(ctx))

			if tt.setup != nil {
				tt.setup(t)
			}

			result, err := repo.GetByID(ctx, findUserId)

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

func TestUserRepository_GetByEmail(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewUserRepository(testDB.Pool)

	var findEmail string

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		expectedError bool
		checkFunc     func(t *testing.T, user *models.User)
	}{
		{
			name: "Get existing user by email",
			setup: func(t *testing.T) {
				user := &models.User{
					Email:          "test@gmail.com",
					HashedPassword: "hashedpassword123",
					FullName:       "Test User",
					Status:         models.UserStatusActive,
				}
				require.NoError(t, repo.Create(ctx, user))
				findEmail = user.Email
			},
			expectedError: false,
			checkFunc: func(t *testing.T, user *models.User) {
				assert.NotNil(t, user)
				assert.Equal(t, findEmail, user.Email)
			},
		},
		{
			name: "Get non-existing user by email",
			setup: func(t *testing.T) {
				findEmail = "nonexistent@gmail.com"
			},
			expectedError: false,
			checkFunc: func(t *testing.T, user *models.User) {
				assert.Nil(t, user)
			},
		},
		{
			name: "Get with empty email",
			setup: func(t *testing.T) {
				findEmail = ""
			},
			expectedError: false,
			checkFunc: func(t *testing.T, user *models.User) {
				assert.Nil(t, user)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testDB.CleanupTestData(ctx))

			if tt.setup != nil {
				tt.setup(t)
			}

			result, err := repo.GetByEmail(ctx, findEmail)

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

func TestUserRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewUserRepository(testDB.Pool)

	var userToUpdate *models.User

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		updateFunc    func(u *models.User)
		expectedError bool
		rowsAffected  int64
		checkFunc     func(t *testing.T, user *models.User)
	}{
		{
			name: "Update full name",
			setup: func(t *testing.T) {
				userToUpdate = &models.User{
					Email:          "test@gmail.com",
					HashedPassword: "hashedpassword123",
					FullName:       "Original Name",
					Status:         models.UserStatusActive,
				}
				require.NoError(t, repo.Create(ctx, userToUpdate))
			},
			updateFunc: func(u *models.User) {
				u.FullName = "Updated Name"
			},
			expectedError: false,
			rowsAffected:  1,
			checkFunc: func(t *testing.T, user *models.User) {
				require.NotNil(t, user)
				assert.Equal(t, "Updated Name", user.FullName)
			},
		},
		{
			name: "Update phone",
			setup: func(t *testing.T) {
				userToUpdate = &models.User{
					Email:          "test@gmail.com",
					HashedPassword: "hashedpassword123",
					FullName:       "Test User",
					Status:         models.UserStatusActive,
				}
				assert.NoError(t, repo.Create(ctx, userToUpdate))
			},
			updateFunc: func(u *models.User) {
				phone := "0123456789"
				u.Phone = &phone
			},
			expectedError: false,
			rowsAffected:  1,
			checkFunc: func(t *testing.T, user *models.User) {
				require.NotNil(t, user)
				require.NotNil(t, user.Phone)
				assert.Equal(t, "0123456789", *user.Phone)
			},
		},
		{
			name: "Update gender",
			setup: func(t *testing.T) {
				userToUpdate = &models.User{
					Email:          "test@gmail.com",
					HashedPassword: "hashedpassword123",
					FullName:       "Test User",
					Status:         models.UserStatusActive,
				}
				require.NoError(t, repo.Create(ctx, userToUpdate))
			},
			updateFunc: func(u *models.User) {
				gender := models.GenderMale
				u.Gender = &gender
			},
			expectedError: false,
			rowsAffected:  1,
			checkFunc: func(t *testing.T, user *models.User) {
				require.NotNil(t, user)
				require.NotNil(t, user.Gender)
				assert.Equal(t, models.GenderMale, *user.Gender)
			},
		},
		{
			name: "Update status to active",
			setup: func(t *testing.T) {
				userToUpdate = &models.User{
					Email:          "test@gmail.com",
					HashedPassword: "hashedpassword123",
					FullName:       "Test User",
					Status:         models.UserStatusActive,
				}
				require.NoError(t, repo.Create(ctx, userToUpdate))
			},
			updateFunc: func(u *models.User) {
				u.Status = models.UserStatusActive
			},
			expectedError: false,
			rowsAffected:  1,
			checkFunc: func(t *testing.T, user *models.User) {
				require.NotNil(t, user)
				assert.Equal(t, models.UserStatusActive, user.Status)
			},
		},
		{
			name: "Update non-existing user",
			setup: func(t *testing.T) {
				userToUpdate = &models.User{
					ID:       uuid.New(),
					Email:    "test@gmail.com",
					FullName: "Test User",
					Status:   models.UserStatusActive,
				}
			},
			updateFunc:    nil,
			expectedError: false,
			rowsAffected:  0,
			checkFunc: func(t *testing.T, user *models.User) {
				require.Nil(t, user)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testDB.CleanupTestData(ctx))

			if tt.setup != nil {
				tt.setup(t)
			}

			if tt.updateFunc != nil {
				tt.updateFunc(userToUpdate)
			}

			rowsAffected, err := repo.Update(ctx, userToUpdate)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.rowsAffected, rowsAffected)

			updatedUser, err := repo.GetByID(ctx, userToUpdate.ID)
			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, updatedUser)
			}
		})
	}
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewUserRepository(testDB.Pool)

	var userId uuid.UUID
	var newPassword string

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		expectedError bool
		rowsAffected  int64
		checkFunc     func(t *testing.T)
	}{
		{
			name: "Update password success",
			setup: func(t *testing.T) {
				user := &models.User{
					Email:          "test@gmail.com",
					HashedPassword: "oldhashedpassword",
					FullName:       "Test User",
					Status:         models.UserStatusActive,
				}
				require.NoError(t, repo.Create(ctx, user))
				userId = user.ID
				newPassword = "newhashedpassword123"
			},
			expectedError: false,
			rowsAffected:  1,
			checkFunc: func(t *testing.T) {
				updatedUser, err := repo.GetByID(ctx, userId)
				require.NoError(t, err)
				require.NotNil(t, updatedUser)
				assert.Equal(t, newPassword, updatedUser.HashedPassword)
			},
		},
		{
			name: "Update password for non-existing user",
			setup: func(t *testing.T) {
				userId = uuid.New()
				newPassword = "somepassword"
			},
			expectedError: false,
			rowsAffected:  0,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testDB.CleanupTestData(ctx))

			if tt.setup != nil {
				tt.setup(t)
			}

			rowsAffected, err := repo.UpdatePassword(ctx, userId, newPassword)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, tt.rowsAffected, rowsAffected)

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t)
			}
		})
	}
}

func TestUserRepository_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewUserRepository(testDB.Pool)

	var userId uuid.UUID
	var newStatus models.UserStatus

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		expectedError bool
		rowsAffected  int64
		checkFunc     func(t *testing.T)
	}{
		{
			name: "Update status to active",
			setup: func(t *testing.T) {
				user := &models.User{
					Email:          "test@gmail.com",
					HashedPassword: "hashedpassword123",
					FullName:       "Test User",
					Status:         models.UserStatusPending,
				}
				require.NoError(t, repo.Create(ctx, user))
				userId = user.ID
				newStatus = models.UserStatusActive
			},
			expectedError: false,
			rowsAffected:  1,
			checkFunc: func(t *testing.T) {
				updatedUser, err := repo.GetByID(ctx, userId)
				require.NoError(t, err)
				require.NotNil(t, updatedUser)
				assert.Equal(t, models.UserStatusActive, updatedUser.Status)
			},
		},
		{
			name: "Update status to inactive",
			setup: func(t *testing.T) {
				user := &models.User{
					Email:          "test@gmail.com",
					HashedPassword: "hashedpassword123",
					FullName:       "Test User",
					Status:         models.UserStatusActive,
				}
				require.NoError(t, repo.Create(ctx, user))
				userId = user.ID
				newStatus = models.UserStatusInactive
			},
			expectedError: false,
			rowsAffected:  1,
			checkFunc: func(t *testing.T) {
				updatedUser, err := repo.GetByID(ctx, userId)
				require.NoError(t, err)
				require.NotNil(t, updatedUser)
				assert.Equal(t, models.UserStatusInactive, updatedUser.Status)
			},
		},
		{
			name: "Update status for non-existing user",
			setup: func(t *testing.T) {
				userId = uuid.New()
				newStatus = models.UserStatusActive
			},
			expectedError: false,
			rowsAffected:  0,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testDB.CleanupTestData(ctx))

			if tt.setup != nil {
				tt.setup(t)
			}

			rowsAffected, err := repo.UpdateStatus(ctx, userId, newStatus)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, tt.rowsAffected, rowsAffected)

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t)
			}
		})
	}
}

func TestUserRepository_SoftDelete(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewUserRepository(testDB.Pool)

	var userId uuid.UUID

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		expectedError bool
		rowsAffected  int64
		checkFunc     func(t *testing.T)
	}{
		{
			name: "Soft delete existing user",
			setup: func(t *testing.T) {
				user := &models.User{
					Email:          "test@gmail.com",
					HashedPassword: "hashedpassword123",
					FullName:       "Test User",
					Status:         models.UserStatusActive,
				}
				require.NoError(t, repo.Create(ctx, user))
				userId = user.ID
			},
			expectedError: false,
			rowsAffected:  1,
			checkFunc: func(t *testing.T) {
				deletedUser, err := repo.GetByID(ctx, userId)
				require.NoError(t, err)
				assert.Nil(t, deletedUser, "Soft deleted user should not be found")
			},
		},
		{
			name: "Soft delete non-existing user",
			setup: func(t *testing.T) {
				userId = uuid.New()
			},
			expectedError: false,
			rowsAffected:  0,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testDB.CleanupTestData(ctx))

			if tt.setup != nil {
				tt.setup(t)
			}

			rowsAffected, err := repo.SoftDelete(ctx, userId)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, tt.rowsAffected, rowsAffected)

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t)
			}
		})
	}
}
