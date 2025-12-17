package interceptor

import (
	"context"

	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func ErrorHandlerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		resp, err := handler(ctx, req)

		if err != nil {
			if appErr, oke := err.(*apperr.AppError); oke {
				switch appErr.GRPCCode {
				case codes.Internal, codes.Unknown:
					logger.Error("Internal error",
						zap.String("method", info.FullMethod),
						zap.String("code", appErr.Code),
						zap.String("message", appErr.Message),
						zap.Any("details", appErr.Details),
						zap.Error(appErr.OriginalError),
					)

				case codes.NotFound, codes.AlreadyExists, codes.InvalidArgument:
					logger.Debug("Client error",
						zap.String("method", info.FullMethod),
						zap.String("code", appErr.Code),
						zap.String("message", appErr.Message),
					)

				default:
					logger.Warn("Request error",
						zap.String("method", info.FullMethod),
						zap.String("code", appErr.Code),
						zap.String("message", appErr.Message),
					)
				}

			} else {
				logger.Error("Internal error",
					zap.String("method", info.FullMethod),
					zap.Error(err),
				)
			}

			return resp, apperr.ToGRPC(err)
		}

		return resp, nil
	}
}
