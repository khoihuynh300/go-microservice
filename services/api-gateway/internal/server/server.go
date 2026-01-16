package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/khoihuynh300/go-microservice/api-gateway/internal/config"
	"github.com/khoihuynh300/go-microservice/api-gateway/internal/handler"
	"github.com/khoihuynh300/go-microservice/api-gateway/internal/middleware"
	"github.com/khoihuynh300/go-microservice/shared/pkg/storage"
	productpb "github.com/khoihuynh300/go-microservice/shared/proto/product"
	userpb "github.com/khoihuynh300/go-microservice/shared/proto/user"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

func New(logger *zap.Logger) (*Server, error) {
	ctx := context.Background()

	s := &Server{
		logger: logger,
	}

	// Initialize storage client
	storageClient, err := storage.NewMinIOStorage(storage.MinIOConfig{
		Endpoint:   config.GetMinIOEndpoint(),
		AccessKey:  config.GetMinIOAccessKey(),
		SecretKey:  config.GetMinIOSecretKey(),
		BucketName: config.GetMinIOBucketName(),
		UseSSL:     config.GetMinIOUseSSL(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Set up gRPC-Gateway
	gwmux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(middleware.CustomHeaderMatcher),
		runtime.WithErrorHandler(middleware.CustomErrorHandler),
		runtime.WithForwardResponseOption(middleware.SuccessResponseModifier),
		runtime.WithMarshalerOption("*", &middleware.CustomMarshaler{}),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Register user service handler
	if err := userpb.RegisterUserServiceHandlerFromEndpoint(ctx, gwmux, config.GetUserServiceURL(), opts); err != nil {
		return nil, fmt.Errorf("failed to register user service handler: %w", err)
	}

	if err := productpb.RegisterProductServiceHandlerFromEndpoint(ctx, gwmux, config.GetProductServiceURL(), opts); err != nil {
		return nil, fmt.Errorf("failed to register product service handler: %w", err)
	}

	// Initialize upload handler
	uploadHandler := handler.NewUploadHandler(storageClient)

	// Set up HTTP server
	router := mux.NewRouter()
	s.setupRoutes(router, gwmux, uploadHandler)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", config.GetHost(), config.GetPort()),
		Handler:      router,
		ReadTimeout:  time.Duration(config.GetReadTimeout()) * time.Second,
		WriteTimeout: time.Duration(config.GetWriteTimeout()) * time.Second,
	}

	return s, nil
}

func (s *Server) setupRoutes(router *mux.Router, gwmux *runtime.ServeMux, uploadHandler *handler.UploadHandler) {
	// health check route
	router.HandleFunc("/health", handler.HealthCheck).Methods("GET")

	// Path: /v1
	api := router.PathPrefix("/v1").Subrouter()
	api.Use(func(next http.Handler) http.Handler {
		handler := middleware.AuthMiddleware(next)
		handler = middleware.LoggingMiddleware(handler, s.logger)
		handler = middleware.TracingMiddleware(handler)
		return handler
	})

	// upload routes
	upload := api.PathPrefix("/upload").Subrouter()
	upload.HandleFunc("/avatar/presigned-url", uploadHandler.GetAvatarPresignedURL).Methods("POST")
	upload.HandleFunc("/products/{product_id}/thumbnail/presigned-url", uploadHandler.GetProductImagePresignedURL).Methods("POST")
	upload.HandleFunc("/products/{product_id}/image/presigned-url", uploadHandler.GetProductImagePresignedURL).Methods("POST")
	upload.HandleFunc("/categories/{category_id}/image/presigned-url", uploadHandler.GetCategoryImagePresignedURL).Methods("POST")

	// gRPC-Gateway routes
	api.PathPrefix("").Handler(gwmux)
}

func (s *Server) Run() error {
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
