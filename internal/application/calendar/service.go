package calendar

import (
	"app/internal/adapters/outbound/persistence"
	"app/internal/domain/models"
	"app/internal/ports/outbound"
	"fmt"
	"log/slog"
	"time"
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

func (s *CalendarService) CancelAppointment(req models.CancelEventRequest) error {
	calendar, err := s.calendarRepo.GetByCustomerID(req.CustomerID)
	if err != nil {
		slog.Error("failed to get calendar", "error", err, "customer_id", req.CustomerID)
		return err
	}
	if calendar == nil {
		return fmt.Errorf("calendar not found for customer: %s", req.CustomerID)
	}

	if err := s.calendarPort.DeleteEvent(calendar.RefreshToken, outbound.DeleteEventInput{
		Date:     req.Date,
		Time:     req.Time,
		Timezone: calendar.Timezone,
	}); err != nil {
		slog.Error("failed to delete event", "error", err)
		return err
	}

	slog.Info("appointment cancelled", "customer_id", req.CustomerID, "date", req.Date, "reason", req.Reason)
	return nil
}

func (s *CalendarService) RescheduleAppointment(req models.RescheduleEventRequest) error {
	calendar, err := s.calendarRepo.GetByCustomerID(req.CustomerID)
	if err != nil {
		slog.Error("failed to get calendar", "error", err, "customer_id", req.CustomerID)
		return err
	}
	if calendar == nil {
		return fmt.Errorf("calendar not found for customer: %s", req.CustomerID)
	}

	newStartTime := fmt.Sprintf("%sT%s:00", req.NewDate, req.NewTime)
	newEndTime := fmt.Sprintf("%sT%s:00", req.NewDate, addOneHour(req.NewTime))

	if err := s.calendarPort.ValidateAvailability(outbound.ValidateAvailabilityInput{
		RefreshToken: calendar.RefreshToken,
		Timezone:     calendar.Timezone,
		StartTime:    newStartTime,
		EndTime:      newEndTime,
	}); err != nil {
		slog.Error("new time slot not available", "error", err)
		return err
	}

	if err := s.calendarPort.UpdateEvent(calendar.RefreshToken, outbound.UpdateEventInput{
		OriginalDate: req.OriginalDate,
		OriginalTime: req.OriginalTime,
		NewDate:      req.NewDate,
		NewTime:      req.NewTime,
		Timezone:     calendar.Timezone,
	}); err != nil {
		slog.Error("failed to update event", "error", err)
		return err
	}

	slog.Info("appointment rescheduled", "customer_id", req.CustomerID, "new_date", req.NewDate, "reason", req.Reason)
	return nil
}

func addOneHour(timeStr string) string {
	if timeStr == "" {
		return "10:00"
	}
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return "10:00"
	}
	return t.Add(time.Hour).Format("15:04")
}
