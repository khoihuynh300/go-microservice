package interceptor

import (
	"context"
	"time"

	zaplogger "github.com/khoihuynh300/go-microservice/shared/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func LoggingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		logger := zaplogger.FromContext(ctx)
		logger.Info("grpc request", zap.String("method", info.FullMethod))

		resp, err := handler(ctx, req)

		logger.Info("grpc response",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)
		return resp, err
	}
}
