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

	"buf.build/go/protovalidate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/khoihuynh300/go-microservice/shared/pkg/interceptor"
	zaplogger "github.com/khoihuynh300/go-microservice/shared/pkg/logger"
	userpb "github.com/khoihuynh300/go-microservice/shared/proto/user"
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

	cfg := config.LoadConfig()

	logger, err := zaplogger.New(cfg.ServiceName, cfg.Env)
	if err != nil {
		return err
	}
	defer logger.Sync()

	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dbpool, err := initDB(dbCtx, cfg.DBUrl)
	if err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}
	defer dbpool.Close()

	hasher := passwordhasher.NewBcryptHasher(bcrypt.DefaultCost)
	jwtService := jwtprovider.NewJwtService(cfg.JwtAccessSecret, cfg.AccessTokenTTL, cfg.JwtRefreshSecret, cfg.RefreshTokenTTL)

	// repositories
	userRepository := impl.NewUserRepository(dbpool)
	refreshTokenRepository := impl.NewRefreshTokenRepository(dbpool)
	registryTokenRepository := impl.NewRegistryTokenRepository(dbpool)

	// services
	authService := service.NewAuthService(
		userRepository,
		refreshTokenRepository,
		registryTokenRepository,
		hasher,
		jwtService,
		logger,
		cfg,
	)

	// grpc handlers
	healthHandler := health.NewServer()
	userHandler := grpchandler.NewUserHandler(authService)

	validator, err := protovalidate.New()
	if err != nil {
		return fmt.Errorf("failed to initialize validator: %w", err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryUnaryInterceptor(logger),
			interceptor.ValidationUnaryInterceptor(validator),
			interceptor.LoggingUnaryInterceptor(logger),
			interceptor.ErrorHandlerInterceptor(logger),
		),
	)

	// register grpc services
	healthpb.RegisterHealthServer(grpcServer, healthHandler)
	userpb.RegisterUserServiceServer(grpcServer, userHandler)

	if cfg.Env == "DEV" {
		reflection.Register(grpcServer)
	}

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("failed to serve grpc server", zap.Error(err))
		}
	}()

	logger.Info("user service listening on", zap.String("addr", cfg.GRPCAddr))
	healthHandler.SetServingStatus(cfg.ServiceName, healthpb.HealthCheckResponse_SERVING)

	<-ctx.Done()
	logger.Info("shutdown signal received")
	healthHandler.SetServingStatus(cfg.ServiceName, healthpb.HealthCheckResponse_NOT_SERVING)

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
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return pool, nil
}
