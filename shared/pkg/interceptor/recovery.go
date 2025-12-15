package interceptor

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RecoveryUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				var panicMsg string
				switch x := r.(type) {
				case string:
					panicMsg = x
				case error:
					panicMsg = x.Error()
				default:
					panicMsg = fmt.Sprintf("%v", x)
				}

				log.Printf(
					"panic recovered: method=%s panic=%s\n%s",
					info.FullMethod,
					panicMsg,
					debug.Stack(),
				)

				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}
