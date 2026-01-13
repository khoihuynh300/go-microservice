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

	storageClient, err := storage.NewMinIOStorage(storage.MinIOConfig{
		Endpoint:   config.GetEndpoint(),
		AccessKey:  config.GetAccessKey(),
		SecretKey:  config.GetSecretKey(),
		BucketName: config.GetBucketName(),
		UseSSL:     config.GetUseSSL(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	gwmux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(middleware.CustomHeaderMatcher),
		runtime.WithErrorHandler(middleware.CustomErrorHandler),
		runtime.WithForwardResponseOption(middleware.SuccessResponseModifier),
		runtime.WithMarshalerOption("*", &middleware.CustomMarshaler{}),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := userpb.RegisterUserServiceHandlerFromEndpoint(ctx, gwmux, config.GetUserServiceURL(), opts); err != nil {
		return nil, fmt.Errorf("failed to register user service handler: %w", err)
	}

	router := mux.NewRouter()

	uploadHandler := handler.NewUploadHandler(storageClient)

	s.setupRoutes(router, gwmux, uploadHandler)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", config.GetHost(), config.GetPort()),
		Handler:      router,
		ReadTimeout:  time.Duration(config.GetReadTimeout()) * time.Second,
		WriteTimeout: time.Duration(config.GetWriteTimeout()) * time.Second,
	}

	return s, nil
}

func (s *Server) Run() error {
	s.logger.Info("API Gateway listening",
		zap.String("host", config.GetHost()),
		zap.String("port", config.GetPort()))
	return s.httpServer.ListenAndServe()
}

func (s *Server) setupRoutes(router *mux.Router, gwmux *runtime.ServeMux, uploadHandler *handler.UploadHandler) {
	// health check route
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

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
	upload.HandleFunc("/products/{product_id}/images/presigned-url", uploadHandler.GetProductImagePresignedURL).Methods("POST")
	upload.HandleFunc("/categories/{category_id}/image/presigned-url", uploadHandler.GetCategoryImagePresignedURL).Methods("POST")

	// gRPC-Gateway routes
	api.PathPrefix("").Handler(gwmux)
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
