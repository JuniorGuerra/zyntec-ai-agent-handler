package outbound

import (
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)

type CalendarPort interface {
	GetAuthorizationURL() string
	ExchangeToken(code string) (*oauth2.Token, error)
	ValidateAvailability(event *calendar.Event) error
	CreateEvent(event *calendar.Event) error
}
