package sqs

import (
	"app/internal/ports/outbound"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
)

type WhatsAppHandler struct {
	sqsAdapter outbound.SQSAdapter
}

func NewWhatsAppHandler(sqsAdapter outbound.SQSAdapter) *WhatsAppHandler {
	return &WhatsAppHandler{
		sqsAdapter: sqsAdapter,
	}
}

func (h *WhatsAppHandler) HandleMessage(event events.SQSEvent) error {
	for _, record := range event.Records {
		slog.Info("Received message", "message", record.Body)
	}

	return nil
}
