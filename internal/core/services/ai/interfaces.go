package ai

import "app/internal/core/models"

type AISession interface {
	GenerateResponse(message string) (string, error)
}

type AIModels interface {
	SetSystemInstruction(instruction string, history []models.Message) AISession
	SetModel(model string)
}
