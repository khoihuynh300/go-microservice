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
	now := time.Now()

	params := sqlc.CreateUserParams{
		ID:             uuid.New(),
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
		FullName:       user.FullName,
		Status:         sqlc.UserStatusEnum(user.Status),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	result, err := r.queries(ctx).CreateUser(ctx, params)
	if err != nil {
		return err
	}

	user.ID = result.ID

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	row, err := r.queries(ctx).GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return r.mapToUser(row), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	row, err := r.queries(ctx).GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}
	return r.mapToUser(row), nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) (int64, error) {
	params := sqlc.UpdateUserParams{
		ID:          user.ID,
		FullName:    user.FullName,
		Phone:       convert.PtrToText(user.Phone),
		DateOfBirth: convert.PtrToDate(user.DateOfBirth),
		Gender:      convert.PtrToGenderEnum(user.Gender),
		UpdatedAt:   time.Now(),
	}

	return r.queries(ctx).UpdateUser(ctx, params)
}

func (r *userRepository) UpdateAvatar(ctx context.Context, id uuid.UUID, avatarURL string) (int64, error) {
	params := sqlc.UpdateUserAvatarParams{
		ID:        id,
		AvatarUrl: pgtype.Text{String: avatarURL, Valid: true},
		UpdatedAt: time.Now(),
	}

	return r.queries(ctx).UpdateUserAvatar(ctx, params)
}

func (r *userRepository) VerifyEmail(ctx context.Context, id uuid.UUID) (int64, error) {
	now := time.Now()
	params := sqlc.VerifyUserEmailParams{
		ID:              id,
		EmailVerifiedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:       now,
	}

	return r.queries(ctx).VerifyUserEmail(ctx, params)
}

func (r *userRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) (int64, error) {
	params := sqlc.UpdateUserPasswordParams{
		ID:             id,
		HashedPassword: hashedPassword,
		UpdatedAt:      time.Now(),
	}

	return r.queries(ctx).UpdateUserPassword(ctx, params)
}

func (r *userRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.UserStatus) (int64, error) {
	params := sqlc.UpdateUserStatusParams{
		ID:        id,
		Status:    sqlc.UserStatusEnum(status),
		UpdatedAt: time.Now(),
	}

	return r.queries(ctx).UpdateUserStatus(ctx, params)
}

func (r *userRepository) SoftDelete(ctx context.Context, id uuid.UUID) (int64, error) {
	now := time.Now()

	params := sqlc.SoftDeleteUserParams{
		ID:        id,
		DeletedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: now,
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
