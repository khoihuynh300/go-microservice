package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/khoihuynh300/go-microservice/notification-service/internal/config"
	"github.com/khoihuynh300/go-microservice/notification-service/internal/events/handlers"
	"github.com/khoihuynh300/go-microservice/notification-service/internal/service"
	"github.com/khoihuynh300/go-microservice/notification-service/internal/template"
	zaplogger "github.com/khoihuynh300/go-microservice/shared/pkg/logger"
	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/kafka"
	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/topics"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := config.LoadConfig(); err != nil {
		return err
	}

	logger, err := zaplogger.New(config.GetServiceName(), config.GetEnv())
	if err != nil {
		return err
	}
	defer logger.Sync()

	templateParser, err := template.NewParser()
	if err != nil {
		return err
	}

	emailService := service.NewEmailService(
		templateParser,
		config.GetSMTPHost(),
		config.GetSMTPPort(),
		config.GetSMTPUsername(),
		config.GetSMTPPassword(),
		config.GetUseTLS(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaConsumer, err := startKafkaConsumer(ctx, emailService, logger)
	if err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down gracefully...")

	cancel()

	if err := kafkaConsumer.Close(); err != nil {
		logger.Error("Error closing Kafka consumer", zap.Error(err))
	}

	logger.Info("Service stopped")
	return nil
}

func startKafkaConsumer(ctx context.Context, emailService service.EmailService, logger *zap.Logger) (kafka.Consumer, error) {
	topicConsume := []string{topics.UserEventsTopic}
	kafkaConsumer := kafka.NewConsumer(config.GetKafkaBrokers(), topicConsume, config.GetKafkaConsumerGroup())

	userEventHandler := handlers.NewUserEventHandler(emailService, config.GetBaseURL())
	kafkaConsumer.RegisterHandler(topics.UserEventsTopic, userEventHandler.HandleEvent)

	go func() {
		logger.Info("Starting Kafka consumer...")
		if err := kafkaConsumer.Start(ctx, logger); err != nil {
			logger.Error("Kafka consumer error", zap.Error(err))
		}
	}()

	return kafkaConsumer, nil
}
