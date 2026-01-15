package ai

import "app/internal/core/models"

type AISession interface {
	GenerateResponse(message string) (*AIResponse, error)
}

type AIModels interface {
	CreateSession(model, instruction string, history []models.Message) AISession
}
