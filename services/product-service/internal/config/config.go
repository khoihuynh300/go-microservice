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

	// MinIO
	MinIOEndpoint   string `mapstructure:"MINIO_ENDPOINT" validate:"required"`
	MinIOAccessKey  string `mapstructure:"MINIO_ACCESS_KEY" validate:"required"`
	MinIOSecretKey  string `mapstructure:"MINIO_SECRET_KEY" validate:"required"`
	MinIOBucketName string `mapstructure:"MINIO_BUCKET_NAME" validate:"required"`
	MinIOUseSSL     bool   `mapstructure:"MINIO_USE_SSL"`
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

func GetMinIOEndpoint() string {
	return config.MinIOEndpoint
}

func GetMinIOAccessKey() string {
	return config.MinIOAccessKey
}

func GetMinIOSecretKey() string {
	return config.MinIOSecretKey
}

func GetMinIOBucketName() string {
	return config.MinIOBucketName
}

func GetMinIOUseSSL() bool {
	return config.MinIOUseSSL
}
