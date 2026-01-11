package handlers

import (
	"app/internal/adapters/repositories"
	"app/internal/core/models"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

type ChatbotAPIHandler struct {
	repository repositories.Repository
}

func NewChatbotAPIHandler(repository repositories.Repository) *ChatbotAPIHandler {
	return &ChatbotAPIHandler{
		repository: repository,
	}
}

func (h *ChatbotAPIHandler) Handle(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	slog.Info("Saving message", "request", request)

	err := h.repository.SaveSession(models.Session{
		ID:              request.RequestContext.RequestID,
		CreatedAt:       time.Now(),
		SessionExpiryAt: time.Now().Add(1 * time.Hour),
	})

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("Error saving session: %v", err),
			StatusCode: 500,
		}, nil
	}

	err = h.repository.SaveMessage(models.Message{
		ID:        request.RequestContext.RequestID,
		CreatedAt: time.Now(),
		SessionID: request.RequestContext.RequestID,
		Body:      request.Body,
	})

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("Error saving message: %v", err),
			StatusCode: 500,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Message saved successfully: %v", request.Body),
		StatusCode: 200,
	}, nil
}
