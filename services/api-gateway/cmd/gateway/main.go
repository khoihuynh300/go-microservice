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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Shutdown signal received", zap.String("signal", sig.String()))
		cancel()
	}()

	srv := server.New(logger)

	errChan := make(chan error, 1)
	go func() {
		if err := srv.Run(ctx); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		logger.Fatal("Server error", zap.Error(err))
	case <-ctx.Done():
		logger.Info("Shutting down gracefully...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error during shutdown", zap.Error(err))
		}
	}

	logger.Info("API Gateway shutdown complete")
}
