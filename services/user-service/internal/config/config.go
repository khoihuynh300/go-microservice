package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

var (
	// Service
	ServiceName string
	GRPCAddr    string
	Env         string

	// Database
	DBUrl string

	// Security
	JwtAccessSecret  string
	JwtRefreshSecret string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration

	// Redis
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int
)

func LoadConfig() error {
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	viper.SetDefault("ENV", "DEV")
	viper.SetDefault("SERVICE_NAME", "user-service")
	viper.SetDefault("GRPC_ADDR", "localhost:5001")
	viper.SetDefault("ACCESS_TOKEN_TTL", "15m")
	viper.SetDefault("REFRESH_TOKEN_TTL", "168h")
	viper.SetDefault("REDIS_DB", 0)

	requiredVars := []string{
		"DATABASE_URL",
		"JWT_ACCESS_SECRET",
		"JWT_REFRESH_SECRET",
		"REDIS_HOST",
		"REDIS_PORT",
	}

	for _, key := range requiredVars {
		if !viper.IsSet(key) {
			return fmt.Errorf("required environment variable %s is not set", key)
		}
	}

	Env = viper.GetString("ENV")
	ServiceName = viper.GetString("SERVICE_NAME")
	GRPCAddr = viper.GetString("GRPC_ADDR")

	DBUrl = viper.GetString("DATABASE_URL")

	JwtAccessSecret = viper.GetString("JWT_ACCESS_SECRET")
	JwtRefreshSecret = viper.GetString("JWT_REFRESH_SECRET")
	AccessTokenTTL = viper.GetDuration("ACCESS_TOKEN_TTL")
	RefreshTokenTTL = viper.GetDuration("REFRESH_TOKEN_TTL")

	RedisHost = viper.GetString("REDIS_HOST")
	RedisPort = viper.GetInt("REDIS_PORT")
	RedisPassword = viper.GetString("REDIS_PASSWORD")
	RedisDB = viper.GetInt("REDIS_DB")

	return nil
}
