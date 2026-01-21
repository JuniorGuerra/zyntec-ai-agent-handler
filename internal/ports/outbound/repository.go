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

type CalendarRepository interface {
	Save(calendar models.Calendar) error
	GetByCustomerID(customerID string) (*models.Calendar, error)
}

type CalendarAppointmentsRepository interface {
	Save(appointment models.CalendarEvent) error
	GetByCustomerPhoneNumber(customerPhoneNumber string) ([]models.CalendarEvent, error)
	Delete(customerPhoneNumber, eventID string) error
}
