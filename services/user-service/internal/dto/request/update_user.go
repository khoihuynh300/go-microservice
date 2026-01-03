package request

import "time"

type UpdateUserRequest struct {
	FullName    *string
	DateOfBirth *time.Time
	Gender      *string
}
