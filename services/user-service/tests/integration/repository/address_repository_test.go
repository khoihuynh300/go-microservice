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

func createTestUserForAddress(t *testing.T, ctx context.Context) *models.User {
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

func createTestAddress(userID uuid.UUID) *models.Address {
	return &models.Address{
		UserID:       userID,
		AddressType:  models.AddressTypeHome,
		FullName:     "Test User",
		Phone:        "0123456789",
		AddressLine1: "123 Main Street",
		AddressLine2: "Room 1",
		Ward:         "Ward 1",
		City:         "Ho Chi Minh",
		Country:      "Vietnam",
		IsDefault:    false,
	}
}

func TestAddressRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewAddressRepository(testDB.Pool)

	var addressToCreate *models.Address

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		expectedError bool
		checkFunc     func(t *testing.T, address *models.Address)
	}{
		{
			name: "Create address success",
			setup: func(t *testing.T) {
				user := createTestUserForAddress(t, ctx)
				addressToCreate = createTestAddress(user.ID)
			},
			expectedError: false,
			checkFunc: func(t *testing.T, address *models.Address) {
				assert.NotEqual(t, uuid.Nil, address.ID)
				assert.False(t, address.CreatedAt.IsZero())
				assert.False(t, address.UpdatedAt.IsZero())
			},
		},
		{
			name: "Create address with work type",
			setup: func(t *testing.T) {
				user := createTestUserForAddress(t, ctx)
				addressToCreate = createTestAddress(user.ID)
				addressToCreate.AddressType = models.AddressTypeWork
			},
			expectedError: false,
			checkFunc: func(t *testing.T, address *models.Address) {
				assert.Equal(t, models.AddressTypeWork, address.AddressType)
			},
		},
		{
			name: "Create address with invalid user ID",
			setup: func(t *testing.T) {
				addressToCreate = createTestAddress(uuid.New())
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

			err := repo.Create(ctx, addressToCreate)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, addressToCreate)
			}
		})
	}
}

func TestAddressRepository_ListByUserID(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewAddressRepository(testDB.Pool)

	var userIDToFind uuid.UUID

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		expectedError bool
		checkFunc     func(t *testing.T, addresses []*models.Address)
	}{
		{
			name: "List addresses for user with multiple addresses",
			setup: func(t *testing.T) {
				user := createTestUserForAddress(t, ctx)
				userIDToFind = user.ID

				for i := 0; i < 3; i++ {
					addr := createTestAddress(user.ID)
					addr.AddressLine1 = "Address " + string(rune('A'+i))
					require.NoError(t, repo.Create(ctx, addr))
				}
			},
			expectedError: false,
			checkFunc: func(t *testing.T, addresses []*models.Address) {
				assert.Len(t, addresses, 3)
			},
		},
		{
			name: "List addresses for user with no addresses",
			setup: func(t *testing.T) {
				user := createTestUserForAddress(t, ctx)
				userIDToFind = user.ID
			},
			expectedError: false,
			checkFunc: func(t *testing.T, addresses []*models.Address) {
				assert.Empty(t, addresses)
			},
		},
		{
			name: "List addresses for non-existing user",
			setup: func(t *testing.T) {
				userIDToFind = uuid.New()
			},
			expectedError: false,
			checkFunc: func(t *testing.T, addresses []*models.Address) {
				assert.Empty(t, addresses)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testDB.CleanupTestData(ctx))

			if tt.setup != nil {
				tt.setup(t)
			}

			result, err := repo.ListByUserID(ctx, userIDToFind)

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

func TestAddressRepository_GetByIDAndUserID(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewAddressRepository(testDB.Pool)

	var addressID uuid.UUID
	var userID uuid.UUID

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		expectedError bool
		checkFunc     func(t *testing.T, address *models.Address)
	}{
		{
			name: "Get existing address by ID and user ID",
			setup: func(t *testing.T) {
				user := createTestUserForAddress(t, ctx)
				userID = user.ID

				addr := createTestAddress(user.ID)
				require.NoError(t, repo.Create(ctx, addr))
				addressID = addr.ID
			},
			expectedError: false,
			checkFunc: func(t *testing.T, address *models.Address) {
				assert.NotNil(t, address)
				assert.Equal(t, addressID, address.ID)
				assert.Equal(t, userID, address.UserID)
			},
		},
		{
			name: "Get address with wrong user ID",
			setup: func(t *testing.T) {
				user := createTestUserForAddress(t, ctx)

				addr := createTestAddress(user.ID)
				require.NoError(t, repo.Create(ctx, addr))
				addressID = addr.ID
				userID = uuid.New()
			},
			expectedError: false,
			checkFunc: func(t *testing.T, address *models.Address) {
				assert.Nil(t, address)
			},
		},
		{
			name: "Get non-existing address",
			setup: func(t *testing.T) {
				addressID = uuid.New()
				userID = uuid.New()
			},
			expectedError: false,
			checkFunc: func(t *testing.T, address *models.Address) {
				assert.Nil(t, address)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, testDB.CleanupTestData(ctx))

			if tt.setup != nil {
				tt.setup(t)
			}

			result, err := repo.GetByIDAndUserID(ctx, addressID, userID)

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

func TestAddressRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewAddressRepository(testDB.Pool)

	var addressToUpdate *models.Address

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		updateFunc    func(a *models.Address)
		expectedError bool
		rowsAffected  int64
		checkFunc     func(t *testing.T, address *models.Address)
	}{
		{
			name: "Update address line",
			setup: func(t *testing.T) {
				user := createTestUserForAddress(t, ctx)
				addressToUpdate = createTestAddress(user.ID)
				require.NoError(t, repo.Create(ctx, addressToUpdate))
			},
			updateFunc: func(a *models.Address) {
				a.AddressLine1 = "456 Updated Street"
			},
			expectedError: false,
			rowsAffected:  1,
			checkFunc: func(t *testing.T, address *models.Address) {
				require.NotNil(t, address)
				assert.Equal(t, "456 Updated Street", address.AddressLine1)
			},
		},
		{
			name: "Update phone and full name",
			setup: func(t *testing.T) {
				user := createTestUserForAddress(t, ctx)
				addressToUpdate = createTestAddress(user.ID)
				require.NoError(t, repo.Create(ctx, addressToUpdate))
			},
			updateFunc: func(a *models.Address) {
				a.Phone = "0987654321"
				a.FullName = "Updated Name"
			},
			expectedError: false,
			rowsAffected:  1,
			checkFunc: func(t *testing.T, address *models.Address) {
				require.NotNil(t, address)
				assert.Equal(t, "0987654321", address.Phone)
				assert.Equal(t, "Updated Name", address.FullName)
			},
		},
		{
			name: "Update address type",
			setup: func(t *testing.T) {
				user := createTestUserForAddress(t, ctx)
				addressToUpdate = createTestAddress(user.ID)
				require.NoError(t, repo.Create(ctx, addressToUpdate))
			},
			updateFunc: func(a *models.Address) {
				a.AddressType = models.AddressTypeWork
			},
			expectedError: false,
			rowsAffected:  1,
			checkFunc: func(t *testing.T, address *models.Address) {
				require.NotNil(t, address)
				assert.Equal(t, models.AddressTypeWork, address.AddressType)
			},
		},
		{
			name: "Update non-existing address",
			setup: func(t *testing.T) {
				user := createTestUserForAddress(t, ctx)
				addressToUpdate = createTestAddress(user.ID)
				addressToUpdate.ID = uuid.New()
			},
			updateFunc: func(a *models.Address) {
				a.FullName = "Should Not Update"
			},
			expectedError: false,
			rowsAffected:  0,
			checkFunc: func(t *testing.T, address *models.Address) {
				assert.Nil(t, address)
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
				tt.updateFunc(addressToUpdate)
			}

			rowsAffected, err := repo.Update(ctx, addressToUpdate)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.Equal(t, tt.rowsAffected, rowsAffected)

			require.NoError(t, err)

			updatedAddr, err := repo.GetByIDAndUserID(ctx, addressToUpdate.ID, addressToUpdate.UserID)
			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, updatedAddr)
			}
		})
	}
}

func TestAddressRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewAddressRepository(testDB.Pool)

	var addressIDToDelete uuid.UUID
	var userID uuid.UUID

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		expectedError bool
		rowsAffected  int64
		checkFunc     func(t *testing.T)
	}{
		{
			name: "Delete existing address",
			setup: func(t *testing.T) {
				user := createTestUserForAddress(t, ctx)
				userID = user.ID

				addr := createTestAddress(user.ID)
				require.NoError(t, repo.Create(ctx, addr))
				addressIDToDelete = addr.ID
			},
			expectedError: false,
			rowsAffected:  1,
			checkFunc: func(t *testing.T) {
				deletedAddr, err := repo.GetByIDAndUserID(ctx, addressIDToDelete, userID)
				require.NoError(t, err)
				assert.Nil(t, deletedAddr, "Address should be deleted")
			},
		},
		{
			name: "Delete non-existing address",
			setup: func(t *testing.T) {
				addressIDToDelete = uuid.New()
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

			rowsAffected, err := repo.Delete(ctx, addressIDToDelete)

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

func TestAddressRepository_SetDefaultAddress(t *testing.T) {
	ctx := context.Background()
	repo := impl.NewAddressRepository(testDB.Pool)

	var userID uuid.UUID
	var addressIDToSetDefault uuid.UUID

	tests := []struct {
		name          string
		setup         func(t *testing.T)
		expectedError bool
		rowsAffected  int64
		checkFunc     func(t *testing.T)
	}{
		{
			name: "Set default address success",
			setup: func(t *testing.T) {
				user := createTestUserForAddress(t, ctx)
				userID = user.ID

				addr1 := createTestAddress(user.ID)
				require.NoError(t, repo.Create(ctx, addr1))

				addr2 := createTestAddress(user.ID)
				addr2.AddressLine1 = "Second Address"
				require.NoError(t, repo.Create(ctx, addr2))

				addressIDToSetDefault = addr2.ID
			},
			expectedError: false,
			rowsAffected:  2,
			checkFunc: func(t *testing.T) {
				addr, err := repo.GetByIDAndUserID(ctx, addressIDToSetDefault, userID)
				require.NoError(t, err)
				require.NotNil(t, addr)
				assert.True(t, addr.IsDefault, "Address should be set as default")
			},
		},
		{
			name: "Set default for non-existing address",
			setup: func(t *testing.T) {
				user := createTestUserForAddress(t, ctx)
				userID = user.ID
				addressIDToSetDefault = uuid.New()
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

			rowsAffected, err := repo.SetDefaultAddress(ctx, userID, addressIDToSetDefault)

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
