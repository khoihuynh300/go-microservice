package interceptor

import (
	"context"
	"time"

	"github.com/khoihuynh300/go-microservice/shared/pkg/contextkeys"
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

		logger, _ := ctx.Value(contextkeys.LoggerKey).(*zap.Logger)
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
