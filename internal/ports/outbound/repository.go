package outbound

import "app/internal/domain/models"

type SessionRepository interface {
	Save(session models.Session) error
	Get(sessionID string) (*models.Session, error)
}

type MessageRepository interface {
	Save(messages ...models.Message) error
	GetHistory(sessionID string) ([]models.Message, error)
}

type CustomerRepository interface {
	Get(businessPhoneNumber string) (*models.Customer, error)
}
