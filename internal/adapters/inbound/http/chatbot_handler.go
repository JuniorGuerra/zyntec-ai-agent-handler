package http

import (
	"app/internal/application/chatbot"
	"app/internal/domain/models"
	"app/internal/ports/outbound"
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

type ChatbotHandler struct {
	service        *chatbot.Service
	sqsAdapter     outbound.SQSAdapter
	whatsappSQSUrl string
}

func NewChatbotHandler(service *chatbot.Service, sqsAdapter outbound.SQSAdapter, whatsappSQSUrl string) *ChatbotHandler {
	return &ChatbotHandler{service: service, sqsAdapter: sqsAdapter, whatsappSQSUrl: whatsappSQSUrl}
}

func (h *ChatbotHandler) HandleWebhook(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	response, err := h.service.ProcessMessage(
		req.Me.ID,
		req.Payload.From.String(),
		req.Payload.Body,
		req.Payload.FromMe,
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}

	// indepent of the response, we need to send the message to SQS
	err = h.sqsAdapter.SendMessage(
		context.Background(),
		outbound.SQSMessage{
			QueueURL: h.whatsappSQSUrl,
			Body: outbound.SQSMessageBody{
				CustomerID: req.Me.ID,
				Message:    response.Message,
			},
		},
	)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}

	jsonBody, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       string(jsonBody),
		StatusCode: 200,
	}, nil
}
