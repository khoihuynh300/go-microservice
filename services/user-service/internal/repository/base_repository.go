package repository

import "context"

type Repository interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
