package config

import (
	"time"

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

	// Security
	JwtAccessSecret  string        `mapstructure:"JWT_ACCESS_SECRET" validate:"required"`
	JwtRefreshSecret string        `mapstructure:"JWT_REFRESH_SECRET" validate:"required"`
	AccessTokenTTL   time.Duration `mapstructure:"ACCESS_TOKEN_TTL"`
	RefreshTokenTTL  time.Duration `mapstructure:"REFRESH_TOKEN_TTL"`

	// Redis
	RedisHost     string `mapstructure:"REDIS_HOST" validate:"required"`
	RedisPort     int    `mapstructure:"REDIS_PORT" validate:"required"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	// Kafka
	KafkaBrokers []string `mapstructure:"KAFKA_BROKERS" validate:"required"`
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
	viper.SetDefault("SERVICE_NAME", "user-service")
	viper.SetDefault("GRPC_ADDR", "localhost:5001")
	viper.SetDefault("ACCESS_TOKEN_TTL", "15m")
	viper.SetDefault("REFRESH_TOKEN_TTL", "168h")
	viper.SetDefault("REDIS_DB", 0)

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

func GetJwtAccessSecret() string {
	return config.JwtAccessSecret
}

func GetJwtRefreshSecret() string {
	return config.JwtRefreshSecret
}

func GetAccessTokenTTL() time.Duration {
	return config.AccessTokenTTL
}

func GetRefreshTokenTTL() time.Duration {
	return config.RefreshTokenTTL
}

func GetRedisHost() string {
	return config.RedisHost
}

func GetRedisPort() int {
	return config.RedisPort
}

func GetRedisPassword() string {
	return config.RedisPassword
}
func GetRedisDB() int {
	return config.RedisDB
}

func GetKafkaBrokers() []string {
	return config.KafkaBrokers
}
