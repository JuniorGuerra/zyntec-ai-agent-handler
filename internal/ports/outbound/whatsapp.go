package outbound

import "app/internal/domain/models"

type WhatsAppPort interface {
	SendMessage(input models.SendMessageInput) error
}
