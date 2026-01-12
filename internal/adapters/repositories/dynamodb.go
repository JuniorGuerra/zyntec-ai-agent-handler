package repositories

import (
	"app/internal/core/models"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBRepository struct {
	dynamoDBClient *dynamodb.Client
}

var (
	sessionTableName  = aws.String("zyntec_agent_sessions")
	messageTableName  = aws.String("zyntec_agent_messages")
	customerTableName = aws.String("zyntec_agent_customers")
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
		slog.Error("failed to marshal session", "error", err, "session", session)
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
		slog.Error("failed to marshal message", "error", err, "message", message)
		return err
	}

	_, err = r.dynamoDBClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: messageTableName,
		Item:      item,
	})

	return err
}

func (r *DynamoDBRepository) GetSession(sessionID string) (*models.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.dynamoDBClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: sessionTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: sessionID},
		},
	})

	if err != nil {
		slog.Error("failed to get session", "error", err, "session_id", sessionID)
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	var session models.Session
	err = attributevalue.UnmarshalMap(result.Item, &session)
	if err != nil {
		slog.Error("failed to unmarshal session", "error", err)
		return nil, err
	}

	return &session, nil
}

func (r *DynamoDBRepository) GetCustomer(businessPhoneNumber string) (*models.Customer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.dynamoDBClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: customerTableName,
		Key: map[string]types.AttributeValue{
			"business_phone_number": &types.AttributeValueMemberS{Value: businessPhoneNumber},
		},
	})

	if err != nil {
		slog.Error("failed to get customer", "error", err, "business_phone", businessPhoneNumber)
		return nil, err
	}

	if result.Item == nil {
		return nil, fmt.Errorf("customer not found: %s", businessPhoneNumber)
	}

	var customer models.Customer
	err = attributevalue.UnmarshalMap(result.Item, &customer)
	if err != nil {
		slog.Error("failed to unmarshal customer", "error", err, "bussiess_phone", businessPhoneNumber)
		return nil, err
	}

	return &customer, nil
}

func (r *DynamoDBRepository) GetMessageHistory(sessionID string) ([]models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.dynamoDBClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              messageTableName,
		KeyConditionExpression: aws.String("session_id = :sid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":sid": &types.AttributeValueMemberS{Value: sessionID},
		},
		Limit:            aws.Int32(20),
		ScanIndexForward: aws.Bool(false),
	})

	if err != nil {
		slog.Error("failed to query message history", "error", err, "session_id", sessionID)
		return nil, err
	}

	var messages []models.Message
	err = attributevalue.UnmarshalListOfMaps(result.Items, &messages)
	if err != nil {
		slog.Error("failed to unmarshal messages", "error", err, "session_id", sessionID)
		return nil, err
	}

	return messages, nil
}

func (r *DynamoDBRepository) UpdateSession(session models.Session) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	item, err := attributevalue.MarshalMap(session)
	if err != nil {
		slog.Error("failed to marshal session", "error", err, "session", session)
		return err
	}

	_, err = r.dynamoDBClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: sessionTableName,
		Item:      item,
	})

	if err != nil {
		slog.Error("failed to update session", "error", err, "session", session)
		return err
	}

	return nil
}
