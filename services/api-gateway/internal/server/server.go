package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/khoihuynh300/go-microservice/api-gateway/internal/config"
	"github.com/khoihuynh300/go-microservice/api-gateway/internal/middleware"
	userpb "github.com/khoihuynh300/go-microservice/shared/proto/user"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

func New(logger *zap.Logger) *Server {
	return &Server{
		logger: logger,
	}
}

func (s *Server) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(middleware.CustomHeaderMatcher),
		runtime.WithErrorHandler(middleware.CustomErrorHandler),
		runtime.WithForwardResponseOption(middleware.SuccessResponseModifier),
		runtime.WithMarshalerOption("*", &middleware.CustomMarshaler{}),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := userpb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, config.GetUserServiceURL(), opts); err != nil {
		return err
	}

	// order of middleware: Tracing -> Logging -> Auth
	handler := middleware.AuthMiddleware(mux)
	handler = middleware.LoggingMiddleware(handler, s.logger)
	handler = middleware.TracingMiddleware(handler)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", config.GetHost(), config.GetPort()),
		Handler:      handler,
		ReadTimeout:  time.Duration(config.GetReadTimeout()) * time.Second,
		WriteTimeout: time.Duration(config.GetWriteTimeout()) * time.Second,
	}

	s.logger.Info("API Gateway listening",
		zap.String("host", config.GetHost()),
		zap.String("port", config.GetPort()))
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
