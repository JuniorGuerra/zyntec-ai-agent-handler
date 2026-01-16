package config

import (
	"fmt"
	"os"
)

type Config struct {
	GeminiAPIKey   string
	AWSRegion      string
	WhatsappSQSUrl string
}

var WhatsappSQSUrl string

func LoadConfig() (*Config, error) {
	cfg := &Config{
		GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
		AWSRegion:      getEnvOrDefault("AWS_REGION", "us-east-1"),
		WhatsappSQSUrl: getEnvOrDefault("WHATSAPP_SQS_URL", "https://sqs.us-east-1.amazonaws.com/492017761132/send-whatsapp-message-queue"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	WhatsappSQSUrl = cfg.WhatsappSQSUrl
	return cfg, nil
}

func (c *Config) validate() error {
	if c.GeminiAPIKey == "" {
		return fmt.Errorf("GEMINI_API_KEY is required")
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
