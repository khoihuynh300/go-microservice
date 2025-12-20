package impl

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/khoihuynh300/go-microservice/user-service/db/generated"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository"
)

type registryTokenRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

var _ repository.RegistryTokenRepository = (*registryTokenRepository)(nil)

func NewRegistryTokenRepository(db *pgxpool.Pool) *registryTokenRepository {
	return &registryTokenRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *registryTokenRepository) Create(ctx context.Context, token_hash string, userID uuid.UUID, expiresAt time.Time) error {
	params := sqlc.CreateRegistryTokenParams{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: token_hash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	return r.queries.CreateRegistryToken(ctx, params)
}

func (r *registryTokenRepository) GetUserIdByToken(ctx context.Context, token string) (*models.RegistryToken, error) {
	row, err := r.queries.GetActiveRegistryToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var usedAt *time.Time
	if row.UsedAt.Valid {
		usedAt = &row.UsedAt.Time
	}

	var invalidatedAt *time.Time
	if row.InvalidatedAt.Valid {
		invalidatedAt = &row.InvalidatedAt.Time
	}

	return &models.RegistryToken{
		ID:            row.ID,
		UserID:        row.UserID,
		TokenHash:     row.TokenHash,
		CreatedAt:     row.CreatedAt,
		ExpiresAt:     row.ExpiresAt,
		UsedAt:        usedAt,
		InvalidatedAt: invalidatedAt,
	}, nil
}

func (r *registryTokenRepository) InvalidateToken(ctx context.Context, token_hash string) error {
	return r.queries.InvalidateRegistryTokens(ctx, uuid.Nil)
}

func (r *registryTokenRepository) MarkTokenAsUsed(ctx context.Context, token_hash string) error {
	return r.queries.MarkRegistryTokenAsUsed(ctx, token_hash)
}
