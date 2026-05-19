package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Port        string
	Environment string
}

func Load() *Config {
	viper.AutomaticEnv()
	viper.SetConfigFile(".env")
	viper.ReadInConfig()

	return &Config{
		Port:        viper.GetString("PORT"),
		Environment: viper.GetString("ENVIRONMENT"),
	}
}