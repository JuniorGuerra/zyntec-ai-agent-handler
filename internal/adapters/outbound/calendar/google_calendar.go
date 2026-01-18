package calendar

import (
	"app/internal/ports/outbound"
	"context"
	"log/slog"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

const state = "state-token-random"

type GoogleCalendarAdapter struct {
	oauth2.Config
}

func NewGoogleCalendarAdapter(
	clientID string,
	clientSecret string,
	redirectURL string,
) outbound.CalendarPort {
	return &GoogleCalendarAdapter{
		Config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{calendar.CalendarScope},
			Endpoint:     google.Endpoint,
		},
	}
}

func (g *GoogleCalendarAdapter) GetAuthorizationURL() string {
	return g.Config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)
}

func (g *GoogleCalendarAdapter) ExchangeToken(code string) (*oauth2.Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	token, err := g.Config.Exchange(ctx, code)
	if err != nil {
		slog.Error("failed to exchange token", "error", err)
		return nil, err
	}

	return token, nil
}

func (g *GoogleCalendarAdapter) CreateEvent(event *calendar.Event) error {
	return nil
}

func (g *GoogleCalendarAdapter) ValidateAvailability(event *calendar.Event) error {
	return nil
}
