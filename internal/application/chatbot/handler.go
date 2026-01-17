package chatbot

import (
	"app/internal/domain/models"
	"app/internal/domain/utils"
	"app/internal/ports/outbound"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	sessionRepo  outbound.SessionRepository
	messageRepo  outbound.MessageRepository
	customerRepo outbound.CustomerRepository
	aiService    outbound.AIPort
}

func NewService(
	sessionRepo outbound.SessionRepository,
	messageRepo outbound.MessageRepository,
	customerRepo outbound.CustomerRepository,
	aiService outbound.AIPort,
) *Service {
	return &Service{
		sessionRepo:  sessionRepo,
		messageRepo:  messageRepo,
		customerRepo: customerRepo,
		aiService:    aiService,
	}
}

type ChatResponse struct {
	Message string           `json:"message"`
	Action  *outbound.Action `json:"action,omitempty"`
}

func (s *Service) ProcessMessage(businessID, customerPhone, messageBody string, fromMe bool) (*ChatResponse, error) {
	customer, err := s.customerRepo.Get(businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	if !customer.IsActive {
		return nil, fmt.Errorf("customer is not active")
	}

	session, err := s.getOrCreateSession(businessID, customerPhone)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create session: %w", err)
	}

	if s.isHumanAgentActive(session, customer.AgentTimeout, fromMe) {
		if err := s.messageRepo.Save(s.createMessage(session, models.CustomerRole, messageBody)); err != nil {
			return nil, fmt.Errorf("failed to save message: %w", err)
		}

		if fromMe {
			session.IsHumanAgent = true
		}
		session.UpdatedAt = time.Now()

		if err := s.sessionRepo.Save(*session); err != nil {
			return nil, fmt.Errorf("failed to save session: %w", err)
		}

		return &ChatResponse{Message: "Session is in human agent mode"}, nil
	}

	messages, err := s.messageRepo.GetHistory(session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	aiSession := s.aiService.CreateSession(customer.AIModel, customer.AIPrompt, messages)
	aiResponse, err := aiSession.GenerateResponse(messageBody)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI response: %w", err)
	}

	if aiResponse.HasAction() {
		if err := s.handleAction(session, aiResponse.Action); err != nil {
			slog.Error("failed to handle action", "error", err, "action", aiResponse.Action.Type)
		}
	}

	customerMsg := s.createMessage(session, models.CustomerRole, messageBody)
	assistantMsg := s.createMessage(session, models.AssistantRole, aiResponse.Message)
	if err := s.messageRepo.Save(customerMsg, assistantMsg); err != nil {
		return nil, fmt.Errorf("failed to save messages: %w", err)
	}

	return &ChatResponse{
		Message: aiResponse.Message,
		Action:  aiResponse.Action,
	}, nil
}

func (s *Service) getOrCreateSession(businessID, customerPhone string) (*models.Session, error) {
	sessionID := utils.GenerateSessionID(businessID, customerPhone)
	session, err := s.sessionRepo.Get(sessionID)
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
			BusinessPhoneNumber: businessID,
			CustomerPhoneNumber: customerPhone,
			IsHumanAgent:        false,
		}

		if err := s.sessionRepo.Save(*session); err != nil {
			slog.Error("Error saving session", "session", session, "error", err)
			return nil, err
		}
	}

	return session, nil
}

func (s *Service) isHumanAgentActive(session *models.Session, agentTimeout int, fromMe bool) bool {
	timeoutExpired := session.UpdatedAt.Add(time.Duration(agentTimeout) * time.Minute).Before(time.Now())

	if session.IsHumanAgent && timeoutExpired {
		session.IsHumanAgent = false
	}

	return fromMe || (session.IsHumanAgent && !timeoutExpired)
}

func (s *Service) createMessage(session *models.Session, role models.Role, message string) models.Message {
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

func (s *Service) handleAction(session *models.Session, action *outbound.Action) error {
	switch action.Type {
	case outbound.ActionTransferToHuman:
		session.IsHumanAgent = true
		session.UpdatedAt = time.Now()
		if err := s.sessionRepo.Save(*session); err != nil {
			return fmt.Errorf("failed to save session: %w", err)
		}
		slog.Info("transferred to human agent", "session_id", session.ID, "reason", action.Args["reason"])

	case outbound.ActionScheduleAppointment:
		// TODO: Enviar a SQS para procesamiento asíncrono
		slog.Info("appointment scheduled",
			"session_id", session.ID,
			"service", action.Args["service"],
			"date", action.Args["date"],
			"time", action.Args["time"],
			"notes", action.Args["notes"],
		)

	default:
		slog.Warn("unknown action type", "action", action.Type)
	}

	return nil
}
