package cache

import (
	"context"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	Get(ctx context.Context, key string) (string, error)

	GetObject(ctx context.Context, key string, dest any) error

	Delete(ctx context.Context, keys ...string) error

	Exists(ctx context.Context, key string) (bool, error)

	SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error)

	TTL(ctx context.Context, key string) (time.Duration, error)

	Expire(ctx context.Context, key string, ttl time.Duration) error

	Keys(ctx context.Context, pattern string) ([]string, error)

	Ping(ctx context.Context) error

	Close() error
}
