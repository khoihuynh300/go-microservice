package interceptor

import (
	"context"
	"fmt"
	"time"

	"github.com/khoihuynh300/go-microservice/shared/pkg/contextkeys"
	mdkeys "github.com/khoihuynh300/go-microservice/shared/pkg/metadata"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func LoggingUnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		traceID, err := extractMetadata(md, mdkeys.TraceIDHeader)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "missing trace ID")
		}
		if traceID == "" {
			return nil, status.Error(codes.Unauthenticated, "missing trace ID")
		}

		ctxLogger := logger.With(zap.String("trace_id", traceID))
		ctx = context.WithValue(ctx, contextkeys.TraceIDKey, traceID)
		ctx = context.WithValue(ctx, contextkeys.LoggerKey, ctxLogger)
		ctxLogger.Info("grpc request", zap.String("method", info.FullMethod))

		resp, err := handler(ctx, req)

		ctxLogger.Info("grpc response",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)
		return resp, err
	}
}

func extractMetadata(md metadata.MD, key string) (string, error) {
	values := md.Get(key)
	if len(values) > 0 {
		return values[0], nil
	}
	return "", fmt.Errorf("key '%s' not found in metadata", key)
}
