package ai

import (
	"app/internal/core/models"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiService struct {
	client *genai.Client
	model  string
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

func (s *GeminiService) GenerateResponse(history []models.Message, message string) (string, error) {
	ctx := context.Background()
	model := s.client.GenerativeModel(s.model)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text("You are a helpful assistant."),
		},
		Role: models.AssistantRole,
	}

	cs := model.StartChat()

	historyGemini := []*genai.Content{}

	for _, msg := range history {
		historyGemini = append(historyGemini, &genai.Content{
			Parts: []genai.Part{
				genai.Text(msg.Message),
			},
			Role: msg.Role,
		})
	}

	cs.History = historyGemini

	response, err := cs.SendMessage(ctx, genai.Text("Hello"))
	if err != nil {
		slog.Error("failed to generate response", "error", err)
		return "", err
	}

	var responseText strings.Builder
	for _, part := range response.Candidates[0].Content.Parts {
		responseText.WriteString(fmt.Sprint(part))
	}

	return responseText.String(), nil
}
