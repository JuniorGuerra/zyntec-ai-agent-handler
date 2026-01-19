package sqs

import (
	"app/internal/application/calendar"
	"app/internal/domain/models"
	"encoding/json"
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
		if err := h.ProcessMessage(record); err != nil {
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

func (h *CalendarHandler) ProcessMessage(message events.SQSMessage) error {
	var req models.CalendarEventRequest
	if err := json.Unmarshal([]byte(message.Body), &req); err != nil {
		slog.Error("failed to unmarshal calendar event request", "error", err)
		return err
	}

	if err := h.service.ScheduleAppointment(req); err != nil {
		return err
	}

	return nil
}
