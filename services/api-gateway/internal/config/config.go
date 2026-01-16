package config

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var cfg Config

type Config struct {
	// Server settings
	Env          string `mapstructure:"ENV" validate:"oneof=DEV STAG PROD TEST"`
	ServiceName  string `mapstructure:"SERVICE_NAME"`
	Host         string `mapstructure:"HOST"`
	Port         string `mapstructure:"PORT"`
	ReadTimeout  int    `mapstructure:"READ_TIMEOUT" validate:"gte=0"`
	WriteTimeout int    `mapstructure:"WRITE_TIMEOUT" validate:"gte=0"`

	// Jwt Secret Key
	Secret string `mapstructure:"SECRET" validate:"required"`

	// Microservices URLs
	UserServiceURL    string `mapstructure:"USER_SERVICE_URL" validate:"required"`
	ProductServiceURL string `mapstructure:"PRODUCT_SERVICE_URL" validate:"required"`

	// MinIO
	MinIOEndpoint      string        `mapstructure:"MINIO_ENDPOINT" validate:"required"`
	MinIOAccessKey     string        `mapstructure:"MINIO_ACCESS_KEY" validate:"required"`
	MinIOSecretKey     string        `mapstructure:"MINIO_SECRET_KEY" validate:"required"`
	MinIOBucketName    string        `mapstructure:"MINIO_BUCKET_NAME" validate:"required"`
	MinIOUseSSL        bool          `mapstructure:"MINIO_USE_SSL"`
	PresignedURLExpiry time.Duration `mapstructure:"PRESIGNED_URL_EXPIRY"`
}

func LoadConfig() error {
	validate := validator.New()

	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	viper.AutomaticEnv()

	// Server default values
	viper.SetDefault("ENV", "PROD")
	viper.SetDefault("SERVICE_NAME", "api-gateway")
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("READ_TIMEOUT", 30)
	viper.SetDefault("WRITE_TIMEOUT", 15)

	// MinIO default values
	viper.SetDefault("MINIO_USE_SSL", true)
	viper.SetDefault("PRESIGNED_URL_EXPIRY", "15m")

	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}

	if err := validate.Struct(cfg); err != nil {
		return err
	}

	return nil
}

func GetEnv() string {
	return cfg.Env
}

func GetServiceName() string {
	return cfg.ServiceName
}

func GetHost() string {
	return cfg.Host
}

func GetPort() string {
	return cfg.Port
}

func GetReadTimeout() int {
	return cfg.ReadTimeout
}

func GetWriteTimeout() int {
	return cfg.WriteTimeout
}

func GetSecret() string {
	return cfg.Secret
}

func GetUserServiceURL() string {
	return cfg.UserServiceURL
}

func GetProductServiceURL() string {
	return cfg.ProductServiceURL
}

func GetMinIOEndpoint() string {
	return cfg.MinIOEndpoint
}

func GetMinIOAccessKey() string {
	return cfg.MinIOAccessKey
}

func GetMinIOSecretKey() string {
	return cfg.MinIOSecretKey
}

func GetMinIOBucketName() string {
	return cfg.MinIOBucketName
}

func GetMinIOUseSSL() bool {
	return cfg.MinIOUseSSL
}

func GetPresignedURLExpiry() time.Duration {
	return cfg.PresignedURLExpiry
}
