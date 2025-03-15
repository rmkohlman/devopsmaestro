package config

import (
	"log"

	"github.com/spf13/viper"
)

func LoadConfig(configPath string) {
	viper.SetConfigName("config")
	viper.AddConfigPath(configPath)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error loading config file: %v", err)
	}
}
