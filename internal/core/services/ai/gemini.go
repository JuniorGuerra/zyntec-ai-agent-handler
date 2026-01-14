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

type GeminiService struct {
	client *genai.Client
	model  string
}

type GeminiSession struct {
	model   string
	session *genai.ChatSession
}

func NewGeminiService(apiKey string) (AIModels, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		slog.Error("failed to create Gemini client", "error", err)
		return nil, err
	}

	return &GeminiService{
		client: client,
		model:  "gemini-2.0-flash-exp",
	}, nil
}

func (s *GeminiService) SetModel(model string) {
	if model == "" {
		return
	}

	s.model = model
}

func (s *GeminiService) SetSystemInstruction(instruction string, history []models.Message) AISession {
	model := s.client.GenerativeModel(s.model)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text(instruction),
		},
	}

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
		model:   s.model,
		session: cs,
	}
}

func (gs *GeminiSession) GenerateResponse(message string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if gs.session == nil {
		slog.Error("session is nil")
		return "", fmt.Errorf("session is nil")
	}

	response, err := gs.session.SendMessage(ctx, genai.Text(message))
	if err != nil {
		slog.Error("failed to generate response", "error", err, "model", gs.model, "message", message)
		return "", err
	}

	var responseText strings.Builder
	for _, part := range response.Candidates[0].Content.Parts {
		responseText.WriteString(fmt.Sprint(part))
	}

	return responseText.String(), nil
}
