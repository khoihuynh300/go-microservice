package api_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	userpb "github.com/khoihuynh300/go-microservice/shared/proto/user"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestUser struct {
	ID       uuid.UUID
	Email    string
	Password string
	FullName string
}

func TestAuthAPI_Register(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		setup        func(t *testing.T)
		request      *userpb.RegisterRequest
		expectedCode codes.Code
		checkFunc    func(t *testing.T, resp *userpb.RegisterResponse)
	}{
		{
			name:  "Register success",
			setup: nil,
			request: &userpb.RegisterRequest{
				Email:    "test@gmail.com",
				Password: "Password123!",
				FullName: "New User",
			},
			expectedCode: codes.OK,
			checkFunc: func(t *testing.T, resp *userpb.RegisterResponse) {
				assert.NotEmpty(t, resp.UserId)
			},
		},
		{
			name: "Register with existing email",
			setup: func(t *testing.T) {
				CreateVerifiedUser(ctx, t, "existing@test.com", "Password123!", "Existing User")
			},
			request: &userpb.RegisterRequest{
				Email:    "existing@test.com",
				Password: "Password123!",
				FullName: "Another User",
			},
			expectedCode: codes.AlreadyExists,
			checkFunc:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, cleanupTestData(ctx))

			if tt.setup != nil {
				tt.setup(t)
			}

			resp, err := client.Register(ctx, tt.request)

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

func TestAuthAPI_Login(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		setup        func(t *testing.T)
		request      *userpb.LoginRequest
		expectedCode codes.Code
		checkFunc    func(t *testing.T, resp *userpb.TokenResponse)
	}{
		{
			name: "Login success",
			setup: func(t *testing.T) {
				CreateVerifiedUser(ctx, t, "login@test.com", "Password123!", "Login User")
			},
			request: &userpb.LoginRequest{
				Email:    "login@test.com",
				Password: "Password123!",
			},
			expectedCode: codes.OK,
			checkFunc: func(t *testing.T, resp *userpb.TokenResponse) {
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			},
		},
		{
			name: "Login with wrong password",
			setup: func(t *testing.T) {
				CreateVerifiedUser(ctx, t, "wrongpass@test.com", "Password123!", "Wrong Pass User")
			},
			request: &userpb.LoginRequest{
				Email:    "wrongpass@test.com",
				Password: "WrongPassword!",
			},
			expectedCode: codes.Unauthenticated,
			checkFunc:    nil,
		},
		{
			name:  "Login with non-existing user",
			setup: nil,
			request: &userpb.LoginRequest{
				Email:    "nonexistent@test.com",
				Password: "Password123!",
			},
			expectedCode: codes.Unauthenticated,
			checkFunc:    nil,
		},
		{
			name: "Login with unverified account",
			setup: func(t *testing.T) {
				_, err := client.Register(ctx, &userpb.RegisterRequest{
					Email:    "unverified@test.com",
					Password: "Password123!",
					FullName: "Unverified User",
				})
				require.NoError(t, err)
			},
			request: &userpb.LoginRequest{
				Email:    "unverified@test.com",
				Password: "Password123!",
			},
			expectedCode: codes.PermissionDenied,
			checkFunc:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, cleanupTestData(ctx))

			if tt.setup != nil {
				tt.setup(t)
			}

			resp, err := client.Login(ctx, tt.request)

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

func TestAuthAPI_RefreshToken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		setup        func(t *testing.T) string // returns refresh token
		expectedCode codes.Code
		checkFunc    func(t *testing.T, resp *userpb.TokenResponse)
	}{
		{
			name: "Refresh token success",
			setup: func(t *testing.T) string {
				user := CreateVerifiedUser(ctx, t, "refresh@test.com", "Password123!", "Refresh User")
				tokens := LoginUser(ctx, t, user.Email, user.Password)
				return tokens.RefreshToken
			},
			expectedCode: codes.OK,
			checkFunc: func(t *testing.T, resp *userpb.TokenResponse) {
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			},
		},
		{
			name: "Refresh with invalid token",
			setup: func(t *testing.T) string {
				return "invalid-refresh-token"
			},
			expectedCode: codes.Unauthenticated,
			checkFunc:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, cleanupTestData(ctx))

			refreshToken := tt.setup(t)

			resp, err := client.Refresh(ctx, &userpb.RefreshRequest{
				RefreshToken: refreshToken,
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

func CreateVerifiedUser(ctx context.Context, t *testing.T, email, password, fullName string) *TestUser {
	resp, err := client.Register(ctx, &userpb.RegisterRequest{
		Email:    email,
		Password: password,
		FullName: fullName,
	})
	require.NoError(t, err)

	userID, err := uuid.Parse(resp.UserId)
	require.NoError(t, err)

	_, err = testDB.Pool.Exec(ctx, `
		UPDATE users
		SET status = $1, email_verified_at = NOW()
		WHERE id = $2
	`, models.UserStatusActive, userID)
	require.NoError(t, err)

	return &TestUser{
		ID:       userID,
		Email:    email,
		Password: password,
		FullName: fullName,
	}
}

func LoginUser(ctx context.Context, t *testing.T, email, password string) *userpb.TokenResponse {
	resp, err := client.Login(ctx, &userpb.LoginRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	return resp
}
