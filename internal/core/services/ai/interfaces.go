package services

type AIModels interface {
	GenerateResponse(sessionID string, message string) (string, error)
}
