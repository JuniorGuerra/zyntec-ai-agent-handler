package ai

import "app/internal/core/models"

type AIModels interface {
	GenerateResponse(history []models.Message, message string) (string, error)
}
