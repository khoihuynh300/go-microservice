package service_test

import (
	"context"
	"errors"
	"testing"

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

type AddressServiceTestSuite struct {
	ctrl           *gomock.Controller
	userRepo       *mock_repository.MockUserRepository
	addressRepo    *mock_repository.MockAddressRepository
	addressService service.AddressService
}

func NewAddressServiceTestSuite(t *testing.T) *AddressServiceTestSuite {
	ctrl := gomock.NewController(t)
	userRepo := mock_repository.NewMockUserRepository(ctrl)
	addressRepo := mock_repository.NewMockAddressRepository(ctrl)
	addressService := service.NewAddressService(userRepo, addressRepo)
	return &AddressServiceTestSuite{
		ctrl:           ctrl,
		userRepo:       userRepo,
		addressRepo:    addressRepo,
		addressService: addressService,
	}
}

func TestAddressService_CreateUserAddress(t *testing.T) {
	testUserID := uuid.New()
	testAddressID := uuid.New()

	tests := []struct {
		name          string
		userID        string
		req           *request.CreateUserAddressRequest
		setupMock     func(suite *AddressServiceTestSuite)
		expectedError error
		checkFunc     func(t *testing.T, address *models.Address, err error)
	}{
		{
			name:   "Create Address Success",
			userID: testUserID.String(),
			req: &request.CreateUserAddressRequest{
				AddressType:  "home",
				FullName:     "John Doe",
				Phone:        "0123456789",
				AddressLine1: "123 Main St",
				AddressLine2: "Room 4",
				Ward:         "Ward 1",
				City:         "Ho Chi Minh",
				Country:      "Vietnam",
				IsDefault:    false,
			},
			setupMock: func(s *AddressServiceTestSuite) {
				user := &models.User{ID: testUserID, Email: "test@gmail.com", Status: models.UserStatusActive}
				s.userRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(user, nil)
				s.addressRepo.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})
				s.addressRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, addr *models.Address) error {
					addr.ID = testAddressID
					return nil
				})
			},
			expectedError: nil,
			checkFunc: func(t *testing.T, address *models.Address, err error) {
				assert.NotNil(t, address)
				assert.Equal(t, "John Doe", address.FullName)
				assert.Equal(t, models.AddressTypeHome, address.AddressType)
			},
		},
		{
			name:   "Create Address With Default Success",
			userID: testUserID.String(),
			req: &request.CreateUserAddressRequest{
				AddressType:  "work",
				FullName:     "John Doe",
				Phone:        "0123456789",
				AddressLine1: "456 Office St",
				City:         "Hanoi",
				Country:      "Vietnam",
				IsDefault:    true,
			},
			setupMock: func(s *AddressServiceTestSuite) {
				user := &models.User{ID: testUserID, Email: "test@gmail.com", Status: models.UserStatusActive}
				s.userRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(user, nil)
				s.addressRepo.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})
				s.addressRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, addr *models.Address) error {
					addr.ID = testAddressID
					return nil
				})
				s.addressRepo.EXPECT().SetDefaultAddress(gomock.Any(), testUserID, testAddressID).Return(int64(1), nil)
			},
			expectedError: nil,
			checkFunc: func(t *testing.T, address *models.Address, err error) {
				assert.NotNil(t, address)
				assert.True(t, address.IsDefault)
			},
		},
		{
			name:   "User Not Found",
			userID: testUserID.String(),
			req: &request.CreateUserAddressRequest{
				AddressType:  "home",
				FullName:     "John Doe",
				Phone:        "0123456789",
				AddressLine1: "123 Main St",
				City:         "Ho Chi Minh",
				Country:      "Vietnam",
			},
			setupMock: func(s *AddressServiceTestSuite) {
				s.userRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(nil, nil)
			},
			expectedError: apperr.ErrUserNotFound,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := NewAddressServiceTestSuite(t)
			defer suite.ctrl.Finish()

			ctx := context.WithValue(context.Background(), contextkeys.LoggerKey, zap.NewNop())
			tt.setupMock(suite)

			address, err := suite.addressService.CreateUserAddress(ctx, tt.userID, tt.req)

			assert.True(t, errors.Is(err, tt.expectedError))

			if tt.checkFunc != nil {
				tt.checkFunc(t, address, err)
			}
		})
	}
}

