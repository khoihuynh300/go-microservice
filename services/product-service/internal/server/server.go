package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/khoihuynh300/go-microservice/product-service/internal/config"
	grpchandler "github.com/khoihuynh300/go-microservice/product-service/internal/handler/grpc"
	"github.com/khoihuynh300/go-microservice/product-service/internal/repository/impl"
	"github.com/khoihuynh300/go-microservice/product-service/internal/service"
	"github.com/khoihuynh300/go-microservice/shared/pkg/interceptor"
	"github.com/khoihuynh300/go-microservice/shared/pkg/storage"
	productpb "github.com/khoihuynh300/go-microservice/shared/proto/product"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer    *grpc.Server
	logger        *zap.Logger
	dbPool        *pgxpool.Pool
	healthHandler *health.Server
}

func New(logger *zap.Logger) (*Server, error) {
	dbpool, err := initDB(config.GetDBUrl())
	if err != nil {
		return nil, fmt.Errorf("failed to init db: %w", err)
	}

	productRepository := impl.NewProductRepository(dbpool)
	productImageRepository := impl.NewProductImageRepository(dbpool)
	categoryRepository := impl.NewCategoryRepository(dbpool)

	minioStorage, err := storage.NewMinIOStorage(storage.MinIOConfig{
		Endpoint:   config.GetMinIOEndpoint(),
		AccessKey:  config.GetMinIOAccessKey(),
		SecretKey:  config.GetMinIOSecretKey(),
		BucketName: config.GetMinIOBucketName(),
		UseSSL:     config.GetMinIOUseSSL(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init minio storage: %w", err)
	}

	productService := service.NewProductService(productRepository, productImageRepository, categoryRepository, minioStorage)
	categoryService := service.NewCategoryService(categoryRepository, minioStorage)

	healthHandler := health.NewServer()
	productHandler := grpchandler.NewProductHandler(productService, categoryService)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.TracingInterceptor(logger),
			interceptor.RecoveryUnaryInterceptor(),
			interceptor.LoggingUnaryInterceptor(),
			interceptor.AuthInterceptor(),
			interceptor.ValidationUnaryInterceptor(),
			interceptor.ErrorHandlerInterceptor(),
		),
	)

	healthpb.RegisterHealthServer(grpcServer, healthHandler)
	productpb.RegisterProductServiceServer(grpcServer, productHandler)

	if config.GetEnv() == "DEV" {
		reflection.Register(grpcServer)
	}

	return &Server{
		grpcServer:    grpcServer,
		logger:        logger,
		dbPool:        dbpool,
		healthHandler: healthHandler,
	}, nil
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", config.GetGRPCAddr())
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.logger.Info("product service listening on", zap.String("addr", config.GetGRPCAddr()))
	s.healthHandler.SetServingStatus(config.GetServiceName(), healthpb.HealthCheckResponse_SERVING)

	return s.grpcServer.Serve(lis)
}

func (s *Server) GracefulStop() {
	s.healthHandler.SetServingStatus(config.GetServiceName(), healthpb.HealthCheckResponse_NOT_SERVING)
	s.grpcServer.GracefulStop()
	if s.dbPool != nil {
		s.dbPool.Close()
	}
}

func (s *Server) Stop() {
	s.grpcServer.Stop()
	if s.dbPool != nil {
		s.dbPool.Close()
	}
}

func initDB(dbURL string) (*pgxpool.Pool, error) {
	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(dbCtx, dbURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(dbCtx); err != nil {
		return nil, err
	}
	return pool, nil
}
