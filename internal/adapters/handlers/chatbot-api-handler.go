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
	"github.com/google/uuid"
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
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}

	if !req.Payload.From.IsValidFromType() {
		return events.APIGatewayProxyResponse{
			Body:       `{"error": "Invalid from type"}`,
			StatusCode: 400,
		}, nil
	}

	customer, err := h.repository.GetCustomer(req.Me.ID)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}

	if !customer.IsActive {
		return events.APIGatewayProxyResponse{
			Body:       `{"error": "Customer is not active"}`,
			StatusCode: 400,
		}, nil
	}

	session, err := h.getOrCreateSession(req.Me.ID, req.Payload.From.String())
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}

	if h.isHumanAgentActive(session, customer.AgentTimeout, req.Payload.FromMe) {
		err = h.repository.SaveMessages(h.addMessage(session, models.CustomerRole, req.Payload.Body))
		if err != nil {
			return events.APIGatewayProxyResponse{
				Body:       fmt.Sprintf(`{"error": "%v"}`, err),
				StatusCode: 500,
			}, nil
		}

		if req.Payload.FromMe {
			session.IsHumanAgent = true
		}
		session.UpdatedAt = time.Now()

		err = h.repository.SaveSession(*session)
		if err != nil {
			return events.APIGatewayProxyResponse{
				Body:       fmt.Sprintf(`{"error": "%v"}`, err),
				StatusCode: 500,
			}, nil
		}

		return events.APIGatewayProxyResponse{
			Body:       `{"message": "Session is in human agent mode"}`,
			StatusCode: 200,
		}, nil
	}

	messages, err := h.repository.GetMessageHistory(session.ID)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}

	aiSession := h.aiService.CreateSession(customer.AIModel, customer.AIPrompt, messages)

	aiResponse, err := aiSession.GenerateResponse(req.Payload.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}

	customerMsg := h.addMessage(session, models.CustomerRole, req.Payload.Body)
	assistantMsg := h.addMessage(session, models.AssistantRole, aiResponse)
	err = h.repository.SaveMessages(customerMsg, assistantMsg)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}

	// err = h.wahaService.SendWhatsAppMessage(session.CustomerPhoneNumber, aiResponse)
	// if err != nil {
	// 	return events.APIGatewayProxyResponse{
	// 		Body:       fmt.Sprintf(`{"error": "%v"}`, err),
	// 		StatusCode: 500,
	// 	}, nil
	// }

	resp := map[string]string{"message": aiResponse}
	jsonBody, _ := json.Marshal(resp)

	return events.APIGatewayProxyResponse{
		Body:       string(jsonBody),
		StatusCode: 200,
	}, nil
}

func (h *ChatbotAPIHandler) addMessage(session *models.Session, role models.Role, message string) models.Message {
	return models.Message{
		CreatedAt:           time.Now(),
		SessionID:           session.ID,
		ID:                  uuid.New().String(),
		Timestamp:           fmt.Sprintf("%d", time.Now().UnixNano()),
		Message:             message,
		BusinessPhoneNumber: session.BusinessPhoneNumber,
		CustomerPhoneNumber: session.CustomerPhoneNumber,
		Role:                role,
	}
}

func (h *ChatbotAPIHandler) getOrCreateSession(id, from string) (*models.Session, error) {
	sessionID := utils.GenerateSessionID(id, from)
	session, err := h.repository.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session == nil {
		now := time.Now()
		session = &models.Session{
			ID:                  sessionID,
			CreatedAt:           now,
			UpdatedAt:           now,
			SessionExpiryAt:     now.Add(1 * time.Hour),
			BusinessPhoneNumber: id,
			CustomerPhoneNumber: from,
			IsHumanAgent:        false,
		}

		err = h.repository.SaveSession(*session)
		if err != nil {
			slog.Error("Error saving session", "session", session, "error", err)
			return nil, err
		}
	}

	return session, nil
}

func (h *ChatbotAPIHandler) isHumanAgentActive(session *models.Session, agentTimeout int, fromMe bool) bool {
	timeoutExpired := session.UpdatedAt.Add(time.Duration(agentTimeout) * time.Minute).Before(time.Now())

	if session.IsHumanAgent && timeoutExpired {
		session.IsHumanAgent = false
	}

	return fromMe || (session.IsHumanAgent && !timeoutExpired)
}
