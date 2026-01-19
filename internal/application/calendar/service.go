package calendar

import (
	"app/internal/adapters/outbound/persistence"
	"app/internal/domain/models"
	"app/internal/ports/outbound"
	"fmt"
	"log/slog"
)

type CalendarService struct {
	calendarPort outbound.CalendarPort
	calendarRepo *persistence.CalendarRepository
}

func NewCalendarService(
	calendarPort outbound.CalendarPort,
	calendarRepo *persistence.CalendarRepository,
) *CalendarService {
	return &CalendarService{
		calendarPort: calendarPort,
		calendarRepo: calendarRepo,
	}
}

func (s *CalendarService) ScheduleAppointment(req models.CalendarEventRequest) error {
	calendar, err := s.calendarRepo.GetByCustomerID(req.CustomerID)
	if err != nil {
		slog.Error("failed to get calendar", "error", err, "customer_id", req.CustomerID)
		return err
	}
	if calendar == nil {
		return fmt.Errorf("calendar not found for customer: %s", req.CustomerID)
	}

	if err := s.calendarPort.ValidateAvailability(
		outbound.ValidateAvailabilityInput{
			RefreshToken: calendar.RefreshToken,
			Timezone:     calendar.Timezone,
			StartTime:    req.StartTime,
			EndTime:      req.EndTime,
		},
	); err != nil {
		slog.Error("availability validation failed", "error", err)
		return err
	}

	eventInput := outbound.CalendarEventInput{
		Title:       req.Title,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Timezone:    calendar.Timezone,
		Attendees:   req.Attendees,
	}

	if err := s.calendarPort.CreateEvent(calendar.RefreshToken, eventInput); err != nil {
		slog.Error("failed to create event", "error", err)
		return err
	}

	slog.Info("appointment scheduled", "customer_id", req.CustomerID, "title", req.Title)
	return nil
}
