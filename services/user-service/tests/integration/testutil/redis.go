package testutil

import (
	"context"
	"fmt"
	"strconv"

	"github.com/khoihuynh300/go-microservice/shared/pkg/cache"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

type TestRedis struct {
	Container *redis.RedisContainer
	Client    cache.Cache
	Host      string
	Port      int
}

func StartRedisContainer(ctx context.Context) (*TestRedis, error) {
	redisContainer, err := redis.Run(ctx, "redis:7-alpine")
	if err != nil {
		return nil, fmt.Errorf("failed to start redis container: %w", err)
	}

	host, err := redisContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get redis host: %w", err)
	}

	mappedPort, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		return nil, fmt.Errorf("failed to get redis port: %w", err)
	}

	port, err := strconv.Atoi(mappedPort.Port())
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis port: %w", err)
	}

	cacheClient, err := cache.NewClient(&cache.Config{
		Host:     host,
		Port:     port,
		Password: "",
		DB:       0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	return &TestRedis{
		Container: redisContainer,
		Client:    cacheClient,
		Host:      host,
		Port:      port,
	}, nil
}

func (r *TestRedis) TearDown(ctx context.Context) {
	if r.Client != nil {
		r.Client.Close()
	}
	if r.Container != nil {
		r.Container.Terminate(ctx)
	}
}

func (r *TestRedis) FlushAll(ctx context.Context) error {
	if client, ok := r.Client.(*cache.Client); ok {
		return client.GetClient().FlushAll(ctx).Err()
	}
	return nil
}
