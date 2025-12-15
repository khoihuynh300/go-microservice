package main

import (
	"context"
	"log"
	"net"

	"buf.build/go/protovalidate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/khoihuynh300/go-microservice/shared/pkg/interceptor"
	userpb "github.com/khoihuynh300/go-microservice/shared/proto/user"
	"github.com/khoihuynh300/go-microservice/user-service/internal/config"
	grpchandler "github.com/khoihuynh300/go-microservice/user-service/internal/handler/grpc"
	"github.com/khoihuynh300/go-microservice/user-service/internal/repository"
	"github.com/khoihuynh300/go-microservice/user-service/internal/security/jwtprovider"
	passwordhasher "github.com/khoihuynh300/go-microservice/user-service/internal/security/password"
	"github.com/khoihuynh300/go-microservice/user-service/internal/service"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func initDB(ctx context.Context, dbURL string) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal("unable to connect db", err)
	}
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("cannot ping db", err)
	}
	return pool
}

func main() {
	ctx := context.Background()

	cfg := config.LoadConfig()

	dbpool := initDB(ctx, cfg.DBUrl)
	defer dbpool.Close()

	hasher := passwordhasher.NewBcryptHasher(bcrypt.DefaultCost)
	jwtService := jwtprovider.NewJwtService(cfg.JwtAccessSecret, cfg.AccessTokenTTL, cfg.JwtRefreshSecret, cfg.RefreshTokenTTL)

	userRepository := repository.NewUserRepository(dbpool)
	refreshTokenRepository := repository.NewRefreshTokenRepository(dbpool)

	authService := service.NewAuthService(userRepository, refreshTokenRepository, hasher, jwtService)

	userHandler := grpchandler.NewUserHandler(authService)

	validator, err := protovalidate.New()
	if err != nil {
		log.Fatal("failed to initialize validator:", err)
	}
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryUnaryInterceptor(),
			interceptor.LoggingUnaryInterceptor(),
			interceptor.ValidationUnaryInterceptor(validator),
		),
	)
	userpb.RegisterUserServiceServer(grpcServer, userHandler)

	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("user service listening on " + cfg.GRPCAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
