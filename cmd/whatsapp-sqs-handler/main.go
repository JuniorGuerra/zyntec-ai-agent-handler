package main

import (
	config "app/cmd/config/whatsapp-sqs-config"
	"app/internal/adapters/inbound/sqs"
	"app/internal/adapters/outbound/persistence"
	"app/internal/adapters/outbound/whatsapp"
	"app/internal/application/waha"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}

	dbClient := persistence.NewDynamoDBClient()
	customerRepo := persistence.NewCustomerRepository(dbClient)

	wahaAdapter := whatsapp.NewWahaAdapter(cfg.WAHAURL)
	wahaHandler := waha.NewService(customerRepo, wahaAdapter)
	handler := sqs.NewWhatsAppHandler(wahaHandler)

	lambda.Start(handler.HandleSQSMessage)
}
