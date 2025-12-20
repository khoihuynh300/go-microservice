package models

import (
	"time"

	"github.com/google/uuid"
)

type RegistryToken struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	TokenHash     string
	CreatedAt     time.Time
	ExpiresAt     time.Time
	UsedAt        *time.Time
	InvalidatedAt *time.Time
}

func (rt *RegistryToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsUsed() && !rt.IsInvalidated()
}

func (rt *RegistryToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

func (rt *RegistryToken) IsUsed() bool {
	return rt.UsedAt != nil
}

func (rt *RegistryToken) IsInvalidated() bool {
	return rt.InvalidatedAt != nil
}
