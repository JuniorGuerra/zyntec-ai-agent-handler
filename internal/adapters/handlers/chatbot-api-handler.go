package handlers

import (
	"app/internal/adapters/repositories"
	"app/internal/core/models"
	"app/internal/core/services/ai"
	whatsappsvc "app/internal/core/services/whatsapp-svc"
	"app/internal/core/utils"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

type ChatbotAPIHandler struct {
	repository  repositories.Repository
	aiService   ai.AIModels
	wahaService whatsappsvc.WhatsAppService
}

func NewChatbotAPIHandler(repository repositories.Repository, aiService ai.AIModels, wahaService whatsappsvc.WhatsAppService) *ChatbotAPIHandler {
	return &ChatbotAPIHandler{
		repository:  repository,
		aiService:   aiService,
		wahaService: wahaService,
	}
}

func (h *ChatbotAPIHandler) Handle(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	req := &models.WebhookRequest{}
	if err := json.Unmarshal([]byte(request.Body), req); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("Error unmarshalling request: %v", err),
			StatusCode: 500,
		}, nil
	}

	slog.Info("Request", "request", req)

	sessionID := utils.GenerateSessionID(req.Me.ID, req.Payload.From)
	session, err := h.repository.GetSession(sessionID) // TODO: or participant
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("Error getting session: %v", err),
			StatusCode: 500,
		}, nil
	}

	slog.Info("Session found", "session", session)

	if session == nil {
		session = &models.Session{
			ID:                  sessionID,
			CreatedAt:           time.Now(),
			SessionExpiryAt:     time.Now().Add(1 * time.Hour),
			BusinessPhoneNumber: req.Me.ID,
			CustomerPhoneNumber: req.Payload.From,
		}

		err = h.repository.SaveSession(*session)
		if err != nil {
			return events.APIGatewayProxyResponse{
				Body:       fmt.Sprintf("Error saving session: %v", err),
				StatusCode: 500,
			}, nil
		}
	}

	err = h.repository.SaveMessage(models.Message{
		ID:                  request.RequestContext.RequestID,
		CreatedAt:           time.Now(),
		SessionID:           session.ID,
		Timestamp:           req.Payload.Timestamp,
		Message:             req.Payload.Body,
		BusinessPhoneNumber: session.BusinessPhoneNumber,
		CustomerPhoneNumber: session.CustomerPhoneNumber,
		Role:                models.CustomerRole,
	})

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("Error saving message: %v", err),
			StatusCode: 500,
		}, nil
	}

	messages, err := h.repository.GetMessageHistory(session.ID)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("Error getting message history: %v", err),
			StatusCode: 500,
		}, nil
	}

	aiResponse, err := h.aiService.GenerateResponse(messages, req.Payload.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("Error generating AI response: %v", err),
			StatusCode: 500,
		}, nil
	}

	err = h.repository.SaveMessage(models.Message{
		ID:                  request.RequestContext.RequestID,
		CreatedAt:           time.Now(),
		SessionID:           session.ID,
		Timestamp:           req.Payload.Timestamp,
		Message:             aiResponse,
		BusinessPhoneNumber: session.BusinessPhoneNumber,
		CustomerPhoneNumber: session.CustomerPhoneNumber,
		Role:                models.AssistantRole,
	})

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("Error saving AI message: %v", err),
			StatusCode: 500,
		}, nil
	}

	// err = h.wahaService.SendWhatsAppMessage(session.CustomerPhoneNumber, aiResponse)
	// if err != nil {
	// 	return events.APIGatewayProxyResponse{
	// 		Body:       fmt.Sprintf("Error sending WhatsApp message: %v", err),
	// 		StatusCode: 500,
	// 	}, nil
	// }

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Message saved successfully: %v", request.Body),
		StatusCode: 200,
	}, nil
}
