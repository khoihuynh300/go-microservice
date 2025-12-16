package interceptor

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func LoggingUnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		logger.Info("unary request", zap.String("method", info.FullMethod))
		start := time.Now()
		resp, err := handler(ctx, req)

		logger.Info("unary response",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)
		return resp, err
	}
}
