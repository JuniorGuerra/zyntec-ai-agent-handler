package ai

import (
	"app/internal/domain/models"
	"app/internal/ports/outbound"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiAdapter struct {
	client       *genai.Client
	defaultModel string
}

type geminiSession struct {
	model   string
	session *genai.ChatSession
}

var availableTools = &genai.Tool{
	FunctionDeclarations: []*genai.FunctionDeclaration{
		{
			Name:        string(outbound.ActionTransferToHuman),
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
			Name:        string(outbound.ActionScheduleAppointment),
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
					"email": {
						Type:        genai.TypeString,
						Description: "Correo electrónico del cliente para enviar confirmación (opcional)",
					},
				},
				Required: []string{"service"},
			},
		},
		{
			Name:        string(outbound.ActionCancelAppointment),
			Description: "Cancela una cita existente cuando el cliente solicita cancelarla",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"date": {
						Type:        genai.TypeString,
						Description: "Fecha de la cita a cancelar (formato: YYYY-MM-DD)",
					},
					"time": {
						Type:        genai.TypeString,
						Description: "Hora de la cita a cancelar (formato: HH:MM)",
					},
					"reason": {
						Type:        genai.TypeString,
						Description: "Motivo de la cancelación",
					},
				},
				Required: []string{"date"},
			},
		},
		{
			Name:        string(outbound.ActionRescheduleAppointment),
			Description: "Reprograma una cita existente cuando el cliente solicita cambiar la fecha u hora",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"original_date": {
						Type:        genai.TypeString,
						Description: "Fecha original de la cita (formato: YYYY-MM-DD)",
					},
					"original_time": {
						Type:        genai.TypeString,
						Description: "Hora original de la cita (formato: HH:MM)",
					},
					"new_date": {
						Type:        genai.TypeString,
						Description: "Nueva fecha para la cita (formato: YYYY-MM-DD)",
					},
					"new_time": {
						Type:        genai.TypeString,
						Description: "Nueva hora para la cita (formato: HH:MM)",
					},
					"reason": {
						Type:        genai.TypeString,
						Description: "Motivo del cambio",
					},
				},
				Required: []string{"new_date"},
			},
		},
		{
			Name:        string(outbound.ActionSendLocation),
			Description: "Envía la ubicación del negocio en Google Maps cuando el cliente confirma que quiere recibir la ubicación/mapa",
			Parameters: &genai.Schema{
				Type:       genai.TypeObject,
				Properties: map[string]*genai.Schema{},
			},
		},
	},
}

func NewGeminiAdapter(apiKey string) (*GeminiAdapter, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		slog.Error("failed to create Gemini client", "error", err)
		return nil, err
	}

	return &GeminiAdapter{
		client:       client,
		defaultModel: "gemini-2.0-flash-exp",
	}, nil
}

func (a *GeminiAdapter) CreateSession(modelName, instruction string, history []models.Message, extraContext outbound.ExtraContext) outbound.AISession {
	if modelName == "" {
		modelName = a.defaultModel
	}

	model := a.client.GenerativeModel(modelName)

	now := time.Now()
	dateContext := fmt.Sprintf(
		"[INFORMACIÓN DEL SISTEMA - FECHA ACTUAL]\nHoy es: %s\nFecha: %s\nHora: %s\nUsa esta información para interpretar referencias temporales como 'mañana', 'próxima semana', 'hoy', etc.\n\n",
		now.Format("Monday, 02 January 2006"),
		now.Format("2006-01-02"),
		now.Format("15:04"),
	)

	toolInstructions := `
[INSTRUCCIONES CRÍTICAS SOBRE FUNCIONES]
Tienes acceso a funciones para gestionar citas y ubicación. DEBES invocar estas funciones cuando el usuario confirme una acción:

- schedule_appointment: INVOCAR cuando el usuario CONFIRME agendar una cita (después de que diga "sí", "confirmo", "ok", "aja", etc.)
- cancel_appointment: INVOCAR cuando el usuario CONFIRME cancelar una cita
- reschedule_appointment: INVOCAR cuando el usuario CONFIRME reprogramar una cita
- transfer_to_human: INVOCAR cuando necesites transferir a un humano
- send_location: INVOCAR cuando el usuario quiera recibir la ubicación/mapa del negocio (después de que confirme con "sí", "ok", "envíamela", etc.)

IMPORTANTE:
1. NUNCA digas "ya agendé/cancelé/reprogramé" sin haber invocado la función correspondiente
2. Primero recopila la información necesaria (fecha, hora, servicio)
3. Confirma con el usuario los datos
4. Cuando el usuario confirme, INVOCA la función con los parámetros correctos
5. NO escribas código, NO uses print(), simplemente invoca la función directamente

UBICACIÓN:
- Cuando el usuario pregunte por la ubicación/dirección del negocio, responde con la dirección del contexto y pregunta si quiere recibir la ubicación en el mapa
- Cuando confirme, invoca send_location para enviar el pin de ubicación

`

	systemParts := []genai.Part{
		genai.Text(dateContext),
		genai.Text(toolInstructions),
		genai.Text(instruction),
	}

	if len(extraContext) > 0 {
		var contextBuilder strings.Builder
		contextBuilder.WriteString("\n\nContexto adicional del cliente:\n")
		for k, v := range extraContext {
			contextBuilder.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
		}
		systemParts = append(systemParts, genai.Text(contextBuilder.String()))
	}

	model.SystemInstruction = &genai.Content{Parts: systemParts}
	model.Tools = []*genai.Tool{availableTools}
	model.ToolConfig = &genai.ToolConfig{
		FunctionCallingConfig: &genai.FunctionCallingConfig{
			Mode: genai.FunctionCallingAuto,
		},
	}

	cs := model.StartChat()

	var historyGemini []*genai.Content
	for _, msg := range history {
		if !msg.Role.IsValidRole() {
			slog.Error("invalid role", "role", msg.Role)
			continue
		}

		historyGemini = append(historyGemini, &genai.Content{
			Parts: []genai.Part{genai.Text(msg.Message)},
			Role:  msg.Role.String(),
		})
	}

	cs.History = historyGemini

	return &geminiSession{
		model:   modelName,
		session: cs,
	}
}

func (gs *geminiSession) GenerateResponse(message string) (*outbound.AIResponse, error) {
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

	aiResponse := &outbound.AIResponse{}
	var responseText strings.Builder

	for _, part := range response.Candidates[0].Content.Parts {
		switch v := part.(type) {
		case genai.Text:
			responseText.WriteString(string(v))
		case genai.FunctionCall:
			aiResponse.Action = &outbound.Action{
				Type: outbound.ActionType(v.Name),
				Args: v.Args,
			}
			slog.Info("function call detected", "action", v.Name, "args", v.Args)
		}
	}

	aiResponse.Message = responseText.String()
	return aiResponse, nil
}
