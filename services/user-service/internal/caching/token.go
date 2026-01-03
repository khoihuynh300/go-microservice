package caching

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/shared/pkg/cache"
	"github.com/khoihuynh300/go-microservice/user-service/internal/utils"
)

const (
	EmailVerifyPrefix   = "user:verify_email"
	PasswordResetPrefix = "user:reset_password"
	EmailChangePrefix   = "user:change_email"
)

const (
	EmailVerifyTTL   = 15 * time.Minute
	PasswordResetTTL = 30 * time.Minute
	EmailChangeTTL   = 15 * time.Minute
)

var (
	ErrTokenInvalidOrExpired = errors.New("token invalid or expired")
)

type TokenCache struct {
	cache cache.Cache
}

func NewTokenCache(cache cache.Cache) *TokenCache {
	return &TokenCache{
		cache: cache,
	}
}

func (tc *TokenCache) SetEmailVerifyToken(ctx context.Context, email string) (string, error) {
	tokenStr := uuid.New().String()
	tokenHash := utils.HashToken(tokenStr)

	key := fmt.Sprintf("%s:%s", EmailVerifyPrefix, tokenHash)

	err := tc.cache.Set(ctx, key, email, EmailVerifyTTL)
	if err != nil {
		return "", fmt.Errorf("failed to set email verify token: %w", err)
	}

	return tokenStr, nil
}

func (tc *TokenCache) VerifyEmailToken(ctx context.Context, tokenStr string) (string, error) {
	tokenHash := utils.HashToken(tokenStr)
	key := fmt.Sprintf("%s:%s", EmailVerifyPrefix, tokenHash)

	email, err := tc.cache.Get(ctx, key)
	if err != nil {
		return "", ErrTokenInvalidOrExpired
	}

	_ = tc.cache.Delete(ctx, key)

	return email, nil
}

func (tc *TokenCache) SetPasswordResetToken(ctx context.Context, email string) (string, error) {
	tokenStr := uuid.New().String()
	tokenHash := utils.HashToken(tokenStr)
	key := fmt.Sprintf("%s:%s", PasswordResetPrefix, tokenHash)

	err := tc.cache.Set(ctx, key, email, PasswordResetTTL)
	if err != nil {
		return "", fmt.Errorf("failed to set password reset token: %w", err)
	}

	return tokenStr, nil
}

func (tc *TokenCache) VerifyPasswordResetToken(ctx context.Context, tokenStr string) (string, error) {
	tokenHash := utils.HashToken(tokenStr)
	key := fmt.Sprintf("%s:%s", PasswordResetPrefix, tokenHash)

	email, err := tc.cache.Get(ctx, key)
	if err != nil {
		return "", ErrTokenInvalidOrExpired
	}

	_ = tc.cache.Delete(ctx, key)

	return email, nil
}
