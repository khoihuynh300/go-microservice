package main

import (
	"context"
	"log"

	"github.com/khoihuynh300/go-microservice/api-gateway/internal/config"
	"github.com/khoihuynh300/go-microservice/api-gateway/internal/server"

	zaplogger "github.com/khoihuynh300/go-microservice/shared/pkg/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.LoadConfig()

	logger, err := zaplogger.New(cfg.ServiceName, cfg.Env)
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	srv := server.New(cfg, logger)

	if err := srv.Run(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
