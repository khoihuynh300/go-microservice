package repository

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
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain"
)

type userRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
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

func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return r.mapToUser(row), nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}
	return r.mapToUser(row), nil
}

func (r *userRepository) List(ctx context.Context, status domain.UserStatus, limit, offset int) ([]*domain.User, error) {
	params := sqlc.ListUsersParams{
		Status: string(status),
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	rows, err := r.queries.ListUsers(ctx, params)
	if err != nil {
		return nil, err
	}

	users := make([]*domain.User, 0, len(rows))
	for i, row := range rows {
		users[i] = r.mapToUser(row)
	}
	return users, nil
}

func (r *userRepository) Count(ctx context.Context, status domain.UserStatus) (int64, error) {
	count, err := r.queries.CountUsers(ctx, string(status))
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	params := sqlc.UpdateUserParams{
		ID:          user.ID,
		FullName:    user.FullName,
		Phone:       pgtype.Text{String: user.Phone, Valid: user.Phone != ""},
		AvatarUrl:   pgtype.Text{String: user.AvatarURL, Valid: user.AvatarURL != ""},
		DateOfBirth: pgtype.Date{Time: *user.DateOfBirth, Valid: user.DateOfBirth != nil},
		Gender:      pgtype.Text{String: string(user.Gender), Valid: user.Gender != ""},
		UpdatedAt:   time.Now(),
	}

	result, err := r.queries.UpdateUser(ctx, params)
	if err != nil {
		return err
	}
	user.UpdatedAt = result.UpdatedAt
	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, id, hashedPassword string) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	params := sqlc.UpdateUserPasswordParams{
		ID:             userID,
		HashedPassword: hashedPassword,
		UpdatedAt:      time.Now(),
	}

	return r.queries.UpdateUserPassword(ctx, params)
}

func (r *userRepository) UpdateStatus(ctx context.Context, id string, status domain.UserStatus) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	params := sqlc.UpdateUserStatusParams{
		ID:        userID,
		Status:    string(status),
		UpdatedAt: time.Now(),
	}
	return r.queries.UpdateUserStatus(ctx, params)
}

func (r *userRepository) SoftDelete(ctx context.Context, id string) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	params := sqlc.SoftDeleteUserParams{
		ID:        userID,
		DeletedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	return r.queries.SoftDeleteUser(ctx, params)
}

func (r *userRepository) mapToUser(row sqlc.User) *domain.User {
	user := &domain.User{
		ID:             row.ID,
		Email:          row.Email,
		HashedPassword: row.HashedPassword,
		FullName:       row.FullName,
		Phone:          row.Phone.String,
		AvatarURL:      row.AvatarUrl.String,
		DateOfBirth:    &row.DateOfBirth.Time,
		Gender:         row.Gender.String,
		Status:         domain.UserStatus(row.Status),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}

	if row.DateOfBirth.Valid {
		user.DateOfBirth = &row.DateOfBirth.Time
	}

	return user
}
