package domainerr

import (
	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrEmailAlreadyExists   = apperr.New(apperr.CodeAlreadyExists, "Email already exists", codes.AlreadyExists)
	ErrAccountInactive      = apperr.New(apperr.CodeUnauthorized, "Account is inactive", codes.PermissionDenied)
	ErrUserNotFound         = apperr.New(apperr.CodeNotFound, "User not found", codes.NotFound)
	ErrEmailAlreadyVerified = apperr.New(apperr.CodeConflict, "Email already verified", codes.FailedPrecondition)
)
