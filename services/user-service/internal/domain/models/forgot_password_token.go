package models

import (
	"time"

	"github.com/google/uuid"
)

type ForgotPasswordToken struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	TokenHash     string
	CreatedAt     time.Time
	ExpiresAt     time.Time
	UsedAt        *time.Time
	InvalidatedAt *time.Time
}

func (token *ForgotPasswordToken) IsValid() bool {
	return !token.IsExpired() && !token.IsUsed() && !token.IsInvalidated()
}

func (token *ForgotPasswordToken) IsExpired() bool {
	return time.Now().After(token.ExpiresAt)
}

func (token *ForgotPasswordToken) IsUsed() bool {
	return token.UsedAt != nil
}

func (token *ForgotPasswordToken) IsInvalidated() bool {
	return token.InvalidatedAt != nil
}
