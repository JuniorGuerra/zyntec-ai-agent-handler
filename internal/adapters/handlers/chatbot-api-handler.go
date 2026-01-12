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
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

type ChatbotAPIHandler struct {
	repository  repositories.Repository
	aiService   ai.AIModels
	wahaService whatsappsvc.WhatsAppService

	session  *models.Session
	messages []models.Message
	mu       *sync.Mutex
}

func NewChatbotAPIHandler(repository repositories.Repository, aiService ai.AIModels, wahaService whatsappsvc.WhatsAppService) *ChatbotAPIHandler {
	return &ChatbotAPIHandler{
		repository:  repository,
		aiService:   aiService,
		wahaService: wahaService,
		mu:          &sync.Mutex{},
	}
}

func (h *ChatbotAPIHandler) Handle(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	req := &models.WebhookRequest{}

	if err := json.Unmarshal([]byte(request.Body), req); err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}

	slog.Info("Request", "request", req)

	start := time.Now()
	customer, err := h.repository.GetCustomer(req.Me.ID)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}
	slog.Info("Customer", "customer", customer, "duration", time.Since(start))

	start = time.Now()
	err = h.getOrCreateSession(req.Me.ID, req.Payload.From)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}
	slog.Info("Session", "session", h.session, "duration", time.Since(start))

	start = time.Now()
	go h.addMessageToSession(models.CustomerRole, req.Payload.Body)
	slog.Info("Add message to session", "duration", time.Since(start))

	start = time.Now()
	messages, err := h.repository.GetMessageHistory(h.session.ID)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}
	slog.Info("Get message history", "messages", messages, "duration", time.Since(start))

	start = time.Now()
	h.aiService.SetModel(customer.AIModel)
	h.aiService.SetSystemInstruction(customer.AIPrompt, messages)
	slog.Info("Set model and system instruction", "duration", time.Since(start))

	start = time.Now()
	aiResponse, err := h.aiService.GenerateResponse(req.Payload.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}
	slog.Info("Generate response", "ai_response", aiResponse, "duration", time.Since(start))

	start = time.Now()
	h.addMessageToSession(models.AssistantRole, aiResponse)
	err = h.saveMessages()
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}
	slog.Info("Save messages", "duration", time.Since(start))

	// err = h.wahaService.SendWhatsAppMessage(h.session.CustomerPhoneNumber, aiResponse)
	// if err != nil {
	// 	return events.APIGatewayProxyResponse{
	// 		Body:       fmt.Sprintf(`{"error": "%v"}`, err),
	// 		StatusCode: 500,
	// 	}, nil
	// }

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf(`{"message": "%v"}`, aiResponse),
		StatusCode: 200,
	}, nil
}

func (h *ChatbotAPIHandler) addMessageToSession(role models.Role, message string) {
	if !role.IsValidRole() {
		slog.Error("Invalid role", "role", role)
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.messages = append(h.messages, models.Message{
		CreatedAt:           time.Now(),
		SessionID:           h.session.ID,
		ID:                  uuid.New().String(),
		Timestamp:           fmt.Sprintf("%d", time.Now().UnixNano()),
		Message:             message,
		BusinessPhoneNumber: h.session.BusinessPhoneNumber,
		CustomerPhoneNumber: h.session.CustomerPhoneNumber,
		Role:                role,
	})
}

func (h *ChatbotAPIHandler) getOrCreateSession(id, from string) error {
	sessionID := utils.GenerateSessionID(id, from)
	session, err := h.repository.GetSession(sessionID) // TODO: or participant
	if err != nil {
		return err
	}

	if session == nil {
		session = &models.Session{
			ID:                  sessionID,
			CreatedAt:           time.Now(),
			SessionExpiryAt:     time.Now().Add(1 * time.Hour),
			BusinessPhoneNumber: id,
			CustomerPhoneNumber: from,
		}

		err = h.repository.SaveSession(*session)
		if err != nil {
			return err
		}
	}

	h.session = session
	return nil
}

func (h *ChatbotAPIHandler) saveMessages() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.messages) == 0 {
		return nil
	}

	err := h.repository.SaveMessages(h.messages...)
	if err != nil {
		slog.Error("Error saving messages", "count", len(h.messages), "error", err)
		return err
	}

	return nil
}
