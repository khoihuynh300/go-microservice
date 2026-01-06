package api_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	mdkeys "github.com/khoihuynh300/go-microservice/shared/pkg/const/metadata"
	userpb "github.com/khoihuynh300/go-microservice/shared/proto/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	md := metadata.Pairs(mdkeys.UserIDHeader, userID)
	return metadata.NewOutgoingContext(ctx, md)
}

func TestUserAPI_GetMe(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		setup        func(t *testing.T) string
		expectedCode codes.Code
		checkFunc    func(t *testing.T, resp *userpb.GetUserResponse)
	}{
		{
			name: "Get me success",
			setup: func(t *testing.T) string {
				user := CreateVerifiedUser(ctx, t, "getme@test.com", "Password123!", "Get Me User")
				return user.ID.String()
			},
			expectedCode: codes.OK,
			checkFunc: func(t *testing.T, resp *userpb.GetUserResponse) {
				assert.Equal(t, "getme@test.com", resp.User.Email)
				assert.Equal(t, "Get Me User", resp.User.FullName)
			},
		},
		{
			name: "Get me without auth",
			setup: func(t *testing.T) string {
				return ""
			},
			expectedCode: codes.Unauthenticated,
			checkFunc:    nil,
		},
		{
			name: "Get me with non-existing user",
			setup: func(t *testing.T) string {
				return uuid.New().String()
			},
			expectedCode: codes.NotFound,
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

			resp, err := client.GetMe(reqCtx, &emptypb.Empty{})

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

func TestUserAPI_GetUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		setup        func(t *testing.T) (authUserID string, targetUserID string)
		expectedCode codes.Code
		checkFunc    func(t *testing.T, resp *userpb.GetPublicUserResponse)
	}{
		{
			name: "Get public user profile success",
			setup: func(t *testing.T) (string, string) {
				authUser := CreateVerifiedUser(ctx, t, "auth@test.com", "Password123!", "Auth User")
				targetUser := CreateVerifiedUser(ctx, t, "target@test.com", "Password123!", "Target User")
				return authUser.ID.String(), targetUser.ID.String()
			},
			expectedCode: codes.OK,
			checkFunc: func(t *testing.T, resp *userpb.GetPublicUserResponse) {
				assert.Equal(t, "Target User", resp.User.FullName)
			},
		},
		{
			name: "Get non-existing user",
			setup: func(t *testing.T) (string, string) {
				authUser := CreateVerifiedUser(ctx, t, "auth2@test.com", "Password123!", "Auth User")
				return authUser.ID.String(), uuid.New().String()
			},
			expectedCode: codes.NotFound,
			checkFunc:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, cleanupTestData(ctx))

			authUserID, targetUserID := tt.setup(t)
			reqCtx := ContextWithUserID(ctx, authUserID)

			resp, err := client.GetUser(reqCtx, &userpb.GetUserRequest{
				UserId: targetUserID,
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

func TestUserAPI_UpdateUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		setup        func(t *testing.T) string
		request      *userpb.UpdateUserRequest
		expectedCode codes.Code
		checkFunc    func(t *testing.T, resp *userpb.UpdateUserResponse)
	}{
		{
			name: "Update full name success",
			setup: func(t *testing.T) string {
				user := CreateVerifiedUser(ctx, t, "update@test.com", "Password123!", "Original Name")
				return user.ID.String()
			},
			request: &userpb.UpdateUserRequest{
				FullName: strPtr("Updated Name"),
			},
			expectedCode: codes.OK,
			checkFunc: func(t *testing.T, resp *userpb.UpdateUserResponse) {
				assert.Equal(t, "Updated Name", resp.User.FullName)
			},
		},
		{
			name: "Update date of birth success",
			setup: func(t *testing.T) string {
				user := CreateVerifiedUser(ctx, t, "dob@test.com", "Password123!", "DOB User")
				return user.ID.String()
			},
			request: &userpb.UpdateUserRequest{
				DateOfBirth: strPtr("15-06-1990"),
			},
			expectedCode: codes.OK,
			checkFunc: func(t *testing.T, resp *userpb.UpdateUserResponse) {
				assert.Equal(t, "15-06-1990", resp.User.DateOfBirth.GetValue())
			},
		},
		{
			name: "Update gender success",
			setup: func(t *testing.T) string {
				user := CreateVerifiedUser(ctx, t, "gender@test.com", "Password123!", "Gender User")
				return user.ID.String()
			},
			request: &userpb.UpdateUserRequest{
				Gender: strPtr("male"),
			},
			expectedCode: codes.OK,
			checkFunc: func(t *testing.T, resp *userpb.UpdateUserResponse) {
				assert.Equal(t, "male", resp.User.Gender.GetValue())
			},
		},
		{
			name: "Update with invalid date format",
			setup: func(t *testing.T) string {
				user := CreateVerifiedUser(ctx, t, "invalid@test.com", "Password123!", "Invalid User")
				return user.ID.String()
			},
			request: &userpb.UpdateUserRequest{
				DateOfBirth: strPtr("1990-06-15"),
			},
			expectedCode: codes.InvalidArgument,
			checkFunc:    nil,
		},
		{
			name: "Update with invalid gender",
			setup: func(t *testing.T) string {
				user := CreateVerifiedUser(ctx, t, "invalidgender@test.com", "Password123!", "Invalid Gender User")
				return user.ID.String()
			},
			request: &userpb.UpdateUserRequest{
				Gender: strPtr("invalid"),
			},
			expectedCode: codes.InvalidArgument,
			checkFunc:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, cleanupTestData(ctx))

			userID := tt.setup(t)
			reqCtx := ContextWithUserID(ctx, userID)

			resp, err := client.UpdateUser(reqCtx, tt.request)

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

func TestUserAPI_ChangePassword(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		setup           func(t *testing.T) string
		currentPassword string
		newPassword     string
		expectedCode    codes.Code
	}{
		{
			name: "Change password success",
			setup: func(t *testing.T) string {
				user := CreateVerifiedUser(ctx, t, "changepass@test.com", "Password123!", "Change Pass User")
				return user.ID.String()
			},
			currentPassword: "Password123!",
			newPassword:     "NewPassword456!",
			expectedCode:    codes.OK,
		},
		{
			name: "Change password with wrong current password",
			setup: func(t *testing.T) string {
				user := CreateVerifiedUser(ctx, t, "wrongcurrent@test.com", "Password123!", "Wrong Current User")
				return user.ID.String()
			},
			currentPassword: "WrongPassword!",
			newPassword:     "NewPassword456!",
			expectedCode:    codes.Unauthenticated,
		},
		{
			name: "Change password with short new password",
			setup: func(t *testing.T) string {
				user := CreateVerifiedUser(ctx, t, "shortpass@test.com", "Password123!", "Short Pass User")
				return user.ID.String()
			},
			currentPassword: "Password123!",
			newPassword:     "short",
			expectedCode:    codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, cleanupTestData(ctx))

			userID := tt.setup(t)
			reqCtx := ContextWithUserID(ctx, userID)

			_, err := client.ChangePassword(reqCtx, &userpb.ChangePasswordRequest{
				CurrentPassword: tt.currentPassword,
				NewPassword:     tt.newPassword,
			})

			if tt.expectedCode != codes.OK {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
				return
			}

			require.NoError(t, err)

			_, err = client.Login(ctx, &userpb.LoginRequest{
				Email:    "changepass@test.com",
				Password: tt.newPassword,
			})
			require.NoError(t, err)
		})
	}
}

func strPtr(s string) *string {
	return &s
}
