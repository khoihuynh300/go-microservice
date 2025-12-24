package impl

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/khoihuynh300/go-microservice/user-service/internal/db/generated"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository"
)

type refreshTokenRepository struct {
	baseRepository
}

func NewRefreshTokenRepository(db *pgxpool.Pool) repository.RefreshTokenRepository {
	return &refreshTokenRepository{
		baseRepository: baseRepository{
			db: db,
			q:  sqlc.New(db),
		},
	}
}

func (r *refreshTokenRepository) Save(ctx context.Context, refreshToken *models.RefreshToken) error {
	params := sqlc.CreateRefreshTokenParams{
		ID:        uuid.New(),
		UserID:    refreshToken.UserID,
		TokenHash: refreshToken.TokenHash,
		ExpiresAt: refreshToken.ExpiresAt,
		CreatedAt: time.Now(),
	}

	return r.queries(ctx).CreateRefreshToken(ctx, params)
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, refreshTokenStr string) (*models.RefreshToken, error) {
	row, err := r.queries(ctx).GetRefreshTokenByTokenHash(ctx, refreshTokenStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &models.RefreshToken{
		ID:        row.ID,
		UserID:    row.UserID,
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *refreshTokenRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.queries(ctx).DeleteRefreshTokenByID(ctx, id)
}
