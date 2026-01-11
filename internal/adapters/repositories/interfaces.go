package repositories

import "app/internal/core/models"

type Repository interface {
	SaveSession(session models.Session) error
	SaveMessage(message models.Message) error
}
