package sqs

import (
	"app/internal/application/calendar"
	"app/internal/domain/models"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
)

type CalendarHandler struct {
	service *calendar.CalendarService
}

func NewCalendarHandler(service *calendar.CalendarService) *CalendarHandler {
	return &CalendarHandler{
		service: service,
	}
}

func (h *CalendarHandler) HandleSQSMessage(event events.SQSEvent) error {
	for _, record := range event.Records {
		if err := h.processMessage(record); err != nil {
			slog.Error(
				"failed to process calendar message",
				"error", err,
				"message", record.Body,
				"message_id", record.MessageId,
			)
		}
	}

	return nil
}

type calendarAction struct {
	Action models.CalendarActionType `json:"action"`
}

func (h *CalendarHandler) processMessage(message events.SQSMessage) error {
	var action calendarAction
	if err := json.Unmarshal([]byte(message.Body), &action); err != nil {
		slog.Error("failed to unmarshal action type", "error", err)
		return err
	}

	switch action.Action {
	case models.CalendarActionSchedule:
		return h.handleSchedule(message.Body)
	case models.CalendarActionCancel:
		return h.handleCancel(message.Body)
	case models.CalendarActionReschedule:
		return h.handleReschedule(message.Body)
	default:
		return fmt.Errorf("unknown calendar action: %s", action.Action)
	}
}

func (h *CalendarHandler) handleSchedule(body string) error {
	var req models.CalendarEventRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		slog.Error("failed to unmarshal schedule request", "error", err)
		return err
	}
	return h.service.ScheduleAppointment(req)
}

func (h *CalendarHandler) handleCancel(body string) error {
	var req models.CancelEventRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		slog.Error("failed to unmarshal cancel request", "error", err)
		return err
	}
	return h.service.CancelAppointment(req)
}

func (h *CalendarHandler) handleReschedule(body string) error {
	var req models.RescheduleEventRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		slog.Error("failed to unmarshal reschedule request", "error", err)
		return err
	}
	return h.service.RescheduleAppointment(req)
}
