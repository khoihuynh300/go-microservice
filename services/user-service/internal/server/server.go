package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/khoihuynh300/go-microservice/shared/pkg/cache"
	"github.com/khoihuynh300/go-microservice/shared/pkg/interceptor"
	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/kafka"
	"github.com/khoihuynh300/go-microservice/shared/pkg/storage"
	userpb "github.com/khoihuynh300/go-microservice/shared/proto/user"
	"github.com/khoihuynh300/go-microservice/user-service/internal/caching"
	"github.com/khoihuynh300/go-microservice/user-service/internal/config"
	"github.com/khoihuynh300/go-microservice/user-service/internal/events/publisher"
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

	hasher := passwordhasher.NewBcryptHasher(bcrypt.DefaultCost)
	jwtService := jwtprovider.NewJwtService(
		config.GetJwtAccessSecret(),
		config.GetAccessTokenTTL(),
		config.GetJwtRefreshSecret(),
		config.GetRefreshTokenTTL(),
	)

	userRepository := impl.NewUserRepository(dbpool)
	refreshTokenRepository := impl.NewRefreshTokenRepository(dbpool)
	addressRepository := impl.NewAddressRepository(dbpool)

	redis, err := cache.NewClient(&cache.Config{
		Host:     config.GetRedisHost(),
		Port:     config.GetRedisPort(),
		Password: config.GetRedisPassword(),
		DB:       config.GetRedisDB(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init redis: %w", err)
	}
	tokenCache := caching.NewTokenCache(redis)

	producer := kafka.NewProducer(config.GetKafkaBrokers())
	eventPublisher := publisher.NewKafkaEventPublisher(producer)

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

	authService := service.NewAuthService(
		userRepository,
		refreshTokenRepository,
		tokenCache,
		hasher,
		jwtService,
		eventPublisher,
	)
	userService := service.NewUserService(userRepository, minioStorage)
	addressService := service.NewAddressService(userRepository, addressRepository)

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

	healthpb.RegisterHealthServer(grpcServer, healthHandler)
	userpb.RegisterUserServiceServer(grpcServer, userHandler)

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

	s.logger.Info("user service listening on", zap.String("addr", config.GetGRPCAddr()))
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
