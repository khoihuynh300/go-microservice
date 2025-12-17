package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/khoihuynh300/go-microservice/user-service/db/generated"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain"
)

type refreshTokenRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRefreshTokenRepository(db *pgxpool.Pool) RefreshTokenRepository {
	return &refreshTokenRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *refreshTokenRepository) Save(ctx context.Context, refreshToken *domain.RefreshToken) error {
	params := sqlc.CreateRefreshTokenParams{
		ID:        uuid.New(),
		UserID:    refreshToken.UserID,
		TokenHash: refreshToken.TokenHash,
		ExpiresAt: refreshToken.ExpiresAt,
		CreatedAt: time.Now(),
	}

	return r.queries.CreateRefreshToken(ctx, params)
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, refreshTokenStr string) (*domain.RefreshToken, error) {
	row, err := r.queries.GetRefreshTokenByTokenHash(ctx, refreshTokenStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.RefreshToken{
		ID:        row.ID,
		UserID:    row.UserID,
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *refreshTokenRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteRefreshTokenByID(ctx, id)
}
