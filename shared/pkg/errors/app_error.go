package apperr

import (
	"fmt"

	"google.golang.org/grpc/codes"
)

type AppError struct {
	Code       string        `json:"code"`
	Message    string        `json:"message"`
	Details    []ErrorDetail `json:"details,omitempty"`
	HTTPStatus int           `json:"-"`
	GRPCCode   codes.Code    `json:"-"`
	Err        error         `json:"-"`
}

type ErrorDetail struct {
	Field   string `json:"field"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

func New(code, message string, details []ErrorDetail, httpStatus int, grpcCode codes.Code) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: httpStatus,
		GRPCCode:   grpcCode,
	}
}

func Newf(code string, details []ErrorDetail, httpStatus int, grpcCode codes.Code, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		Details:    details,
		HTTPStatus: httpStatus,
		GRPCCode:   grpcCode,
	}
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Wrap(err error) *AppError {
	e.Err = err
	return e
}
