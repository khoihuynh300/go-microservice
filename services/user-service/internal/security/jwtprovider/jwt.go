package jwtprovider

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/khoihuynh300/go-microservice/user-service/internal/domain"
)

type AccessTokenClaims struct {
	UserID string
	Email  string
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	UserID string
	jwt.RegisteredClaims
}

type JwtService struct {
	access_secret  []byte
	access_ttl     time.Duration
	refresh_secret []byte
	refresh_ttl    time.Duration
}

var (
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")
)

func NewJwtService(accessSecret string, accessTTL time.Duration, refreshSecret string, refreshTTL time.Duration) *JwtService {
	return &JwtService{
		access_secret:  []byte(accessSecret),
		access_ttl:     accessTTL,
		refresh_secret: []byte(refreshSecret),
		refresh_ttl:    refreshTTL,
	}
}

func (s *JwtService) GenerateAccessToken(user *domain.User) (string, error) {
	now := time.Now()

	claims := AccessTokenClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.access_ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.access_secret)
}

func (s *JwtService) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()

	claims := RefreshTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refresh_ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.refresh_secret)
}

func (s *JwtService) VerifyAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return s.access_secret, nil
	})

	if err != nil || !token.Valid {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

func (s *JwtService) VerifyRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return s.refresh_secret, nil
	})

	if err != nil || !token.Valid {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

func (s *JwtService) GetRefreshTTL() time.Duration {
	return s.refresh_ttl
}
