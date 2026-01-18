package calendargeneralconfig

import (
	"os"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		ClientID:     getEnvOrDefault("CLIENT_ID", ""),
		ClientSecret: getEnvOrDefault("CLIENT_SECRET", ""),
		RedirectURL:  getEnvOrDefault("REDIRECT_URL", ""),
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
