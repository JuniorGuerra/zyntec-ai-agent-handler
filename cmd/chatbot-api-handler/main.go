package main

import (
	"context"
	"log/slog"
	"os"

	config "app/cmd/config/chatbot-api-config"
	httphandler "app/internal/adapters/inbound/http"
	"app/internal/adapters/outbound/ai"
	"app/internal/adapters/outbound/persistence"
	"app/internal/adapters/outbound/sqs"
	"app/internal/application/chatbot"
	"app/internal/ports/outbound"

	"github.com/aws/aws-lambda-go/lambda"
	"golang.org/x/sync/errgroup"
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

	var (
		dbClient   *persistence.DynamoDBClient
		aiAdapter  *ai.GeminiAdapter
		sqsAdapter outbound.SQSAdapter
	)

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		dbClient = persistence.NewDynamoDBClient()
		return nil
	})

	g.Go(func() error {
		var err error
		aiAdapter, err = ai.NewGeminiAdapter(cfg.GeminiAPIKey)
		return err
	})

	g.Go(func() error {
		sqsAdapter = sqs.NewSQSAdapter()
		return nil
	})

	if err := g.Wait(); err != nil {
		slog.Error("failed to initialize services", "error", err)
		return
	}

	sessionRepo := persistence.NewSessionRepository(dbClient)
	messageRepo := persistence.NewMessageRepository(dbClient)
	customerRepo := persistence.NewCustomerRepository(dbClient)

	service := chatbot.NewService(sessionRepo, messageRepo, customerRepo, aiAdapter)
	handler := httphandler.NewChatbotHandler(service, sqsAdapter, cfg.WhatsappSQSUrl, cfg.CalendarSQSUrl)

	lambda.Start(handler.HandleWebhook)
}
