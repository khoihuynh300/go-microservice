package apperr

import (
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"
)

func ToGRPC(err error) error {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(*AppError); ok {
		return appErr.ToGRPCStatus().Err()
	}

	if _, ok := status.FromError(err); ok {
		return err
	}

	return status.Error(codes.Internal, "internal server error")
}

func (e *AppError) ToGRPCStatus() *status.Status {
	st := status.New(e.GRPCCode, e.Message)

	var detail protoadapt.MessageV1

	if len(e.Details) > 0 {
		switch e.GRPCCode {
		case codes.InvalidArgument:
			detail = buildBadRequestDetails(e.Details)
		default:
			detail = buildErrorInfoDetails(e.Code, e.Details)
		}

	} else {
		detail = buildErrorInfoDetails(e.Code, nil)
	}

	st, _ = st.WithDetails(detail)

	return st
}

func buildBadRequestDetails(details []ErrorDetail) *errdetails.BadRequest {
	br := &errdetails.BadRequest{
		FieldViolations: make([]*errdetails.BadRequest_FieldViolation, 0, len(details)),
	}

	for _, detail := range details {
		br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       detail.Field,
			Reason:      detail.Code,
			Description: detail.Message,
		})
	}

	return br
}

func buildErrorInfoDetails(code string, details []ErrorDetail) *errdetails.ErrorInfo {
	metadata := make(map[string]string, len(details))

	for _, detail := range details {
		metadata[detail.Field] = fmt.Sprint(detail.Message)
	}

	return &errdetails.ErrorInfo{
		Reason:   code,
		Metadata: metadata,
	}
}

func FromGRPCError(err error) *AppError {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return ErrInternal.Wrap(err)
	}

	errorCode := grpcCodeToErrorCode(st.Code())
	message := st.Message()
	details := []ErrorDetail{}
	httpStatus := runtime.HTTPStatusFromCode(st.Code())

	for _, detail := range st.Details() {
		switch d := detail.(type) {
		case *errdetails.ErrorInfo:
			if d.Reason != "" {
				errorCode = d.Reason
			}

		case *errdetails.BadRequest:
			errorCode = CodeValidationFailed
			for _, violation := range d.GetFieldViolations() {
				details = append(details, ErrorDetail{
					Field:   violation.GetField(),
					Code:    violation.GetReason(),
					Message: violation.GetDescription(),
				})
			}
		}
	}

	return New(errorCode, message, details, httpStatus, st.Code())
}

func grpcCodeToErrorCode(code codes.Code) string {
	switch code {
	case codes.InvalidArgument:
		return CodeValidationFailed
	case codes.NotFound:
		return CodeNotFound
	case codes.AlreadyExists:
		return CodeAlreadyExists
	case codes.PermissionDenied:
		return CodeUnauthorized
	case codes.Unauthenticated:
		return CodeUnauthenticated
	case codes.ResourceExhausted:
		return CodeRateLimitExceeded
	case codes.FailedPrecondition:
		return CodeConflict
	case codes.Unavailable:
		return CodeServiceUnavailable
	case codes.DeadlineExceeded:
		return CodeTimeout
	case codes.Internal:
		return CodeInternal
	default:
		return CodeInternal
	}
}
