package waha

import (
	"app/internal/domain/models"
	"app/internal/ports/outbound"
	"log/slog"
)

type Service struct {
	customerRepo outbound.CustomerRepository
	wahaAdapter  outbound.WhatsAppPort
}

func NewService(customerRepo outbound.CustomerRepository, wahaAdapter outbound.WhatsAppPort) *Service {
	return &Service{
		customerRepo: customerRepo,
		wahaAdapter:  wahaAdapter,
	}
}

func (s *Service) SendMessage(msg outbound.SQSMessageBody) error {
	customer, err := s.customerRepo.Get(msg.CustomerID)
	if err != nil {
		slog.Error("failed to get customer", "error", err, "customer_id", msg.CustomerID)
		return err
	}

	if !customer.IsValid() {
		slog.Error("customer is not valid", "customer_id", msg.CustomerID)
		return nil
	}

	switch msg.ActionType {
	case outbound.SQSActionLocation:
		return s.sendLocation(customer, msg)
	default:
		return s.sendText(customer, msg)
	}
}

func (s *Service) sendText(customer *models.Customer, msg outbound.SQSMessageBody) error {
	return s.wahaAdapter.SendMessage(models.SendMessageInput{
		APIKey:      customer.APIKey,
		PhoneNumber: msg.ChatID,
		Message:     msg.Message,
		URL:         customer.URL,
		SessionName: customer.SessionName,
	})
}

func (s *Service) sendLocation(customer *models.Customer, msg outbound.SQSMessageBody) error {
	if msg.Location == nil {
		slog.Warn("location is nil", "customer_id", msg.CustomerID)
		return nil
	}

	return s.wahaAdapter.SendLocation(models.SendLocationInput{
		APIKey:      customer.APIKey,
		ChatID:      msg.ChatID,
		Latitude:    msg.Location.Latitude,
		Longitude:   msg.Location.Longitude,
		Title:       msg.Location.Title,
		URL:         customer.URL,
		SessionName: customer.SessionName,
	})
}
