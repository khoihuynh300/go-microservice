package apperr

import (
	"fmt"

	"google.golang.org/grpc/codes"
)

type AppError struct {
	Code          string
	Message       string
	GRPCCode      codes.Code
	Details       map[string]any
	OriginalError error
}

func New(code string, message string, grpcCode codes.Code) *AppError {
	return &AppError{
		Code:     code,
		Message:  message,
		GRPCCode: grpcCode,
		Details:  make(map[string]any),
	}
}

func (e *AppError) Error() string {
	if e.OriginalError != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.OriginalError)
	}
	return e.Message
}

func (e *AppError) WithDetail(key string, value any) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

func (e *AppError) WithOriginalError(err error) *AppError {
	e.OriginalError = err
	return e
}
