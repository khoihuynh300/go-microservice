package interceptor

import (
	"context"

	"github.com/khoihuynh300/go-microservice/shared/pkg/const/contextkeys"
	mdkeys "github.com/khoihuynh300/go-microservice/shared/pkg/const/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var publicMethods = []string{
	"/user.UserService/Login",
	"/user.UserService/Register",
	"/user.UserService/VerifyEmail",
	"/user.UserService/ResendVerificationEmail",
	"/user.UserService/Refresh",
	"/user.UserService/ForgotPassword",
	"/user.UserService/ResetPassword",
}

func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		userID, err := extractMetadata(md, mdkeys.UserIDHeader)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "missing user ID")
		}
		if userID == "" {
			return nil, status.Error(codes.Unauthenticated, "missing user ID")
		}

		ctx = context.WithValue(ctx, contextkeys.UserIDKey, userID)

		return handler(ctx, req)
	}
}

func isPublicMethod(method string) bool {
	for _, m := range publicMethods {
		if method == m {
			return true
		}
	}
	return false
}
