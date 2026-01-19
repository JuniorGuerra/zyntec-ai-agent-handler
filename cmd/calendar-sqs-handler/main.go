package main

import (
	config "app/cmd/config/calendar-general-config"
	"app/internal/adapters/inbound/sqs"
	"app/internal/adapters/outbound/calendar"
	"app/internal/adapters/outbound/persistence"
	calendarservice "app/internal/application/calendar"
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

	calendarAdapter := calendar.NewGoogleCalendarAdapter(
		cfg.ClientID,
		cfg.ClientSecret,
		cfg.RedirectURL,
	)

	dbClient := persistence.NewDynamoDBClient()
	calendarRepo := persistence.NewCalendarRepository(dbClient)

	service := calendarservice.NewCalendarService(calendarAdapter, calendarRepo)
	handler := sqs.NewCalendarHandler(service)

	lambda.Start(handler.HandleSQSMessage)
}
