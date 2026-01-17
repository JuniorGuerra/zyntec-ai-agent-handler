package sqs

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"app/internal/ports/outbound"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSAdapter struct {
	client *sqs.Client
}

func NewSQSAdapter() outbound.SQSAdapter {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		slog.Error("failed to load default config", "error", err)
		return nil
	}

	client := sqs.NewFromConfig(cfg)
	return &SQSAdapter{client: client}
}

func (a *SQSAdapter) SendMessage(ctx context.Context, message outbound.SQSMessage) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	body, err := json.Marshal(message.Body)
	if err != nil {
		slog.Error("failed to marshal message body", "error", err)
		return err
	}

	_, err = a.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:          aws.String(message.QueueURL),
		MessageBody:       aws.String(string(body)),
		MessageAttributes: message.Attributes,
	})

	if err != nil {
		slog.Error("failed to send message", "error", err)
		return err
	}

	return nil
}
