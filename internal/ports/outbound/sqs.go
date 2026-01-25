package outbound

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQSActionType string

const (
	SQSActionMessage  SQSActionType = "message"
	SQSActionFile     SQSActionType = "file"
	SQSActionLocation SQSActionType = "location"
)

type SQSMessageBody struct {
	CustomerID string        `json:"customer_id"`
	SessionID  string        `json:"session_id"`
	ChatID     string        `json:"chat_id"`
	Message    string        `json:"message,omitempty"`
	APIKey     string        `json:"api_key"`
	ActionType SQSActionType `json:"action_type"`
	Location   *SQSLocation  `json:"location,omitempty"`
}

type SQSLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Title     string  `json:"title"`
}

type SQSMessage struct {
	QueueURL   string                                 `json:"queue_url"`
	Body       any                                    `json:"body"`
	Attributes map[string]types.MessageAttributeValue `json:"attributes"`
}

type SQSAdapter interface {
	SendMessage(ctx context.Context, message SQSMessage) error
}
