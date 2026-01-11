package main

import (
	"app/internal/adapters/handlers"
	"app/internal/adapters/repositories"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func main() {

	repository := repositories.NewDynamoDBRepository()
	handler := handlers.NewChatbotAPIHandler(repository)

	lambda.Start(handler.Handle)
}
