package sqs

import (
	"app/internal/application/waha"
	"app/internal/ports/outbound"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
)

type WhatsAppHandler struct {
	wahaHandler *waha.Service
}

func NewWhatsAppHandler(wahaHandler *waha.Service) *WhatsAppHandler {
	return &WhatsAppHandler{
		wahaHandler: wahaHandler,
	}
}

func (h *WhatsAppHandler) HandleSQSMessage(event events.SQSEvent) error {
	for _, record := range event.Records {
		if err := h.ProcessMessage(record); err != nil {
			slog.Error(
				"Failed to process message",
				"error", err,
				"message", record.Body,
				"message_id", record.MessageId,
			)
		}
	}

	return nil
}

func (h *WhatsAppHandler) ProcessMessage(message events.SQSMessage) error {
	msg := outbound.SQSMessageBody{}
	if err := json.Unmarshal([]byte(message.Body), &msg); err != nil {
		return err
	}

	if err := h.wahaHandler.SendMessage(msg); err != nil {
		slog.Error("failed to send message", "error", err)
		return err
	}

	return nil
}
