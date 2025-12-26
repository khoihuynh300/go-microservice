package apperr

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

const (
	// common errror codes
	CodeValidationFailed   = "VALIDATION_FAILED"
	CodeNotFound           = "NOT_FOUND"
	CodeConflict           = "CONFLICT"
	CodeAlreadyExists      = "ALREADY_EXISTS"
	CodeInternal           = "INTERNAL_ERROR"
	CodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeTimeout            = "TIMEOUT"

	// auth error codes
	CodeUnauthenticated    = "UNAUTHENTICATED"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeTokenExpired       = "TOKEN_EXPIRED"
	CodeTokenInvalid       = "TOKEN_INVALID"

	//// business error codes
	// user
	CodeEmailAlreadyExists   = "EMAIL_ALREADY_EXISTS"
	CodeAccountInactive      = "ACCOUNT_INACTIVE"
	CodeUserNotFound         = "USER_NOT_FOUND"
	CodeEmailAlreadyVerified = "EMAIL_ALREADY_VERIFIED"
)

var (
	// common errors
	ErrNotFound      = New(CodeNotFound, "Resource not found", nil, http.StatusNotFound, codes.NotFound)
	ErrAlreadyExists = New(CodeAlreadyExists, "Resource already exists", nil, http.StatusConflict, codes.AlreadyExists)
	ErrInternal      = New(CodeInternal, "Internal server error", nil, http.StatusInternalServerError, codes.Internal)

	// auth errors
	ErrUnauthenticated    = New(CodeUnauthenticated, "Authentication required", nil, http.StatusUnauthorized, codes.Unauthenticated)
	ErrInvalidAuthHeader  = New(CodeUnauthenticated, "Invalid authorization header", nil, http.StatusUnauthorized, codes.Unauthenticated)
	ErrUnauthorized       = New(CodeUnauthorized, "Permission denied", nil, http.StatusForbidden, codes.PermissionDenied)
	ErrInvalidCredentials = New(CodeInvalidCredentials, "Incorrect login or password", nil, http.StatusUnauthorized, codes.Unauthenticated)
	ErrTokenExpired       = New(CodeTokenExpired, "Token has expired", nil, http.StatusUnauthorized, codes.Unauthenticated)
	ErrTokenInvalid       = New(CodeTokenInvalid, "Token is invalid", nil, http.StatusUnauthorized, codes.Unauthenticated)

	//// business errors
	// user
	ErrEmailAlreadyExists   = New(CodeEmailAlreadyExists, "Email already exists", nil, http.StatusConflict, codes.AlreadyExists)
	ErrAccountInactive      = New(CodeAccountInactive, "Account is inactive", nil, http.StatusUnauthorized, codes.PermissionDenied)
	ErrUserNotFound         = New(CodeUserNotFound, "User not found", nil, http.StatusNotFound, codes.NotFound)
	ErrEmailAlreadyVerified = New(CodeEmailAlreadyVerified, "Email already verified", nil, http.StatusConflict, codes.FailedPrecondition)
)

func NewErrValidationFailed(details []ErrorDetail) *AppError {
	return New(CodeValidationFailed, "Validation failed", details, http.StatusBadRequest, codes.InvalidArgument)
}
