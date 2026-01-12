package ai

import "app/internal/core/models"

type AIModels interface {
	GenerateResponse(message string) (string, error)
	SetSystemInstruction(instruction string, history []models.Message)
	SetModel(model string)
}
