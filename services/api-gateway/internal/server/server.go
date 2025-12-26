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
	cfg        *config.Config
	httpServer *http.Server
	logger     *zap.Logger
}

func New(cfg *config.Config, logger *zap.Logger) *Server {
	return &Server{
		cfg:    cfg,
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

	if err := userpb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, s.cfg.UserServiceURL, opts); err != nil {
		return err
	}

	handler := middleware.LoggingMiddleware(mux, s.logger)
	handler = middleware.TracingMiddleware(handler)
	handler = middleware.AuthMiddleware(handler, s.cfg)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port),
		Handler:      handler,
		ReadTimeout:  time.Duration(s.cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.cfg.WriteTimeout) * time.Second,
	}

	s.logger.Info("user service listening on", zap.String("addr", s.cfg.Port))
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
