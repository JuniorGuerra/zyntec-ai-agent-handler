package main

import (
	calendargeneralconfig "app/cmd/config/calendar-general-config"
	"app/internal/adapters/outbound/calendar"
	"app/internal/adapters/outbound/persistence"
	calendarstarter "app/internal/application/calendar-starter"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func main() {
	cfg, err := calendargeneralconfig.LoadConfig()

	if err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}

	dbClient := persistence.NewDynamoDBClient()
	calendarRepo := persistence.NewCalendarRepository(dbClient)
	calendarPort := calendar.NewGoogleCalendarAdapter(
		cfg.ClientID,
		cfg.ClientSecret,
		cfg.RedirectURL,
	)

	calendarService := calendarstarter.NewCalendarStarterHandler(
		calendarPort,
		calendarRepo,
	)

	lambda.Start(calendarService.Handler)
}
