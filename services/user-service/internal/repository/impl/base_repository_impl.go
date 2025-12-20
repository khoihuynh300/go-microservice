package impl

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/khoihuynh300/go-microservice/user-service/db/generated"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository"
)

type baseRepository struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewRepository(db *pgxpool.Pool) repository.Repository {
	return &baseRepository{
		db: db,
		q:  sqlc.New(db),
	}
}

func (r *baseRepository) queries(ctx context.Context) *sqlc.Queries {
	if tx := extractTx(ctx); tx != nil {
		return r.q.WithTx(tx)
	}
	return r.q
}

func (r *baseRepository) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// if there's already a transaction, use it
	if extractTx(ctx) != nil {
		return fn(ctx)
	}

	tx, err := r.db.Begin(ctx)

	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	txCtx := injectTx(ctx, tx)

	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
