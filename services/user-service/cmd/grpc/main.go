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
	"github.com/khoihuynh300/go-microservice/shared/pkg/cache"
	"github.com/khoihuynh300/go-microservice/shared/pkg/interceptor"
	zaplogger "github.com/khoihuynh300/go-microservice/shared/pkg/logger"
	userpb "github.com/khoihuynh300/go-microservice/shared/proto/user"
	"github.com/khoihuynh300/go-microservice/user-service/internal/caching"
	"github.com/khoihuynh300/go-microservice/user-service/internal/config"
	grpchandler "github.com/khoihuynh300/go-microservice/user-service/internal/handler/grpc"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository/impl"
	"github.com/khoihuynh300/go-microservice/user-service/internal/security/jwtprovider"
	passwordhasher "github.com/khoihuynh300/go-microservice/user-service/internal/security/password"
	"github.com/khoihuynh300/go-microservice/user-service/internal/service"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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

	logger, err := zaplogger.New(config.ServiceName, config.Env)
	if err != nil {
		return err
	}
	defer logger.Sync()

	dbpool, err := initDB(ctx, config.DBUrl)
	if err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}
	defer dbpool.Close()

	hasher := passwordhasher.NewBcryptHasher(bcrypt.DefaultCost)
	jwtService := jwtprovider.NewJwtService(
		config.JwtAccessSecret,
		config.AccessTokenTTL,
		config.JwtRefreshSecret,
		config.RefreshTokenTTL,
	)

	// repositories
	userRepository := impl.NewUserRepository(dbpool)
	refreshTokenRepository := impl.NewRefreshTokenRepository(dbpool)
	addressRepository := impl.NewAddressRepository(dbpool)

	// caching
	redis, err := cache.NewClient(&cache.Config{
		Host:     config.RedisHost,
		Port:     config.RedisPort,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})
	tokenCache := caching.NewTokenCache(redis)

	// services
	authService := service.NewAuthService(
		userRepository,
		refreshTokenRepository,
		tokenCache,
		hasher,
		jwtService,
	)
	userService := service.NewUserService(userRepository)
	addressService := service.NewAddressService(userRepository, addressRepository)

	// grpc handlers
	healthHandler := health.NewServer()
	userHandler := grpchandler.NewUserHandler(authService, userService, addressService)

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
	userpb.RegisterUserServiceServer(grpcServer, userHandler)

	if config.Env == "DEV" {
		reflection.Register(grpcServer)
	}

	lis, err := net.Listen("tcp", config.GRPCAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("failed to serve grpc server", zap.Error(err))
		}
	}()

	logger.Info("user service listening on", zap.String("addr", config.GRPCAddr))
	healthHandler.SetServingStatus(config.ServiceName, healthpb.HealthCheckResponse_SERVING)

	<-ctx.Done()
	logger.Info("shutdown signal received")
	healthHandler.SetServingStatus(config.ServiceName, healthpb.HealthCheckResponse_NOT_SERVING)

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

func initDB(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
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
