package api_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	userpb "github.com/khoihuynh300/go-microservice/shared/proto/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestAddressAPI_CreateUserAddress(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		setup        func(t *testing.T) string
		request      *userpb.CreateUserAddressRequest
		expectedCode codes.Code
		checkFunc    func(t *testing.T, resp *userpb.CreateUserAddressResponse)
	}{
		{
			name: "Create address success",
			setup: func(t *testing.T) string {
				user := CreateVerifiedUser(ctx, t, "test@gmail.com", "Password123!", "Test User")
				return user.ID.String()
			},
			request: &userpb.CreateUserAddressRequest{
				AddressType:  "home",
				FullName:     "John Doe",
				Phone:        "0901234567",
				AddressLine1: "123 Main Street",
				AddressLine2: "Room 1",
				Ward:         "Ward 1",
				City:         "Ho Chi Minh",
				Country:      "Vietnam",
				IsDefault:    true,
			},
			expectedCode: codes.OK,
			checkFunc: func(t *testing.T, resp *userpb.CreateUserAddressResponse) {
				assert.NotEmpty(t, resp.Address.Id)
				assert.Equal(t, "home", resp.Address.AddressType)
				assert.Equal(t, "John Doe", resp.Address.FullName)
				assert.Equal(t, "0901234567", resp.Address.Phone)
				assert.True(t, resp.Address.IsDefault)
			},
		},
		{
			name: "Create address with invalid address type",
			setup: func(t *testing.T) string {
				user := CreateVerifiedUser(ctx, t, "test@gmail.com", "Password123!", "Test User")
				return user.ID.String()
			},
			request: &userpb.CreateUserAddressRequest{
				AddressType:  "invalid",
				FullName:     "John Doe",
				Phone:        "0901234567",
				AddressLine1: "123 Main Street",
				Ward:         "Ward 1",
				City:         "Ho Chi Minh",
				Country:      "Vietnam",
			},
			expectedCode: codes.InvalidArgument,
			checkFunc:    nil,
		},
		{
			name: "Create address with invalid phone",
			setup: func(t *testing.T) string {
				user := CreateVerifiedUser(ctx, t, "test@gmail.com", "Password123!", "Test User")
				return user.ID.String()
			},
			request: &userpb.CreateUserAddressRequest{
				AddressType:  "home",
				FullName:     "John Doe",
				Phone:        "123",
				AddressLine1: "123 Main Street",
				Ward:         "Ward 1",
				City:         "Ho Chi Minh",
				Country:      "Vietnam",
			},
			expectedCode: codes.InvalidArgument,
			checkFunc:    nil,
		},
		{
			name: "Create address without auth",
			setup: func(t *testing.T) string {
				return ""
			},
			request: &userpb.CreateUserAddressRequest{
				AddressType:  "home",
				FullName:     "John Doe",
				Phone:        "0901234567",
				AddressLine1: "123 Main Street",
				Ward:         "Ward 1",
				City:         "Ho Chi Minh",
				Country:      "Vietnam",
			},
			expectedCode: codes.Unauthenticated,
			checkFunc:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, cleanupTestData(ctx))

			userID := tt.setup(t)

			var reqCtx context.Context
			if userID == "" {
				reqCtx = ctx
			} else {
				reqCtx = ContextWithUserID(ctx, userID)
			}

			resp, err := client.CreateUserAddress(reqCtx, tt.request)

			if tt.expectedCode != codes.OK {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				return
			}

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, resp)
			}
		})
	}
}

