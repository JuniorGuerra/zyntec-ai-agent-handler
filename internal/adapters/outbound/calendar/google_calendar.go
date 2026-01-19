package calendar

import (
	"app/internal/ports/outbound"
	"context"
	"errors"
	"log/slog"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	state      = "state-token-random"
	calendarID = "primary"
)

var ErrTimeSlotNotAvailable = errors.New("time slot not available")

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

func (g *GoogleCalendarAdapter) getClient(refreshToken string) (*calendar.Service, error) {
	ctx := context.Background()

	token := &oauth2.Token{RefreshToken: refreshToken}
	tokenSource := g.Config.TokenSource(ctx, token)
	client := oauth2.NewClient(ctx, tokenSource)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		slog.Error("failed to create calendar service", "error", err)
		return nil, err
	}

	return srv, nil
}

func (g *GoogleCalendarAdapter) ValidateAvailability(input outbound.ValidateAvailabilityInput) error {
	srv, err := g.getClient(input.RefreshToken)
	if err != nil {
		return err
	}

	loc, err := time.LoadLocation(input.Timezone)
	if err != nil {
		slog.Error("failed to load timezone", "error", err, "timezone", input.Timezone)
		return err
	}

	start, err := time.ParseInLocation("2006-01-02T15:04:05", input.StartTime, loc)
	if err != nil {
		slog.Error("failed to parse start time", "error", err, "start_time", input.StartTime)
		return err
	}

	end, err := time.ParseInLocation("2006-01-02T15:04:05", input.EndTime, loc)
	if err != nil {
		slog.Error("failed to parse end time", "error", err, "end_time", input.EndTime)
		return err
	}

	freeBusyReq := &calendar.FreeBusyRequest{
		TimeMin: start.Format(time.RFC3339),
		TimeMax: end.Format(time.RFC3339),
		Items:   []*calendar.FreeBusyRequestItem{{Id: calendarID}},
	}

	resp, err := srv.Freebusy.Query(freeBusyReq).Do()
	if err != nil {
		slog.Error("failed to query freebusy", "error", err)
		return err
	}

	calendarInfo, ok := resp.Calendars[calendarID]
	if ok && len(calendarInfo.Busy) > 0 {
		slog.Warn("time slot not available", "busy_periods", calendarInfo.Busy)
		return ErrTimeSlotNotAvailable
	}

	return nil
}

func (g *GoogleCalendarAdapter) CreateEvent(refreshToken string, input outbound.CalendarEventInput) error {
	srv, err := g.getClient(refreshToken)
	if err != nil {
		return err
	}

	event := &calendar.Event{
		Summary:     input.Title,
		Description: input.Description,
		Start: &calendar.EventDateTime{
			DateTime: input.StartTime,
			TimeZone: input.Timezone,
		},
		End: &calendar.EventDateTime{
			DateTime: input.EndTime,
			TimeZone: input.Timezone,
		},
	}

	if len(input.Attendees) > 0 {
		attendees := make([]*calendar.EventAttendee, len(input.Attendees))
		for i, email := range input.Attendees {
			attendees[i] = &calendar.EventAttendee{Email: email}
		}
		event.Attendees = attendees
	}

	_, err = srv.Events.Insert(calendarID, event).Do()
	if err != nil {
		slog.Error("failed to create calendar event", "error", err)
		return err
	}

	slog.Info("calendar event created", "title", input.Title, "start", input.StartTime)
	return nil
}

func (g *GoogleCalendarAdapter) DeleteEvent(refreshToken string, input outbound.DeleteEventInput) error {
	srv, err := g.getClient(refreshToken)
	if err != nil {
		return err
	}

	eventID, err := g.findEventByDateTime(srv, input.Date, input.Time, input.Timezone)
	if err != nil {
		return err
	}

	if err := srv.Events.Delete(calendarID, eventID).Do(); err != nil {
		slog.Error("failed to delete calendar event", "error", err, "event_id", eventID)
		return err
	}

	slog.Info("calendar event deleted", "event_id", eventID, "date", input.Date)
	return nil
}

func (g *GoogleCalendarAdapter) UpdateEvent(refreshToken string, input outbound.UpdateEventInput) error {
	srv, err := g.getClient(refreshToken)
	if err != nil {
		return err
	}

	eventID, err := g.findEventByDateTime(srv, input.OriginalDate, input.OriginalTime, input.Timezone)
	if err != nil {
		return err
	}

	event, err := srv.Events.Get(calendarID, eventID).Do()
	if err != nil {
		slog.Error("failed to get event for update", "error", err, "event_id", eventID)
		return err
	}

	loc, err := time.LoadLocation(input.Timezone)
	if err != nil {
		return err
	}

	newStart, err := time.ParseInLocation("2006-01-02T15:04", input.NewDate+"T"+input.NewTime, loc)
	if err != nil {
		newStart, _ = time.ParseInLocation("2006-01-02", input.NewDate, loc)
	}

	originalDuration := time.Hour
	if event.Start != nil && event.End != nil {
		startTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)
		endTime, _ := time.Parse(time.RFC3339, event.End.DateTime)
		originalDuration = endTime.Sub(startTime)
	}

	newEnd := newStart.Add(originalDuration)

	event.Start = &calendar.EventDateTime{
		DateTime: newStart.Format("2006-01-02T15:04:05"),
		TimeZone: input.Timezone,
	}
	event.End = &calendar.EventDateTime{
		DateTime: newEnd.Format("2006-01-02T15:04:05"),
		TimeZone: input.Timezone,
	}

	if _, err := srv.Events.Update(calendarID, eventID, event).Do(); err != nil {
		slog.Error("failed to update calendar event", "error", err, "event_id", eventID)
		return err
	}

	slog.Info("calendar event updated", "event_id", eventID, "new_date", input.NewDate)
	return nil
}

func (g *GoogleCalendarAdapter) findEventByDateTime(srv *calendar.Service, date, timeStr, timezone string) (string, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", err
	}

	var searchTime time.Time
	if timeStr != "" {
		searchTime, err = time.ParseInLocation("2006-01-02T15:04", date+"T"+timeStr, loc)
		if err != nil {
			searchTime, _ = time.ParseInLocation("2006-01-02", date, loc)
		}
	} else {
		searchTime, _ = time.ParseInLocation("2006-01-02", date, loc)
	}

	timeMin := searchTime.Add(-1 * time.Hour).Format(time.RFC3339)
	timeMax := searchTime.Add(24 * time.Hour).Format(time.RFC3339)

	events, err := srv.Events.List(calendarID).
		TimeMin(timeMin).
		TimeMax(timeMax).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		slog.Error("failed to list events", "error", err)
		return "", err
	}

	for _, event := range events.Items {
		if event.Start == nil {
			continue
		}
		eventStart, _ := time.Parse(time.RFC3339, event.Start.DateTime)
		if eventStart.In(loc).Format("2006-01-02") == date {
			if timeStr == "" || eventStart.In(loc).Format("15:04") == timeStr {
				return event.Id, nil
			}
		}
	}

	return "", errors.New("event not found")
}
