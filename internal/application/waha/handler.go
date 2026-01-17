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

	err = s.wahaAdapter.SendMessage(models.SendMessageInput{
		APIKey:      customer.APIKey,
		PhoneNumber: customer.BusinessPhoneNumber,
		Message:     msg.Message,
		URL:         customer.URL,
		SessionName: customer.SessionName,
	})

	return err
}