func TestAddressAPI_GetUserAddresses(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		setup        func(t *testing.T) string
		expectedCode codes.Code
		checkFunc    func(t *testing.T, resp *userpb.GetUserAddressesResponse)
	}{
		{
			name: "Get addresses success - empty",
			setup: func(t *testing.T) string {
				user := CreateVerifiedUser(ctx, t, "test@gmail.com", "Password123!", "Test User")
				return user.ID.String()
			},
			expectedCode: codes.OK,
			checkFunc: func(t *testing.T, resp *userpb.GetUserAddressesResponse) {
				assert.Len(t, resp.Addresses, 0)
			},
		},
		{
			name: "Get addresses success - with addresses",
			setup: func(t *testing.T) string {
				user := CreateVerifiedUser(ctx, t, "test@gmail.com", "Password123!", "Test User")
				userID := user.ID.String()
				reqCtx := ContextWithUserID(ctx, userID)

				_, err := client.CreateUserAddress(reqCtx, &userpb.CreateUserAddressRequest{
					AddressType:  "home",
					FullName:     "Home Address",
					Phone:        "0901234567",
					AddressLine1: "Home Street",
					Ward:         "Ward 1",
					City:         "Ho Chi Minh",
					Country:      "Vietnam",
					IsDefault:    true,
				})
				require.NoError(t, err)

				_, err = client.CreateUserAddress(reqCtx, &userpb.CreateUserAddressRequest{
					AddressType:  "work",
					FullName:     "Work Address",
					Phone:        "0907654321",
					AddressLine1: "Work Street",
					Ward:         "Ward 2",
					City:         "Ha Noi",
					Country:      "Vietnam",
				})
				require.NoError(t, err)

				return userID
			},
			expectedCode: codes.OK,
			checkFunc: func(t *testing.T, resp *userpb.GetUserAddressesResponse) {
				assert.Len(t, resp.Addresses, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, cleanupTestData(ctx))

			userID := tt.setup(t)
			reqCtx := ContextWithUserID(ctx, userID)

			resp, err := client.GetUserAddresses(reqCtx, &emptypb.Empty{})

			if tt.expectedCode != codes.OK {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				return
			}

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, resp)
			}
		})
	}
}

func TestAddressAPI_GetUserAddress(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		setup        func(t *testing.T) (userID string, addressID string)
		expectedCode codes.Code
		checkFunc    func(t *testing.T, resp *userpb.GetUserAddressResponse)
	}{
		{
			name: "Get address success",
			setup: func(t *testing.T) (string, string) {
				user := CreateVerifiedUser(ctx, t, "test@gmail.com", "Password123!", "Test User")
				userID := user.ID.String()
				reqCtx := ContextWithUserID(ctx, userID)

				resp, err := client.CreateUserAddress(reqCtx, &userpb.CreateUserAddressRequest{
					AddressType:  "home",
					FullName:     "My Address",
					Phone:        "0901234567",
					AddressLine1: "123 Street",
					Ward:         "Ward 1",
					City:         "Ho Chi Minh",
					Country:      "Vietnam",
				})
				require.NoError(t, err)

				return userID, resp.Address.Id
			},
			expectedCode: codes.OK,
			checkFunc: func(t *testing.T, resp *userpb.GetUserAddressResponse) {
				assert.Equal(t, "My Address", resp.Address.FullName)
			},
		},
		{
			name: "Get non-existing address",
			setup: func(t *testing.T) (string, string) {
				user := CreateVerifiedUser(ctx, t, "test@gmail.com", "Password123!", "Test User")
				return user.ID.String(), uuid.New().String()
			},
			expectedCode: codes.NotFound,
			checkFunc:    nil,
		},
		{
			name: "Get another user's address",
			setup: func(t *testing.T) (string, string) {
				user1 := CreateVerifiedUser(ctx, t, "user1@gmail.com", "Password123!", "User 1")
				reqCtx1 := ContextWithUserID(ctx, user1.ID.String())

				resp, err := client.CreateUserAddress(reqCtx1, &userpb.CreateUserAddressRequest{
					AddressType:  "home",
					FullName:     "User1 Address",
					Phone:        "0901234567",
					AddressLine1: "123 Street",
					Ward:         "Ward 1",
					City:         "Ho Chi Minh",
					Country:      "Vietnam",
				})
				require.NoError(t, err)

				user2 := CreateVerifiedUser(ctx, t, "user2@gmail.com", "Password123!", "User 2")
				return user2.ID.String(), resp.Address.Id
			},
			expectedCode: codes.NotFound, // Should not be able to access another user's address
			checkFunc:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, cleanupTestData(ctx))

			userID, addressID := tt.setup(t)
			reqCtx := ContextWithUserID(ctx, userID)

			resp, err := client.GetUserAddress(reqCtx, &userpb.GetUserAddressRequest{
				AddressId: addressID,
			})

			if tt.expectedCode != codes.OK {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				return
			}

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, resp)
			}
		})
	}
}