func TestAddressService_ListUserAddresses(t *testing.T) {
	testUserID := uuid.New()

	tests := []struct {
		name          string
		userID        string
		setupMock     func(suite *AddressServiceTestSuite)
		expectedError error
		checkFunc     func(t *testing.T, addresses []*models.Address, err error)
	}{
		{
			name:   "List Addresses Success",
			userID: testUserID.String(),
			setupMock: func(s *AddressServiceTestSuite) {
				addresses := []*models.Address{
					{ID: uuid.New(), UserID: testUserID, FullName: "Address 1", AddressType: models.AddressTypeHome},
					{ID: uuid.New(), UserID: testUserID, FullName: "Address 2", AddressType: models.AddressTypeWork},
				}
				s.addressRepo.EXPECT().ListByUserID(gomock.Any(), testUserID).Return(addresses, nil)
			},
			expectedError: nil,
			checkFunc: func(t *testing.T, addresses []*models.Address, err error) {
				assert.NotNil(t, addresses)
				assert.Len(t, addresses, 2)
			},
		},
		{
			name:   "List Addresses Empty",
			userID: testUserID.String(),
			setupMock: func(s *AddressServiceTestSuite) {
				s.addressRepo.EXPECT().ListByUserID(gomock.Any(), testUserID).Return([]*models.Address{}, nil)
			},
			expectedError: nil,
			checkFunc: func(t *testing.T, addresses []*models.Address, err error) {
				assert.NotNil(t, addresses)
				assert.Len(t, addresses, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := NewAddressServiceTestSuite(t)
			defer suite.ctrl.Finish()

			ctx := context.WithValue(context.Background(), contextkeys.LoggerKey, zap.NewNop())
			tt.setupMock(suite)

			addresses, err := suite.addressService.ListUserAddresses(ctx, tt.userID)

			assert.True(t, errors.Is(err, tt.expectedError))

			if tt.checkFunc != nil {
				tt.checkFunc(t, addresses, err)
			}
		})
	}
}

func TestAddressService_GetUserAddress(t *testing.T) {
	testUserID := uuid.New()
	testAddressID := uuid.New()

	tests := []struct {
		name          string
		userID        string
		addressID     string
		setupMock     func(suite *AddressServiceTestSuite)
		expectedError error
		checkFunc     func(t *testing.T, address *models.Address, err error)
	}{
		{
			name:      "Get Address Success",
			userID:    testUserID.String(),
			addressID: testAddressID.String(),
			setupMock: func(s *AddressServiceTestSuite) {
				address := &models.Address{
					ID:          testAddressID,
					UserID:      testUserID,
					FullName:    "John Doe",
					AddressType: models.AddressTypeHome,
				}
				s.addressRepo.EXPECT().GetByIDAndUserID(gomock.Any(), testAddressID, testUserID).Return(address, nil)
			},
			expectedError: nil,
			checkFunc: func(t *testing.T, address *models.Address, err error) {
				assert.NotNil(t, address)
				assert.Equal(t, testAddressID, address.ID)
				assert.Equal(t, "John Doe", address.FullName)
			},
		},
		{
			name:      "Address Not Found",
			userID:    testUserID.String(),
			addressID: testAddressID.String(),
			setupMock: func(s *AddressServiceTestSuite) {
				s.addressRepo.EXPECT().GetByIDAndUserID(gomock.Any(), testAddressID, testUserID).Return(nil, nil)
			},
			expectedError: apperr.ErrAddressNotFound,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := NewAddressServiceTestSuite(t)
			defer suite.ctrl.Finish()

			ctx := context.WithValue(context.Background(), contextkeys.LoggerKey, zap.NewNop())
			tt.setupMock(suite)

			address, err := suite.addressService.GetUserAddress(ctx, tt.userID, tt.addressID)

			assert.True(t, errors.Is(err, tt.expectedError))

			if tt.checkFunc != nil {
				tt.checkFunc(t, address, err)
			}
		})
	}
}

