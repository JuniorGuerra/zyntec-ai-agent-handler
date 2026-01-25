package outbound

import "app/internal/domain/models"

type ActionType string

const (
	ActionTransferToHuman       ActionType = "transfer_to_human"
	ActionScheduleAppointment   ActionType = "schedule_appointment"
	ActionCancelAppointment     ActionType = "cancel_appointment"
	ActionRescheduleAppointment ActionType = "reschedule_appointment"
	ActionSendLocation          ActionType = "send_location"
)

type Action struct {
	Type ActionType     `json:"type"`
	Args map[string]any `json:"args,omitempty"`
}

type AIResponse struct {
	Message string  `json:"message,omitempty"`
	Action  *Action `json:"action,omitempty"`
}

func (r *AIResponse) HasAction() bool {
	return r.Action != nil
}

type AISession interface {
	GenerateResponse(message string) (*AIResponse, error)
}

type ExtraContext map[string]any

type AIPort interface {
	CreateSession(model, instruction string, history []models.Message, extraContext ExtraContext) AISession
}
