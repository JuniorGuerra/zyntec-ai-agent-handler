package outbound

import (
	"golang.org/x/oauth2"
)

type CalendarPort interface {
	GetAuthorizationURL() string
	ExchangeToken(code string) (*oauth2.Token, error)
	ValidateAvailability(input ValidateAvailabilityInput) error
	CreateEvent(refreshToken string, event CalendarEventInput) error
	DeleteEvent(refreshToken string, input DeleteEventInput) error
	UpdateEvent(refreshToken string, input UpdateEventInput) error
}

type CalendarEventInput struct {
	Title       string
	Description string
	StartTime   string
	EndTime     string
	Timezone    string
	Attendees   []string
}

type ValidateAvailabilityInput struct {
	RefreshToken string
	Timezone     string
	StartTime    string
	EndTime      string
}

type DeleteEventInput struct {
	Date     string
	Time     string
	Timezone string
}

type UpdateEventInput struct {
	OriginalDate string
	OriginalTime string
	NewDate      string
	NewTime      string
	Timezone     string
}
