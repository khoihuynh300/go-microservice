package apperr

import (
	"fmt"

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

	if len(e.Details) == 0 {
		return st
	}

	var detail protoadapt.MessageV1

	switch e.GRPCCode {
	case codes.InvalidArgument:
		detail = buildBadRequestDetails(e.Details)
	default:
		detail = buildErrorInfoDetails(e.Code, e.Details)
	}

	st, _ = st.WithDetails(detail)

	return st
}

func buildBadRequestDetails(details map[string]any) *errdetails.BadRequest {
	br := &errdetails.BadRequest{
		FieldViolations: make([]*errdetails.BadRequest_FieldViolation, 0, len(details)),
	}

	for field, value := range details {
		br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       field,
			Description: fmt.Sprint(value),
		})
	}

	return br
}

func buildErrorInfoDetails(code string, details map[string]any) *errdetails.ErrorInfo {
	metadata := make(map[string]string, len(details))

	for key, value := range details {
		metadata[key] = fmt.Sprint(value)
	}

	return &errdetails.ErrorInfo{
		Reason:   code,
		Metadata: metadata,
	}
}
