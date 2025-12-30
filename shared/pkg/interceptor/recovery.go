package interceptor

import (
	"context"
	"fmt"
	"log"

	"github.com/khoihuynh300/go-microservice/shared/pkg/contextkeys"
	"go.uber.org/zap"
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

				logger, ok := ctx.Value(contextkeys.LoggerKey).(*zap.Logger)
				if !ok || logger == nil {
					log.Println("Missing logger in context for recovery interceptor")
				}
				logger.Error("panic recovered",
					zap.String("method", info.FullMethod),
					zap.String("panic", panicMsg),
					zap.Stack("stacktrace"),
				)

				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}
