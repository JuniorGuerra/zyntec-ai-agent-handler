package repositories

import "app/internal/core/models"

type Repository interface {
	SaveSession(session models.Session) error
	SaveMessages(messages ...models.Message) error
	GetSession(sessionID string) (*models.Session, error)
	GetCustomer(businessPhoneNumber string) (*models.Customer, error)
	GetMessageHistory(sessionID string) ([]models.Message, error)
}
