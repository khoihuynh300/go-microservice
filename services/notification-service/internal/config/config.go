package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Env         string `mapstructure:"ENV" validate:"oneof=DEV STAG PROD TEST"`
	ServiceName string `mapstructure:"SERVICE_NAME"`
	BaseURL     string `mapstructure:"BASE_URL" validate:"required,url"`

	KafkaBrokers []string `mapstructure:"KAFKA_BROKERS" validate:"required,dive,required"`

	SMTPHost     string `mapstructure:"SMTP_HOST" validate:"required"`
	SMTPPort     int    `mapstructure:"SMTP_PORT" validate:"required"`
	SMTPUsername string `mapstructure:"SMTP_USERNAME" validate:"required"`
	SMTPPassword string `mapstructure:"SMTP_PASSWORD" validate:"required"`
	UseTLS       bool   `mapstructure:"USE_TLS"`
}

var config Config

func LoadConfig() error {
	validate := validator.New()

	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()

	viper.AutomaticEnv()

	viper.SetDefault("SERVICE_NAME", "notification-service")
	viper.SetDefault("ENV", "PROD")
	viper.SetDefault("USE_TLS", true)

	if err := viper.Unmarshal(&config); err != nil {
		return err
	}

	if err := validate.Struct(config); err != nil {
		return err
	}

	return nil
}

func GetEnv() string {
	return config.Env
}

func GetServiceName() string {
	return config.ServiceName
}

func GetBaseURL() string {
	return config.BaseURL
}

func GetKafkaBrokers() []string {
	return config.KafkaBrokers
}

func GetSMTPHost() string {
	return config.SMTPHost
}

func GetSMTPPort() int {
	return config.SMTPPort
}

func GetSMTPUsername() string {
	return config.SMTPUsername
}

func GetSMTPPassword() string {
	return config.SMTPPassword
}

func GetUseTLS() bool {
	return config.UseTLS
}
