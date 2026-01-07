package jwtprovider

import (
	"time"

	"github.com/khoihuynh300/go-microservice/user-service/internal/domain/models"
)

type JwtProvider interface {
	GenerateAccessToken(user *models.User) (string, error)
	GenerateRefreshToken(userID string) (string, error)

	VerifyAccessToken(token string) (*AccessTokenClaims, error)
	VerifyRefreshToken(token string) (*RefreshTokenClaims, error)

	GetRefreshTTL() time.Duration
}
