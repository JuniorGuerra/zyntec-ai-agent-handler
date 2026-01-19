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
	calendarSQSUrl string
}

func NewChatbotHandler(service *chatbot.Service, sqsAdapter outbound.SQSAdapter, whatsappSQSUrl, calendarSQSUrl string) *ChatbotHandler {
	return &ChatbotHandler{
		service:        service,
		sqsAdapter:     sqsAdapter,
		whatsappSQSUrl: whatsappSQSUrl,
		calendarSQSUrl: calendarSQSUrl,
	}
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

	if err := h.ProcessAction(req.Me.ID, response.Action); err != nil {
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

func (h *ChatbotHandler) ProcessAction(customerID string, action *outbound.Action) error {
	if action == nil {
		return nil
	}

	switch action.Type {
	case outbound.ActionScheduleAppointment:
		return h.sendCalendarEvent(customerID, action)
	case outbound.ActionCancelAppointment:
		return h.sendCancelEvent(customerID, action)
	case outbound.ActionRescheduleAppointment:
		return h.sendRescheduleEvent(customerID, action)
	default:
		return nil
	}
}

func (h *ChatbotHandler) sendCalendarEvent(customerID string, action *outbound.Action) error {
	calendarReq := models.CalendarEventRequest{
		Action:      models.CalendarActionSchedule,
		CustomerID:  customerID,
		Title:       getStringArg(action.Args, "service", "Cita"),
		Description: getStringArg(action.Args, "notes", ""),
		StartTime:   buildStartTime(action.Args, "date", "time"),
		EndTime:     buildEndTime(action.Args, "date", "time"),
	}

	return h.sqsAdapter.SendMessage(
		context.Background(),
		outbound.SQSMessage{
			QueueURL: h.calendarSQSUrl,
			Body:     calendarReq,
		},
	)
}

func (h *ChatbotHandler) sendCancelEvent(customerID string, action *outbound.Action) error {
	cancelReq := models.CancelEventRequest{
		Action:     models.CalendarActionCancel,
		CustomerID: customerID,
		Date:       getStringArg(action.Args, "date", ""),
		Time:       getStringArg(action.Args, "time", ""),
		Reason:     getStringArg(action.Args, "reason", ""),
	}

	return h.sqsAdapter.SendMessage(
		context.Background(),
		outbound.SQSMessage{
			QueueURL: h.calendarSQSUrl,
			Body:     cancelReq,
		},
	)
}

func (h *ChatbotHandler) sendRescheduleEvent(customerID string, action *outbound.Action) error {
	rescheduleReq := models.RescheduleEventRequest{
		Action:       models.CalendarActionReschedule,
		CustomerID:   customerID,
		OriginalDate: getStringArg(action.Args, "original_date", ""),
		OriginalTime: getStringArg(action.Args, "original_time", ""),
		NewDate:      getStringArg(action.Args, "new_date", ""),
		NewTime:      getStringArg(action.Args, "new_time", ""),
		Reason:       getStringArg(action.Args, "reason", ""),
	}

	return h.sqsAdapter.SendMessage(
		context.Background(),
		outbound.SQSMessage{
			QueueURL: h.calendarSQSUrl,
			Body:     rescheduleReq,
		},
	)
}

func getStringArg(args map[string]any, key, defaultVal string) string {
	if val, ok := args[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultVal
}

func buildStartTime(args map[string]any, dateKey, timeKey string) string {
	date := getStringArg(args, dateKey, "")
	timeStr := getStringArg(args, timeKey, "")
	if date != "" && timeStr != "" {
		return fmt.Sprintf("%sT%s:00", date, timeStr)
	}
	return ""
}

func buildEndTime(args map[string]any, dateKey, timeKey string) string {
	startTime := buildStartTime(args, dateKey, timeKey)
	if startTime == "" {
		return ""
	}
	return fmt.Sprintf("%s", startTime[:11]+"23:59:00")
}