func TestAddressService_UpdateUserAddress(t *testing.T) {
	testUserID := uuid.New()
	testAddressID := uuid.New()

	tests := []struct {
		name          string
		userID        string
		addressID     string
		req           *request.UpdateAddressRequest
		setupMock     func(suite *AddressServiceTestSuite)
		expectedError error
		checkFunc     func(t *testing.T, address *models.Address, err error)
	}{
		{
			name:      "Update Address Success",
			userID:    testUserID.String(),
			addressID: testAddressID.String(),
			req: &request.UpdateAddressRequest{
				FullName: ptrString("Updated Name"),
				Phone:    ptrString("9876543210"),
			},
			setupMock: func(s *AddressServiceTestSuite) {
				user := &models.User{ID: testUserID, Status: models.UserStatusActive}
				address := &models.Address{
					ID:       testAddressID,
					UserID:   testUserID,
					FullName: "Old Name",
					Phone:    "0123456789",
				}
				s.userRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(user, nil)
				s.addressRepo.EXPECT().GetByIDAndUserID(gomock.Any(), testAddressID, testUserID).Return(address, nil)
				s.addressRepo.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})
				s.addressRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(int64(1), nil)
			},
			expectedError: nil,
			checkFunc: func(t *testing.T, address *models.Address, err error) {
				assert.NotNil(t, address)
				assert.Equal(t, "Updated Name", address.FullName)
				assert.Equal(t, "9876543210", address.Phone)
			},
		},
		{
			name:      "Update Address With Set Default",
			userID:    testUserID.String(),
			addressID: testAddressID.String(),
			req: &request.UpdateAddressRequest{
				IsDefault: ptrBool(true),
			},
			setupMock: func(s *AddressServiceTestSuite) {
				user := &models.User{ID: testUserID, Status: models.UserStatusActive}
				address := &models.Address{
					ID:        testAddressID,
					UserID:    testUserID,
					FullName:  "John Doe",
					IsDefault: false,
				}
				s.userRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(user, nil)
				s.addressRepo.EXPECT().GetByIDAndUserID(gomock.Any(), testAddressID, testUserID).Return(address, nil)
				s.addressRepo.EXPECT().
					WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						return fn(ctx)
					})
				s.addressRepo.EXPECT().SetDefaultAddress(gomock.Any(), testUserID, testAddressID).Return(int64(1), nil)
				s.addressRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(int64(1), nil)
			},
			expectedError: nil,
			checkFunc: func(t *testing.T, address *models.Address, err error) {
				assert.NotNil(t, address)
				assert.True(t, address.IsDefault)
			},
		},
		{
			name:      "User Not Found",
			userID:    testUserID.String(),
			addressID: testAddressID.String(),
			req: &request.UpdateAddressRequest{
				FullName: ptrString("Updated Name"),
			},
			setupMock: func(s *AddressServiceTestSuite) {
				s.userRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(nil, nil)
			},
			expectedError: apperr.ErrUserNotFound,
			checkFunc:     nil,
		},
		{
			name:      "Account Inactive",
			userID:    testUserID.String(),
			addressID: testAddressID.String(),
			req: &request.UpdateAddressRequest{
				FullName: ptrString("Updated Name"),
			},
			setupMock: func(s *AddressServiceTestSuite) {
				user := &models.User{ID: testUserID, Status: models.UserStatusPending}
				s.userRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(user, nil)
			},
			expectedError: apperr.ErrAccountInactive,
			checkFunc:     nil,
		},
		{
			name:      "Address Not Found",
			userID:    testUserID.String(),
			addressID: testAddressID.String(),
			req: &request.UpdateAddressRequest{
				FullName: ptrString("Updated Name"),
			},
			setupMock: func(s *AddressServiceTestSuite) {
				user := &models.User{ID: testUserID, Status: models.UserStatusActive}
				s.userRepo.EXPECT().GetByID(gomock.Any(), testUserID).Return(user, nil)
				s.addressRepo.EXPECT().GetByIDAndUserID(gomock.Any(), testAddressID, testUserID).Return(nil, nil)
			},
			expectedError: apperr.ErrAddressNotFound,
			checkFunc:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := NewAddressServiceTestSuite(t)
			defer suite.ctrl.Finish()

			ctx := context.WithValue(context.Background(), contextkeys.LoggerKey, zap.NewNop())
			tt.setupMock(suite)

			address, err := suite.addressService.UpdateUserAddress(ctx, tt.userID, tt.addressID, tt.req)

			assert.True(t, errors.Is(err, tt.expectedError))

			if tt.checkFunc != nil {
				tt.checkFunc(t, address, err)
			}
		})
	}
}

func TestAddressService_DeleteUserAddress(t *testing.T) {
	testUserID := uuid.New()
	testAddressID := uuid.New()

	tests := []struct {
		name          string
		userID        string
		addressID     string
		setupMock     func(suite *AddressServiceTestSuite)
		expectedError error
	}{
		{
			name:      "Delete Address Success",
			userID:    testUserID.String(),
			addressID: testAddressID.String(),
			setupMock: func(s *AddressServiceTestSuite) {
				address := &models.Address{
					ID:     testAddressID,
					UserID: testUserID,
				}
				s.addressRepo.EXPECT().GetByIDAndUserID(gomock.Any(), testAddressID, testUserID).Return(address, nil)
				s.addressRepo.EXPECT().Delete(gomock.Any(), testAddressID).Return(int64(1), nil)
			},
			expectedError: nil,
		},
		{
			name:      "Address Not Found",
			userID:    testUserID.String(),
			addressID: testAddressID.String(),
			setupMock: func(s *AddressServiceTestSuite) {
				s.addressRepo.EXPECT().GetByIDAndUserID(gomock.Any(), testAddressID, testUserID).Return(nil, nil)
			},
			expectedError: apperr.ErrAddressNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := NewAddressServiceTestSuite(t)
			defer suite.ctrl.Finish()

			ctx := context.WithValue(context.Background(), contextkeys.LoggerKey, zap.NewNop())
			tt.setupMock(suite)

			err := suite.addressService.DeleteUserAddress(ctx, tt.userID, tt.addressID)

			assert.True(t, errors.Is(err, tt.expectedError))
		})
	}
}

func ptrBool(b bool) *bool {
	return &b
}
