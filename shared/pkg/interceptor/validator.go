package interceptor

import (
	"context"

	"buf.build/go/protovalidate"
	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func ValidationUnaryInterceptor(validator protovalidate.Validator) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if err := validator.Validate(req.(proto.Message)); err != nil {
			details := make([]apperr.ErrorDetail, 0)

			if validationErr, ok := err.(*protovalidate.ValidationError); ok {
				for _, violation := range validationErr.Violations {
					fieldName := violation.Proto.GetField().GetElements()[0].GetFieldName()
					description := violation.Proto.GetMessage()

					details = append(details, apperr.ErrorDetail{
						Field:   fieldName,
						Code:    violation.Proto.GetRuleId(),
						Message: description,
					})
				}
			}

			appError := apperr.NewErrValidationFailed(details)

			return nil, apperr.ToGRPC(appError)
		}

		return handler(ctx, req)
	}
}
