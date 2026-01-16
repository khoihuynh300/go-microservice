package apperr

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

const (
	// common errror codes
	CodeValidationFailed   = "VALIDATION_FAILED"
	CodeNotFound           = "NOT_FOUND"
	CodeBadRequest         = "BAD_REQUEST"
	CodeConflict           = "CONFLICT"
	CodeAlreadyExists      = "ALREADY_EXISTS"
	CodeInternal           = "INTERNAL_ERROR"
	CodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeTimeout            = "TIMEOUT"

	// invalid format
	CodeInvalidDateFormat = "INVALID_DATE_FORMAT"

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

	// address
	CodeAddressNotFound = "ADDRESS_NOT_FOUND"

	// product

	CodeProductNotFound      = "PRODUCT_NOT_FOUND"
	CodeProductAlreadyExists = "PRODUCT_ALREADY_EXISTS"
	CodeProductSKUExists     = "PRODUCT_SKU_EXISTS"
	CodeProductSlugExists    = "PRODUCT_SLUG_EXISTS"
	CodeProductImageNotFound = "PRODUCT_IMAGE_NOT_FOUND"

	// category
	CodeCategoryNotFound          = "CATEGORY_NOT_FOUND"
	CodeParentCategoryNotFound    = "PARENT_CATEGORY_NOT_FOUND"
	CodeCategoryCannotBeOwnParent = "CATEGORY_CANNOT_BE_OWN_PARENT"
	CodeCategoryAlreadyExists     = "CATEGORY_ALREADY_EXISTS"
	CodeCategorySlugExists        = "CATEGORY_SLUG_EXISTS"
	CodeCategoryHasProducts       = "CATEGORY_HAS_PRODUCTS"
	CodeCategoryHasChildren       = "CATEGORY_HAS_CHILDREN"
)

var (
	// common errors
	ErrNotFound      = New(CodeNotFound, "Resource not found", nil, http.StatusNotFound, codes.NotFound)
	ErrAlreadyExists = New(CodeAlreadyExists, "Resource already exists", nil, http.StatusConflict, codes.AlreadyExists)
	ErrInternal      = New(CodeInternal, "Internal server error", nil, http.StatusInternalServerError, codes.Internal)

	// auth errors
	ErrUnauthenticated        = New(CodeUnauthenticated, "Authentication required", nil, http.StatusUnauthorized, codes.Unauthenticated)
	ErrInvalidAuthHeader      = New(CodeUnauthenticated, "Invalid authorization header", nil, http.StatusUnauthorized, codes.Unauthenticated)
	ErrUnauthorized           = New(CodeUnauthorized, "Permission denied", nil, http.StatusForbidden, codes.PermissionDenied)
	ErrInvalidCredentials     = New(CodeInvalidCredentials, "Incorrect login or password", nil, http.StatusUnauthorized, codes.Unauthenticated)
	ErrInvalidCurrentPassword = New(CodeInvalidCredentials, "Current password is incorrect", nil, http.StatusUnauthorized, codes.Unauthenticated)
	ErrTokenExpired           = New(CodeTokenExpired, "Token has expired", nil, http.StatusUnauthorized, codes.Unauthenticated)
	ErrTokenInvalid           = New(CodeTokenInvalid, "Token is invalid", nil, http.StatusUnauthorized, codes.Unauthenticated)
	ErrTokenInvalidOrExpired  = New(CodeTokenInvalid, "Token is invalid or expired", nil, http.StatusUnauthorized, codes.Unauthenticated)

	//// business errors
	// user
	ErrEmailAlreadyExists   = New(CodeEmailAlreadyExists, "Email already exists", nil, http.StatusConflict, codes.AlreadyExists)
	ErrAccountInactive      = New(CodeAccountInactive, "Account is inactive", nil, http.StatusUnauthorized, codes.PermissionDenied)
	ErrUserNotFound         = New(CodeUserNotFound, "User not found", nil, http.StatusNotFound, codes.NotFound)
	ErrEmailAlreadyVerified = New(CodeEmailAlreadyVerified, "Email already verified", nil, http.StatusConflict, codes.FailedPrecondition)

	// address
	ErrAddressNotFound = New(CodeAddressNotFound, "Address not found", nil, http.StatusNotFound, codes.NotFound)

	// product
	ErrProductNotFound      = New(CodeProductNotFound, "Product not found", nil, http.StatusNotFound, codes.NotFound)
	ErrProductAlreadyExists = New(CodeProductAlreadyExists, "Product already exists", nil, http.StatusConflict, codes.AlreadyExists)
	ErrProductSKUExists     = New(CodeProductSKUExists, "Product with the given SKU already exists", nil, http.StatusConflict, codes.AlreadyExists)
	ErrProductSlugExists    = New(CodeProductSlugExists, "Product with the given slug already exists", nil, http.StatusConflict, codes.AlreadyExists)
	ErrProductImageNotFound = New(CodeProductImageNotFound, "Product image not found", nil, http.StatusNotFound, codes.NotFound)

	// category
	ErrCategoryNotFound          = New(CodeCategoryNotFound, "Category not found", nil, http.StatusNotFound, codes.NotFound)
	ErrParentCategoryNotFound    = New(CodeParentCategoryNotFound, "Parent category not found", nil, http.StatusNotFound, codes.NotFound)
	ErrCategoryCannotBeOwnParent = New(CodeCategoryCannotBeOwnParent, "Category cannot be its own parent", nil, http.StatusBadRequest, codes.InvalidArgument)
	ErrCategoryAlreadyExists     = New(CodeCategoryAlreadyExists, "Category already exists", nil, http.StatusConflict, codes.AlreadyExists)
	ErrCategorySlugExists        = New(CodeCategorySlugExists, "Category with the given slug already exists", nil, http.StatusConflict, codes.AlreadyExists)
	ErrCategoryHasProducts       = New(CodeCategoryHasProducts, "Category has associated products and cannot be deleted", nil, http.StatusConflict, codes.FailedPrecondition)
	ErrCategoryHasChildren       = New(CodeCategoryHasChildren, "Category has child categories and cannot be deleted", nil, http.StatusConflict, codes.FailedPrecondition)
)

func NewErrValidationFailed(details []ErrorDetail) *AppError {
	return New(CodeValidationFailed, "Validation failed", details, http.StatusBadRequest, codes.InvalidArgument)
}

func NewErrValidationFailedWithDetail(field, code, message string) *AppError {
	return NewErrValidationFailed([]ErrorDetail{
		{
			Field:   field,
			Code:    code,
			Message: message,
		},
	})
}
