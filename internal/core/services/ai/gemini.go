package ai

import (
	"app/internal/core/models"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type ActionType string

const (
	ActionTransferToHuman    ActionType = "transfer_to_human"
	ActionScheduleAppointment ActionType = "schedule_appointment"
)

type Action struct {
	Type   ActionType     `json:"type"`
	Args   map[string]any `json:"args,omitempty"`
}

type AIResponse struct {
	Message string  `json:"message,omitempty"`
	Action  *Action `json:"action,omitempty"`
}

func (r *AIResponse) HasAction() bool {
	return r.Action != nil
}

type GeminiService struct {
	client       *genai.Client
	defaultModel string
}

type GeminiSession struct {
	model   string
	session *genai.ChatSession
}

var availableTools = &genai.Tool{
	FunctionDeclarations: []*genai.FunctionDeclaration{
		{
			Name:        string(ActionTransferToHuman),
			Description: "Transfiere la conversación a un agente humano cuando el cliente lo solicita explícitamente, cuando el bot no puede resolver el problema, o cuando la situación requiere atención humana",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"reason": {
						Type:        genai.TypeString,
						Description: "Razón por la cual se transfiere a un agente humano",
					},
				},
				Required: []string{"reason"},
			},
		},
		{
			Name:        string(ActionScheduleAppointment),
			Description: "Agenda una cita para el cliente cuando solicita programar una visita, reunión o servicio",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"date": {
						Type:        genai.TypeString,
						Description: "Fecha deseada para la cita (formato: YYYY-MM-DD)",
					},
					"time": {
						Type:        genai.TypeString,
						Description: "Hora deseada para la cita (formato: HH:MM)",
					},
					"service": {
						Type:        genai.TypeString,
						Description: "Tipo de servicio o motivo de la cita",
					},
					"notes": {
						Type:        genai.TypeString,
						Description: "Notas adicionales del cliente",
					},
				},
				Required: []string{"service"},
			},
		},
	},
}

func NewGeminiService(apiKey string) (AIModels, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		slog.Error("failed to create Gemini client", "error", err)
		return nil, err
	}

	return &GeminiService{
		client:       client,
		defaultModel: "gemini-2.0-flash-exp",
	}, nil
}

func (s *GeminiService) CreateSession(modelName, instruction string, history []models.Message) AISession {
	if modelName == "" {
		modelName = s.defaultModel
	}

	model := s.client.GenerativeModel(modelName)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text(instruction),
		},
	}
	model.Tools = []*genai.Tool{availableTools}

	cs := model.StartChat()

	historyGemini := []*genai.Content{}

	for _, msg := range history {
		if !msg.Role.IsValidRole() {
			slog.Error("invalid role", "role", msg.Role)
			continue
		}

		historyGemini = append(historyGemini, &genai.Content{
			Parts: []genai.Part{
				genai.Text(msg.Message),
			},
			Role: msg.Role.String(),
		})
	}

	cs.History = historyGemini

	return &GeminiSession{
		model:   modelName,
		session: cs,
	}
}

func (gs *GeminiSession) GenerateResponse(message string) (*AIResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if gs.session == nil {
		slog.Error("session is nil")
		return nil, fmt.Errorf("session is nil")
	}

	response, err := gs.session.SendMessage(ctx, genai.Text(message))
	if err != nil {
		slog.Error("failed to generate response", "error", err, "model", gs.model, "message", message)
		return nil, err
	}

	if len(response.Candidates) == 0 || response.Candidates[0].Content == nil {
		slog.Error("no candidates in response", "model", gs.model)
		return nil, fmt.Errorf("no response from AI model")
	}

	aiResponse := &AIResponse{}
	var responseText strings.Builder

	for _, part := range response.Candidates[0].Content.Parts {
		switch v := part.(type) {
		case genai.Text:
			responseText.WriteString(string(v))
		case genai.FunctionCall:
			aiResponse.Action = &Action{
				Type: ActionType(v.Name),
				Args: v.Args,
			}
			slog.Info("function call detected", "action", v.Name, "args", v.Args)
		}
	}

	aiResponse.Message = responseText.String()
	return aiResponse, nil
}
