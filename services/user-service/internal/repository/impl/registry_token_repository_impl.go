package impl

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/khoihuynh300/go-microservice/user-service/internal/db/convert"
	sqlc "github.com/khoihuynh300/go-microservice/user-service/internal/db/generated"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository"
)

type registryTokenRepository struct {
	baseRepository
}

func NewRegistryTokenRepository(db *pgxpool.Pool) repository.RegistryTokenRepository {
	return &registryTokenRepository{
		baseRepository: baseRepository{
			db: db,
			q:  sqlc.New(db),
		},
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

	return r.queries(ctx).CreateRegistryToken(ctx, params)
}

func (r *registryTokenRepository) GetByToken(ctx context.Context, token string) (*models.RegistryToken, error) {
	row, err := r.queries(ctx).GetActiveRegistryToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &models.RegistryToken{
		ID:            row.ID,
		UserID:        row.UserID,
		TokenHash:     row.TokenHash,
		CreatedAt:     row.CreatedAt,
		ExpiresAt:     row.ExpiresAt,
		UsedAt:        convert.PtrIfValid(row.UsedAt.Time, row.UsedAt.Valid),
		InvalidatedAt: convert.PtrIfValid(row.InvalidatedAt.Time, row.InvalidatedAt.Valid),
	}, nil
}

func (r *registryTokenRepository) InvalidateAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	return r.queries(ctx).InvalidateRegistryTokens(ctx, userID)
}

func (r *registryTokenRepository) MarkTokenAsUsed(ctx context.Context, token_hash string) error {
	return r.queries(ctx).MarkRegistryTokenAsUsed(ctx, token_hash)
}
