package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/khoihuynh300/go-microservice/api-gateway/internal/config"
	"github.com/khoihuynh300/go-microservice/api-gateway/internal/server"
	"go.uber.org/zap"

	zaplogger "github.com/khoihuynh300/go-microservice/shared/pkg/logger"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := zaplogger.New(config.GetServiceName(), config.GetEnv())
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting API Gateway", zap.String("service", config.GetServiceName()), zap.String("env", config.GetEnv()))

	srv, err := server.New(logger)
	if err != nil {
		logger.Fatal("Failed to initialize server", zap.Error(err))
	}

	go func() {
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server start failed", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	sig := <-stop
	logger.Info("Shutdown signal received", zap.String("signal", sig.String()))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}

	logger.Info("API Gateway shutdown complete")
}
