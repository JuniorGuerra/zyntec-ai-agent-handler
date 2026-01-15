package main

import (
	"context"
	"log/slog"
	"os"

	httphandler "app/internal/adapters/inbound/http"
	"app/internal/adapters/outbound/ai"
	"app/internal/adapters/outbound/persistence"
	"app/internal/application/chatbot"

	"github.com/aws/aws-lambda-go/lambda"
	"golang.org/x/sync/errgroup"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func main() {
	var (
		dbClient   *persistence.DynamoDBClient
		aiAdapter  *ai.GeminiAdapter
	)

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		dbClient = persistence.NewDynamoDBClient()
		return nil
	})

	g.Go(func() error {
		var err error
		aiAdapter, err = ai.NewGeminiAdapter(os.Getenv("GEMINI_API_KEY"))
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("failed to initialize services", "error", err)
		return
	}

	sessionRepo := persistence.NewSessionRepository(dbClient)
	messageRepo := persistence.NewMessageRepository(dbClient)
	customerRepo := persistence.NewCustomerRepository(dbClient)

	service := chatbot.NewService(sessionRepo, messageRepo, customerRepo, aiAdapter)
	handler := httphandler.NewChatbotHandler(service)

	lambda.Start(handler.Handle)
}
