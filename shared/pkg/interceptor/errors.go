package interceptor

import (
	"context"

	apperr "github.com/khoihuynh300/go-microservice/shared/pkg/errors"
	zaplogger "github.com/khoihuynh300/go-microservice/shared/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func ErrorHandlerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		resp, err := handler(ctx, req)

		logger := zaplogger.FromContext(ctx)

		if err != nil {
			if appErr, oke := err.(*apperr.AppError); oke {
				switch appErr.GRPCCode {
				case codes.Internal, codes.Unknown:
					logger.Error("Internal error",
						zap.String("method", info.FullMethod),
						zap.String("code", appErr.Code),
						zap.String("message", appErr.Message),
						zap.Any("details", appErr.Details),
						zap.Error(appErr.Err),
					)

				default:
					logger.Warn("Client error",
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
