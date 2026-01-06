package testutil

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/khoihuynh300/go-microservice/shared/pkg/cache"
	mdkeys "github.com/khoihuynh300/go-microservice/shared/pkg/const/metadata"
	"github.com/khoihuynh300/go-microservice/shared/pkg/interceptor"
	userpb "github.com/khoihuynh300/go-microservice/shared/proto/user"
	"github.com/khoihuynh300/go-microservice/user-service/internal/caching"
	grpchandler "github.com/khoihuynh300/go-microservice/user-service/internal/handler/grpc"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository/impl"
	"github.com/khoihuynh300/go-microservice/user-service/internal/security/jwtprovider"
	passwordhasher "github.com/khoihuynh300/go-microservice/user-service/internal/security/password"
	"github.com/khoihuynh300/go-microservice/user-service/internal/service"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type TestGRPCServer struct {
	Server   *grpc.Server
	Listener *bufconn.Listener
	Conn     *grpc.ClientConn
	Client   userpb.UserServiceClient
}

type GRPCServerConfig struct {
	JwtAccessSecret  string
	JwtRefreshSecret string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
}

func DefaultGRPCServerConfig() *GRPCServerConfig {
	return &GRPCServerConfig{
		JwtAccessSecret:  "test-access-secret-key-for-testing",
		JwtRefreshSecret: "test-refresh-secret-key-for-testing",
		AccessTokenTTL:   15 * time.Minute,
		RefreshTokenTTL:  7 * 24 * time.Hour,
	}
}

func NewTestGRPCServer(
	ctx context.Context,
	db *TestDatabase,
	redis *TestRedis,
	cfg *GRPCServerConfig,
) (*TestGRPCServer, error) {
	if cfg == nil {
		cfg = DefaultGRPCServerConfig()
	}

	lis := bufconn.Listen(bufSize)

	server, err := setupGRPCServer(db, redis.Client, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to setup grpc server: %w", err)
	}

	go func() {
		if err := server.Serve(lis); err != nil {
			fmt.Printf("GRPC server exited with error: %v\n", err)
		}
	}()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(traceIDClientInterceptor()),
	)
	if err != nil {
		server.Stop()
		return nil, fmt.Errorf("failed to dial bufnet: %w", err)
	}

	client := userpb.NewUserServiceClient(conn)

	return &TestGRPCServer{
		Server:   server,
		Listener: lis,
		Conn:     conn,
		Client:   client,
	}, nil
}

func setupGRPCServer(db *TestDatabase, cacheClient cache.Cache, cfg *GRPCServerConfig) (*grpc.Server, error) {
	logger := zap.NewNop()

	// Repositories
	userRepo := impl.NewUserRepository(db.Pool)
	refreshTokenRepo := impl.NewRefreshTokenRepository(db.Pool)
	addressRepo := impl.NewAddressRepository(db.Pool)

	// Security
	hasher := passwordhasher.NewBcryptHasher(bcrypt.DefaultCost)
	jwtService := jwtprovider.NewJwtService(
		cfg.JwtAccessSecret,
		cfg.AccessTokenTTL,
		cfg.JwtRefreshSecret,
		cfg.RefreshTokenTTL,
	)

	// Caching
	tokenCache := caching.NewTokenCache(cacheClient)

	// Services
	authService := service.NewAuthService(
		userRepo,
		refreshTokenRepo,
		tokenCache,
		hasher,
		jwtService,
	)
	userService := service.NewUserService(userRepo)
	addressService := service.NewAddressService(userRepo, addressRepo)

	// Handler
	userHandler := grpchandler.NewUserHandler(authService, userService, addressService)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.TracingInterceptor(logger),
			interceptor.RecoveryUnaryInterceptor(),
			// interceptor.LoggingUnaryInterceptor(),
			interceptor.AuthInterceptor(),
			interceptor.ValidationUnaryInterceptor(),
			interceptor.ErrorHandlerInterceptor(),
		),
	)

	// Register services
	userpb.RegisterUserServiceServer(grpcServer, userHandler)

	return grpcServer, nil
}

func (s *TestGRPCServer) TearDown() {
	if s.Conn != nil {
		s.Conn.Close()
	}
	if s.Server != nil {
		s.Server.Stop()
	}
}

func (s *TestGRPCServer) GetClient() userpb.UserServiceClient {
	return s.Client
}

func traceIDClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		traceID := uuid.New().String()
		ctx = metadata.AppendToOutgoingContext(ctx, mdkeys.TraceIDHeader, traceID)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
