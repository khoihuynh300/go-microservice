package models

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusPending   UserStatus = "pending"
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type User struct {
	ID              uuid.UUID
	Email           string
	HashedPassword  string
	FullName        string
	Phone           *string
	AvatarURL       *string
	DateOfBirth     *time.Time
	Gender          *Gender
	Status          UserStatus
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}