func TestAddressAPI_UpdateUserAddress(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		setup        func(t *testing.T) (userID string, addressID string)
		updateReq    func(addressID string) *userpb.UpdateUserAddressRequest
		expectedCode codes.Code
		checkFunc    func(t *testing.T, resp *userpb.UpdateUserAddressResponse)
	}{
		{
			name: "Update address success",
			setup: func(t *testing.T) (string, string) {
				user := CreateVerifiedUser(ctx, t, "test@gmail.com", "Password123!", "Test User")
				userID := user.ID.String()
				reqCtx := ContextWithUserID(ctx, userID)

				resp, err := client.CreateUserAddress(reqCtx, &userpb.CreateUserAddressRequest{
					AddressType:  "home",
					FullName:     "Original Name",
					Phone:        "0901234567",
					AddressLine1: "Original Street",
					Ward:         "Ward 1",
					City:         "Ho Chi Minh",
					Country:      "Vietnam",
				})
				require.NoError(t, err)

				return userID, resp.Address.Id
			},
			updateReq: func(addressID string) *userpb.UpdateUserAddressRequest {
				return &userpb.UpdateUserAddressRequest{
					AddressId:    addressID,
					FullName:     strPtr("Updated Name"),
					AddressLine1: strPtr("Updated Street"),
				}
			},
			expectedCode: codes.OK,
			checkFunc: func(t *testing.T, resp *userpb.UpdateUserAddressResponse) {
				assert.Equal(t, "Updated Name", resp.Address.FullName)
				assert.Equal(t, "Updated Street", resp.Address.AddressLine1)
			},
		},
		{
			name: "Update non-existing address",
			setup: func(t *testing.T) (string, string) {
				user := CreateVerifiedUser(ctx, t, "test@gmail.com", "Password123!", "Test User")
				return user.ID.String(), uuid.New().String()
			},
			updateReq: func(addressID string) *userpb.UpdateUserAddressRequest {
				return &userpb.UpdateUserAddressRequest{
					AddressId: addressID,
					FullName:  strPtr("Updated Name"),
				}
			},
			expectedCode: codes.NotFound,
			checkFunc:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, cleanupTestData(ctx))

			userID, addressID := tt.setup(t)
			reqCtx := ContextWithUserID(ctx, userID)

			resp, err := client.UpdateUserAddress(reqCtx, tt.updateReq(addressID))

			if tt.expectedCode != codes.OK {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				return
			}

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, resp)
			}
		})
	}
}

func TestAddressAPI_DeleteUserAddress(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		setup        func(t *testing.T) (userID string, addressID string)
		expectedCode codes.Code
	}{
		{
			name: "Delete address success",
			setup: func(t *testing.T) (string, string) {
				user := CreateVerifiedUser(ctx, t, "test@gmail.com", "Password123!", "Test User")
				userID := user.ID.String()
				reqCtx := ContextWithUserID(ctx, userID)

				resp, err := client.CreateUserAddress(reqCtx, &userpb.CreateUserAddressRequest{
					AddressType:  "home",
					FullName:     "Test",
					Phone:        "0901234567",
					AddressLine1: "123 Street",
					Ward:         "Ward 1",
					City:         "Ho Chi Minh",
					Country:      "Vietnam",
				})
				require.NoError(t, err)

				return userID, resp.Address.Id
			},
			expectedCode: codes.OK,
		},
		{
			name: "Delete non-existing address",
			setup: func(t *testing.T) (string, string) {
				user := CreateVerifiedUser(ctx, t, "test@gmail.com", "Password123!", "Test User")
				return user.ID.String(), uuid.New().String()
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, cleanupTestData(ctx))

			userID, addressID := tt.setup(t)
			reqCtx := ContextWithUserID(ctx, userID)

			_, err := client.DeleteUserAddress(reqCtx, &userpb.DeleteUserAddressRequest{
				AddressId: addressID,
			})

			if tt.expectedCode != codes.OK {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				return
			}

			require.NoError(t, err)

			_, err = client.GetUserAddress(reqCtx, &userpb.GetUserAddressRequest{
				AddressId: addressID,
			})
			require.Error(t, err)
			st, _ := status.FromError(err)
			assert.Equal(t, codes.NotFound, st.Code())
		})
	}
}
