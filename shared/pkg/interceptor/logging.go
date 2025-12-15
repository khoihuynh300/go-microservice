package interceptor

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

func LoggingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		log.Printf("started unary call: %s", info.FullMethod)
		start := time.Now()
		resp, err := handler(ctx, req)

		log.Printf("method=%s duration=%s err=%v", info.FullMethod, time.Since(start), err)
		return resp, err
	}
}
