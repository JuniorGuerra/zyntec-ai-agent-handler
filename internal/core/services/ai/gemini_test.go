package ai_test

import (
	"app/internal/core/models"
	"app/internal/core/services/ai"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateResponse(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	geminiService, err := ai.NewGeminiService(apiKey)
	require.NoError(t, err)

	geminiService.SetSystemInstruction("You are a helpful assistant.", []models.Message{})

	res, err := geminiService.GenerateResponse("Hello")
	require.NoError(t, err)
	require.NotEmpty(t, res)
}

func TestGenerateResponseWithHistory(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	geminiService, err := ai.NewGeminiService(apiKey)
	require.NoError(t, err)

	geminiService.SetSystemInstruction("You are a helpful assistant.", []models.Message{
		{
			Message: "Hello",
			Role:    models.CustomerRole,
		},
		{
			Message: "Hello! How can I help you today?",
			Role:    models.AssistantRole,
		},
	})

	res, err := geminiService.GenerateResponse("I need to know how is the weather today")
	fmt.Println(res)
	require.NoError(t, err)
	require.NotEmpty(t, res)
}
