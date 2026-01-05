package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/shared/pkg/const/contextkeys"
	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/dto/request"
	"github.com/khoihuynh300/go-microservice/user-service/internal/service"
	mock_repository "github.com/khoihuynh300/go-microservice/user-service/mocks/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type UserServiceTestSuite struct {
	ctrl        *gomock.Controller
	userRepo    *mock_repository.MockUserRepository
	userService *service.UserService
}

func NewUserServiceTestSuite(t *testing.T) *UserServiceTestSuite {
	ctrl := gomock.NewController(t)
	userRepo := mock_repository.NewMockUserRepository(ctrl)
	userService := service.NewUserService(userRepo)
	return &UserServiceTestSuite{
		ctrl:        ctrl,
		userRepo:    userRepo,
		userService: userService,
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		name          string
		userID        string
		setupMock     func(suite *UserServiceTestSuite)
		expectedError error
		checkFunc     func(t *testing.T, user *models.User, err error)
	}{
		{
			name:   "Get User Success",
			userID: testUserID.String(),
			setupMock: func(s *UserServiceTestSuite) {
				user := &models.User{
					ID:       testUserID,
					Email:    "test@gmail.com",
					FullName: "Test User",
					Status:   models.UserStatusActive,
				}
				s.userRepo.EXPECT().FindByID(gomock.Any(), testUserID).Return(user, nil)
			},
			expectedError: nil,
			checkFunc: func(t *testing.T, user *models.User, err error) {
				assert.NotNil(t, user)
				assert.Equal(t, testUserID, user.ID)
				assert.Equal(t, "test@gmail.com", user.Email)
			},
		},
		{
			name:   "User Not Found",
			userID: testUserID.String(),
			setupMock: func(s *UserServiceTestSuite) {
				s.userRepo.EXPECT().FindByID(gomock.Any(), testUserID).Return(nil, nil)
			},
			expectedError: apperr.ErrUserNotFound,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := NewUserServiceTestSuite(t)
			defer suite.ctrl.Finish()

			ctx := context.WithValue(context.Background(), contextkeys.LoggerKey, zap.NewNop())
			tt.setupMock(suite)

			user, err := suite.userService.GetUserByID(ctx, tt.userID)

			assert.True(t, errors.Is(err, tt.expectedError))

			if tt.checkFunc != nil {
				tt.checkFunc(t, user, err)
			}
		})
	}
}

func TestUserService_UpdateUser(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		name          string
		userID        string
		req           *request.UpdateUserRequest
		setupMock     func(suite *UserServiceTestSuite)
		expectedError error
		checkFunc     func(t *testing.T, user *models.User, err error)
	}{
		{
			name:   "Update User Success - Full Update",
			userID: testUserID.String(),
			req: &request.UpdateUserRequest{
				FullName:    ptrString("Updated Name"),
				DateOfBirth: ptrTime(time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)),
				Gender:      ptrString("male"),
			},
			setupMock: func(s *UserServiceTestSuite) {
				user := &models.User{
					ID:       testUserID,
					Email:    "test@gmail.com",
					FullName: "Old Name",
					Status:   models.UserStatusActive,
				}
				s.userRepo.EXPECT().FindByID(gomock.Any(), testUserID).Return(user, nil)
				s.userRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedError: nil,
			checkFunc: func(t *testing.T, user *models.User, err error) {
				assert.NotNil(t, user)
				assert.Equal(t, "Updated Name", user.FullName)
				assert.Equal(t, time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), *user.DateOfBirth)
				assert.Equal(t, models.GenderMale, *user.Gender)
			},
		},
		{
			name:   "Update User Success - Partial Update (FullName only)",
			userID: testUserID.String(),
			req: &request.UpdateUserRequest{
				FullName: ptrString("New Name"),
			},
			setupMock: func(s *UserServiceTestSuite) {
				user := &models.User{
					ID:       testUserID,
					Email:    "test@gmail.com",
					FullName: "Old Name",
					Status:   models.UserStatusActive,
				}
				s.userRepo.EXPECT().FindByID(gomock.Any(), testUserID).Return(user, nil)
				s.userRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedError: nil,
			checkFunc: func(t *testing.T, user *models.User, err error) {
				assert.NotNil(t, user)
				assert.Equal(t, "New Name", user.FullName)
			},
		},
		{
			name:   "User Not Found",
			userID: testUserID.String(),
			req: &request.UpdateUserRequest{
				FullName: ptrString("New Name"),
			},
			setupMock: func(s *UserServiceTestSuite) {
				s.userRepo.EXPECT().FindByID(gomock.Any(), testUserID).Return(nil, nil)
			},
			expectedError: apperr.ErrUserNotFound,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := NewUserServiceTestSuite(t)
			defer suite.ctrl.Finish()

			ctx := context.WithValue(context.Background(), contextkeys.LoggerKey, zap.NewNop())
			tt.setupMock(suite)

			user, err := suite.userService.UpdateUser(ctx, tt.userID, tt.req)

			assert.True(t, errors.Is(err, tt.expectedError))

			if tt.checkFunc != nil {
				tt.checkFunc(t, user, err)
			}
		})
	}
}

func ptrString(s string) *string {
	return &s
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
