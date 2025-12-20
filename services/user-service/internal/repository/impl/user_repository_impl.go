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
	sqlc "github.com/khoihuynh300/go-microservice/user-service/db/generated"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository"
)

type userRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

var _ repository.UserRepository = (*userRepository)(nil)

func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	params := sqlc.CreateUserParams{
		ID:             uuid.New(),
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
		FullName:       user.FullName,
		Phone:          pgtype.Text{String: user.Phone, Valid: user.Phone != ""},
		Status:         string(user.Status),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	result, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return err
	}

	user.ID = result.ID
	user.CreatedAt = result.CreatedAt
	user.UpdatedAt = result.UpdatedAt

	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return r.mapToUser(row), nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
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
		Status: string(status),
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	rows, err := r.queries.ListUsers(ctx, params)
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
	count, err := r.queries.CountUsers(ctx, string(status))
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	params := sqlc.UpdateUserParams{
		ID:        user.ID,
		FullName:  user.FullName,
		Phone:     pgtype.Text{String: user.Phone, Valid: user.Phone != ""},
		AvatarUrl: pgtype.Text{String: user.AvatarURL, Valid: user.AvatarURL != ""},
		// DateOfBirth: pgtype.Date{Time: *user.DateOfBirth, Valid: user.DateOfBirth != nil},
		Gender:    pgtype.Text{String: string(user.Gender), Valid: user.Gender != ""},
		UpdatedAt: time.Now(),
		Status:    string(user.Status),
		EmailVerifiedAt: pgtype.Timestamptz{
			Time:  *user.EmailVerifiedAt,
			Valid: user.EmailVerifiedAt != nil,
		},
	}

	result, err := r.queries.UpdateUser(ctx, params)
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

	return r.queries.UpdateUserPassword(ctx, params)
}

func (r *userRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.UserStatus) error {
	params := sqlc.UpdateUserStatusParams{
		ID:        id,
		Status:    string(status),
		UpdatedAt: time.Now(),
	}
	return r.queries.UpdateUserStatus(ctx, params)
}

func (r *userRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	params := sqlc.SoftDeleteUserParams{
		ID:        id,
		DeletedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	return r.queries.SoftDeleteUser(ctx, params)
}

func (r *userRepository) mapToUser(row sqlc.User) *models.User {
	user := &models.User{
		ID:             row.ID,
		Email:          row.Email,
		HashedPassword: row.HashedPassword,
		FullName:       row.FullName,
		Phone:          row.Phone.String,
		AvatarURL:      row.AvatarUrl.String,
		Gender:         row.Gender.String,
		Status:         models.UserStatus(row.Status),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}

	if row.DateOfBirth.Valid {
		user.DateOfBirth = &row.DateOfBirth.Time
	}

	if row.EmailVerifiedAt.Valid {
		user.EmailVerifiedAt = &row.EmailVerifiedAt.Time
	}

	return user
}
