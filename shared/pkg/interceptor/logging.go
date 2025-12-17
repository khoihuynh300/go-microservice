package interceptor

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func LoggingUnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		md, _ := metadata.FromIncomingContext(ctx)
		requestID := extractMetadata(md, "x-request-id")

		if requestID != "" {
			ctx = context.WithValue(ctx, "request_id", requestID)
		}

		logger.Info("unary request", zap.String("request_id", requestID), zap.String("method", info.FullMethod))
		resp, err := handler(ctx, req)

		logger.Info("unary response",
			zap.String("request_id", requestID),
			zap.String("method", info.FullMethod),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)
		return resp, err
	}
}

func extractMetadata(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
