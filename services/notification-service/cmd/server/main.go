package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/khoihuynh300/go-microservice/notification-service/internal/config"
	"github.com/khoihuynh300/go-microservice/notification-service/internal/events/consumer"
	"github.com/khoihuynh300/go-microservice/notification-service/internal/events/handlers"
	"github.com/khoihuynh300/go-microservice/notification-service/internal/service"
	"github.com/khoihuynh300/go-microservice/notification-service/internal/template"
	"github.com/khoihuynh300/go-microservice/shared/pkg/messaging/kafka"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	err := config.LoadConfig()
	if err != nil {
		return err
	}

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

	if err := startKafkaConsumer(emailService); err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	return nil
}

func startKafkaConsumer(emailService service.EmailService) error {
	kafkaConsumer := kafka.NewConsumer(config.GetKafkaBrokers())
	eventConsumer := consumer.NewEventConsumer(kafkaConsumer)

	userEventHandler := handlers.NewUserEventHandler(emailService, config.GetBaseURL())
	eventConsumer.RegisterHandler(userEventHandler)

	go func() {
		ctx := context.Background()
		log.Println("Listening for Kafka events...")

		if err := eventConsumer.Start(ctx); err != nil {
			log.Printf("Kafka consumer error: %v", err)
		}
	}()

	return nil
}
