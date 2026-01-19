package outbound

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQSMessageBody struct {
	CustomerID string `json:"customer_id"`
	SessionID  string `json:"session_id"`
	Message    string `json:"message"`
}

type SQSMessage struct {
	QueueURL   string                                 `json:"queue_url"`
	Body       any                                    `json:"body"`
	Attributes map[string]types.MessageAttributeValue `json:"attributes"`
}

type SQSAdapter interface {
	SendMessage(ctx context.Context, message SQSMessage) error
}
