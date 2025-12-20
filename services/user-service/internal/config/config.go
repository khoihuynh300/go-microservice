package config

import (
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceName               string        `mapstructure:"SERVICE_NAME"`
	GRPCAddr                  string        `mapstructure:"GRPC_ADDR"`
	DBUrl                     string        `mapstructure:"DATABASE_URL" validate:"required"`
	JwtAccessSecret           string        `mapstructure:"JWT_ACCESS_SECRET" validate:"required"`
	JwtRefreshSecret          string        `mapstructure:"JWT_REFRESH_SECRET" validate:"required"`
	AccessTokenTTL            time.Duration `mapstructure:"TTL_ACCESS_TOKEN" validate:"required"`
	RefreshTokenTTL           time.Duration `mapstructure:"TTL_REFRESH_TOKEN" validate:"required"`
	RegistryTokenExpiry       time.Duration `mapstructure:"REGISTRY_TOKEN_EXPIRY"`
	ForgotPasswordTokenExpiry time.Duration `mapstructure:"FORGOT_PASSWORD_TOKEN_EXPIRY"`
	Env                       string        `mapstructure:"ENV" validate:"oneof=DEV STAG PROD TEST"`
}

var (
	cfg Config
)

func LoadConfig() *Config {
	validate := validator.New()

	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()

	viper.AutomaticEnv()

	viper.SetDefault("SERVICE_NAME", "user-service")
	viper.SetDefault("GRPC_ADDR", ":5000")
	viper.SetDefault("REGISTRY_TOKEN_EXPIRY", "1h")
	viper.SetDefault("FORGOT_PASSWORD_TOKEN_EXPIRY", "15m")
	viper.SetDefault("ENV", "PROD")

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	if err := validate.Struct(cfg); err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	return &cfg
}
