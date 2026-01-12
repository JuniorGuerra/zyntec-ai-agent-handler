package main

import (
	"app/internal/adapters/handlers"
	"app/internal/adapters/repositories"
	"app/internal/core/services/ai"
	whatsappsvc "app/internal/core/services/whatsapp-svc"
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"golang.org/x/sync/errgroup"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func main() {

	var (
		repository  repositories.Repository
		wahaService whatsappsvc.WhatsAppService
		aiService   ai.AIModels
	)

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		repository = repositories.NewDynamoDBRepository()
		return nil
	})

	g.Go(func() error {
		wahaService = whatsappsvc.NewWahaService()
		return nil
	})

	g.Go(func() error {
		var err error
		aiService, err = ai.NewGeminiService(os.Getenv("GEMINI_API_KEY"))
		return err
	})

	if err := g.Wait(); err != nil {
		slog.Error("failed to initialize services", "error", err)
		return
	}

	handler := handlers.NewChatbotAPIHandler(repository, aiService, wahaService)

	lambda.Start(handler.Handle)
}
