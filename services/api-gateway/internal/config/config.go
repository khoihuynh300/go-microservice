package config

import (
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceName    string `mapstructure:"SERVICE_NAME"`
	Host           string `mapstructure:"HOST"`
	Port           string `mapstructure:"PORT"`
	ReadTimeout    int    `mapstructure:"READ_TIMEOUT" validate:"gte=0"`
	WriteTimeout   int    `mapstructure:"WRITE_TIMEOUT" validate:"gte=0"`
	Secret         string `mapstructure:"SECRET" validate:"required"`
	UserServiceURL string `mapstructure:"USER_SERVICE_URL" validate:"required"`
	Env            string `mapstructure:"ENV" validate:"oneof=DEV STAG PROD TEST"`
}

func LoadConfig() *Config {
	var cfg Config
	validate := validator.New()

	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()

	viper.AutomaticEnv()

	viper.SetDefault("SERVICE_NAME", "api-gateway")
	viper.SetDefault("READ_TIMEOUT", 30)
	viper.SetDefault("WRITE_TIMEOUT", 15)
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("ENV", "PROD")

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	if err := validate.Struct(cfg); err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	return &cfg
}
