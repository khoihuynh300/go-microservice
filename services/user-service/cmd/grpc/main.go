package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	zaplogger "github.com/khoihuynh300/go-microservice/shared/pkg/logger"
	"github.com/khoihuynh300/go-microservice/user-service/internal/config"
	"github.com/khoihuynh300/go-microservice/user-service/internal/server"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger, err := zaplogger.New(config.GetServiceName(), config.GetEnv())
	if err != nil {
		return err
	}
	defer logger.Sync()

	srv, err := server.New(logger)
	if err != nil {
		return err
	}

	go func() {
		if err := srv.Run(); err != nil {
			logger.Error("failed to serve grpc server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	logger.Info("shutdown signal received")

	done := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("grpc server stopped gracefully")
	case <-time.After(10 * time.Second):
		logger.Warn("graceful shutdown timeout, force stop")
		srv.Stop()
	}

	return nil
}
