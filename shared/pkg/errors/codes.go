package apperr

import "google.golang.org/grpc/codes"

const (
	CodeUnauthenticated    = "UNAUTHENTICATED"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeTokenExpired       = "TOKEN_EXPIRED"
	CodeTokenInvalid       = "TOKEN_INVALID"
	CodeValidationFailed   = "VALIDATION_FAILED"
	CodeNotFound           = "NOT_FOUND"
	CodeConflict           = "CONFLICT"
	CodeAlreadyExists      = "ALREADY_EXISTS"
	CodeInternal           = "INTERNAL_ERROR"
)

var (
	ErrUnauthenticated    = New(CodeUnauthenticated, "Authentication required", codes.Unauthenticated)
	ErrUnauthorized       = New(CodeUnauthorized, "Permission denied", codes.PermissionDenied)
	ErrInvalidCredentials = New(CodeInvalidCredentials, "Incorrect login or password", codes.Unauthenticated)
	ErrTokenExpired       = New(CodeTokenExpired, "Token has expired", codes.Unauthenticated)
	ErrTokenInvalid       = New(CodeTokenInvalid, "Token is invalid", codes.Unauthenticated)
	ErrValidationFailed   = New(CodeValidationFailed, "Validation failed", codes.InvalidArgument)
	ErrNotFound           = New(CodeNotFound, "Resource not found", codes.NotFound)
	ErrAlreadyExists      = New(CodeAlreadyExists, "Resource already exists", codes.AlreadyExists)
	ErrInternal           = New(CodeInternal, "Internal server error", codes.Internal)
)
