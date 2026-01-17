package config

import (
	"os"
)

type Config struct {
	WAHAURL string
}

var WhatsappSQSUrl string

func LoadConfig() (*Config, error) {
	cfg := &Config{
		WAHAURL: getEnvOrDefault("WAHA_URL", "https://api.waha.ai/v1/send-message"),
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
