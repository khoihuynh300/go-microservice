package impl

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/khoihuynh300/go-microservice/user-service/internal/db/generated"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository"
	"github.com/khoihuynh300/go-microservice/user-service/internal/utils/convert"
)

type userRepository struct {
	baseRepository
}

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepository{
		baseRepository: baseRepository{
			db: db,
			q:  sqlc.New(db),
		},
	}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	params := sqlc.CreateUserParams{
		ID:             uuid.New(),
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
		FullName:       user.FullName,
		Phone:          convert.PtrToText(user.Phone),
		Status:         sqlc.UserStatusEnum(user.Status),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	result, err := r.queries(ctx).CreateUser(ctx, params)
	if err != nil {
		return err
	}

	user.ID = result.ID
	user.CreatedAt = result.CreatedAt
	user.UpdatedAt = result.UpdatedAt

	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	row, err := r.queries(ctx).GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return r.mapToUser(row), nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	row, err := r.queries(ctx).GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}
	return r.mapToUser(row), nil
}

func (r *userRepository) List(ctx context.Context, status models.UserStatus, limit, offset int) ([]*models.User, error) {
	params := sqlc.ListUsersParams{
		Status: sqlc.UserStatusEnum(status),
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	rows, err := r.queries(ctx).ListUsers(ctx, params)
	if err != nil {
		return nil, err
	}

	users := make([]*models.User, 0, len(rows))
	for i, row := range rows {
		users[i] = r.mapToUser(row)
	}
	return users, nil
}

func (r *userRepository) Count(ctx context.Context, status models.UserStatus) (int64, error) {
	count, err := r.queries(ctx).CountUsers(ctx, sqlc.UserStatusEnum(status))
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	params := sqlc.UpdateUserParams{
		ID:              user.ID,
		FullName:        user.FullName,
		Phone:           convert.PtrToText(user.Phone),
		AvatarUrl:       convert.PtrToText(user.AvatarURL),
		DateOfBirth:     convert.PtrToDate(user.DateOfBirth),
		Gender:          convert.PtrToGenderEnum(user.Gender),
		UpdatedAt:       time.Now(),
		Status:          sqlc.UserStatusEnum(user.Status),
		EmailVerifiedAt: convert.PtrToTimestamptz(user.EmailVerifiedAt),
	}

	result, err := r.queries(ctx).UpdateUser(ctx, params)
	if err != nil {
		return err
	}
	user.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	params := sqlc.UpdateUserPasswordParams{
		ID:             id,
		HashedPassword: hashedPassword,
		UpdatedAt:      time.Now(),
	}

	return r.queries(ctx).UpdateUserPassword(ctx, params)
}

func (r *userRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.UserStatus) error {
	params := sqlc.UpdateUserStatusParams{
		ID:        id,
		Status:    sqlc.UserStatusEnum(status),
		UpdatedAt: time.Now(),
	}
	return r.queries(ctx).UpdateUserStatus(ctx, params)
}

func (r *userRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	params := sqlc.SoftDeleteUserParams{
		ID:        id,
		DeletedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	return r.queries(ctx).SoftDeleteUser(ctx, params)
}

func (r *userRepository) mapToUser(row sqlc.User) *models.User {
	user := &models.User{
		ID:              row.ID,
		Email:           row.Email,
		HashedPassword:  row.HashedPassword,
		FullName:        row.FullName,
		Phone:           convert.PtrIfValid(row.Phone.String, row.Phone.Valid),
		AvatarURL:       convert.PtrIfValid(row.AvatarUrl.String, row.AvatarUrl.Valid),
		Gender:          convert.PtrIfValid(models.Gender(row.Gender.UserGenderEnum), row.Gender.Valid),
		DateOfBirth:     convert.PtrIfValid(row.DateOfBirth.Time, row.DateOfBirth.Valid),
		EmailVerifiedAt: convert.PtrIfValid(row.EmailVerifiedAt.Time, row.EmailVerifiedAt.Valid),
		Status:          models.UserStatus(row.Status),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}

	return user
}
