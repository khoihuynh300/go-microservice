package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	// Service
	ServiceName string `mapstructure:"SERVICE_NAME"`
	GRPCAddr    string `mapstructure:"GRPC_ADDR"`
	Env         string `mapstructure:"ENV"`

	// Database
	DBUrl string `mapstructure:"DATABASE_URL" validate:"required"`
}

var config Config

func LoadConfig() error {
	validate := validator.New()

	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	viper.AutomaticEnv()

	viper.SetDefault("ENV", "DEV")
	viper.SetDefault("SERVICE_NAME", "product-service")
	viper.SetDefault("GRPC_ADDR", "localhost:5000")

	if err := viper.Unmarshal(&config); err != nil {
		return err
	}

	if err := validate.Struct(config); err != nil {
		return err
	}

	return nil
}

func GetServiceName() string {
	return config.ServiceName
}

func GetGRPCAddr() string {
	return config.GRPCAddr
}

func GetEnv() string {
	return config.Env
}

func GetDBUrl() string {
	return config.DBUrl
}
