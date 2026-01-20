package chatbot

import (
	"app/internal/adapters/outbound/calendar"
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
	calendarRepo outbound.CalendarRepository
	calendarPort outbound.CalendarPort
	aiService    outbound.AIPort
}

func NewService(
	sessionRepo outbound.SessionRepository,
	messageRepo outbound.MessageRepository,
	customerRepo outbound.CustomerRepository,
	calendarRepo outbound.CalendarRepository,
	calendarPort outbound.CalendarPort,
	aiService outbound.AIPort,
) *Service {
	return &Service{
		sessionRepo:  sessionRepo,
		messageRepo:  messageRepo,
		customerRepo: customerRepo,
		calendarPort: calendarPort,
		calendarRepo: calendarRepo,
		aiService:    aiService,
	}
}

type ChatResponse struct {
	Message        string `json:"message"`
	CalendarAction any    `json:"calendar_action,omitempty"`
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

	response := &ChatResponse{
		Message: aiResponse.Message,
	}

	if aiResponse.HasAction() {
		result, err := s.handleAction(session, aiResponse.Action)
		if err == calendar.ErrTimeSlotNotAvailable {
			response.Message = "Lo siento, el horario seleccionado no está disponible. Por favor intenta con otro horario."
		} else if err != nil {
			slog.Warn("action validation failed", "error", err, "action", aiResponse.Action.Type)
			response.Message = fmt.Sprintf("Lo siento, no pude completar la acción: %s. Por favor intenta con otro horario.", err.Error())
		} else if result != nil {
			response.CalendarAction = result
		}
	}

	customerMsg := s.createMessage(session, models.CustomerRole, messageBody)
	assistantMsg := s.createMessage(session, models.AssistantRole, response.Message)
	if err := s.messageRepo.Save(customerMsg, assistantMsg); err != nil {
		return nil, fmt.Errorf("failed to save messages: %w", err)
	}

	return response, nil
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

func (s *Service) handleAction(session *models.Session, action *outbound.Action) (any, error) {
	switch action.Type {
	case outbound.ActionTransferToHuman:
		return s.handleTransferToHuman(session, action)

	case outbound.ActionScheduleAppointment:
		return s.handleScheduleAppointment(session, action)

	case outbound.ActionCancelAppointment:
		return s.handleCancelAppointment(session, action)

	case outbound.ActionRescheduleAppointment:
		return s.handleRescheduleAppointment(session, action)

	default:
		slog.Warn("unknown action type", "action", action.Type)
		return nil, nil
	}
}

func (s *Service) handleTransferToHuman(session *models.Session, action *outbound.Action) (any, error) {
	session.IsHumanAgent = true
	session.UpdatedAt = time.Now()
	if err := s.sessionRepo.Save(*session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}
	slog.Info("transferred to human agent", "session_id", session.ID, "reason", action.Args["reason"])
	return nil, nil
}

func (s *Service) handleScheduleAppointment(session *models.Session, action *outbound.Action) (*models.CalendarEventRequest, error) {
	calendar, err := s.calendarRepo.GetByCustomerID(session.BusinessPhoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar: %w", err)
	}

	startTime := buildStartTime(action.Args, "date", "time")
	endTime := buildEndTime(action.Args, "date", "time")

	if err := s.calendarPort.ValidateAvailability(outbound.ValidateAvailabilityInput{
		RefreshToken: calendar.RefreshToken,
		Timezone:     calendar.Timezone,
		StartTime:    startTime,
		EndTime:      endTime,
	}); err != nil {
		return nil, err
	}

	return &models.CalendarEventRequest{
		Action:      models.CalendarActionSchedule,
		CustomerID:  session.BusinessPhoneNumber,
		Title:       getStringArg(action.Args, "service", "Cita"),
		Description: getStringArg(action.Args, "notes", ""),
		StartTime:   startTime,
		EndTime:     endTime,
	}, nil
}

func (s *Service) handleCancelAppointment(session *models.Session, action *outbound.Action) (*models.CancelEventRequest, error) {
	return &models.CancelEventRequest{
		Action:     models.CalendarActionCancel,
		CustomerID: session.BusinessPhoneNumber,
		Date:       getStringArg(action.Args, "date", ""),
		Time:       getStringArg(action.Args, "time", ""),
		Reason:     getStringArg(action.Args, "reason", ""),
	}, nil
}

func (s *Service) handleRescheduleAppointment(session *models.Session, action *outbound.Action) (*models.RescheduleEventRequest, error) {
	calendar, err := s.calendarRepo.GetByCustomerID(session.BusinessPhoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar: %w", err)
	}

	newStartTime := buildStartTime(action.Args, "new_date", "new_time")
	newEndTime := buildEndTime(action.Args, "new_date", "new_time")

	if err := s.calendarPort.ValidateAvailability(outbound.ValidateAvailabilityInput{
		RefreshToken: calendar.RefreshToken,
		Timezone:     calendar.Timezone,
		StartTime:    newStartTime,
		EndTime:      newEndTime,
	}); err != nil {
		return nil, err
	}

	return &models.RescheduleEventRequest{
		Action:       models.CalendarActionReschedule,
		CustomerID:   session.BusinessPhoneNumber,
		OriginalDate: getStringArg(action.Args, "original_date", ""),
		OriginalTime: getStringArg(action.Args, "original_time", ""),
		NewDate:      getStringArg(action.Args, "new_date", ""),
		NewTime:      getStringArg(action.Args, "new_time", ""),
		Reason:       getStringArg(action.Args, "reason", ""),
	}, nil
}
