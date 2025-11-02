package config

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	LocalEnv      = "local"
	SmokeEnv      = "smoke"
	TestingEnv    = "testing"
	ProductionEnv = "production"
)

type Config struct {
	HTTPServerAddr        string `mapstructure:"http_server_addr"`
	LogLevel              string `mapstructure:"log_level"`
	PostgresConnectionURL string `mapstructure:"postgres.connection_url"`
}

func LoadConfig() (Config, error) {
	viper.AddConfigPath(".")
	viper.AddConfigPath("configs")

	viper.SetConfigName(Environment())
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return Config{}, fmt.Errorf("read in config: %w", err)
	}

	var cfg Config

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return Config{}, fmt.Errorf("unmarshal: %w", err)
	}

	log.Info().Str("env", Environment()).Msg("loaded config successfully")

	return cfg, nil
}

func Environment() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		return LocalEnv
	}

	return env
}
