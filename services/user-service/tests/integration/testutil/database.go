package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDatabase struct {
	Container testcontainers.Container
	Pool      *pgxpool.Pool
	ConnStr   string
}

func StartPostgresContainer(ctx context.Context) (*TestDatabase, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := runMigrations(ctx, pool); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &TestDatabase{
		Container: pgContainer,
		Pool:      pool,
		ConnStr:   connStr,
	}, nil
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrationsDir := filepath.Join("..", "..", "..", "migrations")

	files := []string{
		"000001_create_users_table.up.sql",
		"000002_create_user_addresses_table.up.sql",
		"000003_create_refresh_tokens_table.up.sql",
	}

	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}

		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("execute migration %s: %w", file, err)
		}
	}

	return nil
}

func (td *TestDatabase) TearDown(ctx context.Context) {
	if td.Pool != nil {
		td.Pool.Close()
	}
	if td.Container != nil {
		td.Container.Terminate(ctx)
	}
}

func (td *TestDatabase) CleanupTestData(ctx context.Context) error {
	_, err := td.Pool.Exec(ctx, `
        TRUNCATE TABLE refresh_tokens, user_addresses, users RESTART IDENTITY CASCADE
    `)
	return err
}
