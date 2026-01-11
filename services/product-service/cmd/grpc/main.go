package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/khoihuynh300/go-microservice/product-service/internal/config"
	grpchandler "github.com/khoihuynh300/go-microservice/product-service/internal/handler/grpc"
	"github.com/khoihuynh300/go-microservice/product-service/internal/repository/impl"
	"github.com/khoihuynh300/go-microservice/product-service/internal/service"
	"github.com/khoihuynh300/go-microservice/shared/pkg/interceptor"
	zaplogger "github.com/khoihuynh300/go-microservice/shared/pkg/logger"
	productpb "github.com/khoihuynh300/go-microservice/shared/proto/product"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger, err := zaplogger.New(config.GetServiceName(), config.GetEnv())
	if err != nil {
		return err
	}
	defer logger.Sync()

	dbpool, err := initDB(config.GetDBUrl())
	if err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}
	defer dbpool.Close()

	// repositories
	productRepository := impl.NewProductRepository(dbpool)
	productImageRepository := impl.NewProductImageRepository(dbpool)
	categoryRepository := impl.NewCategoryRepository(dbpool)

	// services
	productService := service.NewProductService(productRepository, productImageRepository, categoryRepository)
	categoryService := service.NewCategoryService(categoryRepository)

	// grpc handlers
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

	// register grpc services
	healthpb.RegisterHealthServer(grpcServer, healthHandler)
	productpb.RegisterProductServiceServer(grpcServer, productHandler)

	if config.GetEnv() == "DEV" {
		reflection.Register(grpcServer)
	}

	lis, err := net.Listen("tcp", config.GetGRPCAddr())
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("failed to serve grpc server", zap.Error(err))
		}
	}()

	logger.Info("product service listening on", zap.String("addr", config.GetGRPCAddr()))
	healthHandler.SetServingStatus(config.GetServiceName(), healthpb.HealthCheckResponse_SERVING)

	<-ctx.Done()
	logger.Info("shutdown signal received")
	healthHandler.SetServingStatus(config.GetServiceName(), healthpb.HealthCheckResponse_NOT_SERVING)

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("grpc server stopped gracefully")
	case <-time.After(10 * time.Second):
		logger.Warn("graceful shutdown timeout, force stop")
		grpcServer.Stop()
	}

	return nil
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
