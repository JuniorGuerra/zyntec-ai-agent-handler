package repositories

import (
	"app/internal/core/models"
	"context"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBRepository struct {
	dynamoDBClient *dynamodb.Client
}

var (
	sessionTableName = aws.String("sessions")
	messageTableName = aws.String("messages")
)

func NewDynamoDBRepository() *DynamoDBRepository {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		slog.Error("failed to load default config", "error", err)
		return nil
	}

	client := dynamodb.NewFromConfig(cfg)

	return &DynamoDBRepository{
		dynamoDBClient: client,
	}
}

func (r *DynamoDBRepository) SaveSession(session models.Session) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	item, err := attributevalue.MarshalMap(session)
	if err != nil {
		slog.Error("failed to marshal session", "error", err)
		return err
	}

	_, err = r.dynamoDBClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: sessionTableName,
		Item:      item,
	})

	return err
}

func (r *DynamoDBRepository) SaveMessage(message models.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	item, err := attributevalue.MarshalMap(message)
	if err != nil {
		slog.Error("failed to marshal message", "error", err)
		return err
	}

	_, err = r.dynamoDBClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: messageTableName,
		Item:      item,
	})

	return err
}
