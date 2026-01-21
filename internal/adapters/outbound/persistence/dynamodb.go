package persistence

import (
	"app/internal/domain/models"
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

var (
	sessionTableName              = aws.String("zyntec_agent_sessions")
	messageTableName              = aws.String("zyntec_agent_messages")
	customerTableName             = aws.String("zyntec_agent_customers")
	calendarTableName             = aws.String("zyntec_agent_calendars")
	calendarAppointmentsTableName = aws.String("zyntec_agent_calendar_appointments")
)

type DynamoDBClient struct {
	client *dynamodb.Client
}

func NewDynamoDBClient() *DynamoDBClient {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		slog.Error("failed to load default config", "error", err)
		return nil
	}

	client := dynamodb.NewFromConfig(cfg)
	return &DynamoDBClient{client: client}
}

// SessionRepository implementation

type SessionRepository struct {
	db *DynamoDBClient
}

func NewSessionRepository(db *DynamoDBClient) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Save(session models.Session) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	item, err := attributevalue.MarshalMap(session)
	if err != nil {
		slog.Error("failed to marshal session", "error", err, "session", session)
		return err
	}

	_, err = r.db.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: sessionTableName,
		Item:      item,
	})
	return err
}

func (r *SessionRepository) Get(sessionID string) (*models.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.db.client.GetItem(ctx, &dynamodb.GetItemInput{
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
	if err := attributevalue.UnmarshalMap(result.Item, &session); err != nil {
		slog.Error("failed to unmarshal session", "error", err)
		return nil, err
	}

	return &session, nil
}

// MessageRepository implementation

type MessageRepository struct {
	db *DynamoDBClient
}

func NewMessageRepository(db *DynamoDBClient) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Save(messages ...models.Message) error {
	if len(messages) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var writeRequests []types.WriteRequest
	for _, message := range messages {
		item, err := attributevalue.MarshalMap(message)
		if err != nil {
			slog.Error("failed to marshal message", "error", err, "message", message)
			return err
		}

		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{Item: item},
		})
	}

	_, err := r.db.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			*messageTableName: writeRequests,
		},
	})
	if err != nil {
		slog.Error("failed to batch write messages", "error", err, "count", len(messages))
		return err
	}

	return nil
}

func (r *MessageRepository) GetHistory(sessionID string) ([]models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.db.client.Query(ctx, &dynamodb.QueryInput{
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
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &messages); err != nil {
		slog.Error("failed to unmarshal messages", "error", err, "session_id", sessionID)
		return nil, err
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// CustomerRepository implementation

type CustomerRepository struct {
	db *DynamoDBClient
}

func NewCustomerRepository(db *DynamoDBClient) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Get(businessPhoneNumber string) (*models.Customer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.db.client.GetItem(ctx, &dynamodb.GetItemInput{
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
	if err := attributevalue.UnmarshalMap(result.Item, &customer); err != nil {
		slog.Error("failed to unmarshal customer", "error", err, "business_phone", businessPhoneNumber)
		return nil, err
	}

	return &customer, nil
}

// CalendarRepository implementation

type CalendarRepository struct {
	db *DynamoDBClient
}

func NewCalendarRepository(db *DynamoDBClient) *CalendarRepository {
	return &CalendarRepository{db: db}
}

func (r *CalendarRepository) Save(calendar models.Calendar) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	item, err := attributevalue.MarshalMap(calendar)
	if err != nil {
		slog.Error("failed to marshal calendar", "error", err, "calendar", calendar)
		return err
	}

	_, err = r.db.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: calendarTableName,
		Item:      item,
	})
	return err
}

func (r *CalendarRepository) GetByCustomerID(customerID string) (*models.Calendar, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.db.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: calendarTableName,
		Key: map[string]types.AttributeValue{
			"customer_id": &types.AttributeValueMemberS{Value: customerID},
		},
	})
	if err != nil {
		slog.Error("failed to get calendar", "error", err, "customer_id", customerID)
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	var calendar models.Calendar
	if err := attributevalue.UnmarshalMap(result.Item, &calendar); err != nil {
		slog.Error("failed to unmarshal calendar", "error", err, "customer_id", customerID)
		return nil, err
	}

	return &calendar, nil
}

type CalendarAppointmentsRepository struct {
	db *DynamoDBClient
}

func NewCalendarAppointmentsRepository(db *DynamoDBClient) *CalendarAppointmentsRepository {
	return &CalendarAppointmentsRepository{db: db}
}

func (r *CalendarAppointmentsRepository) Save(appointment models.CalendarEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	item, err := attributevalue.MarshalMap(appointment)
	if err != nil {
		slog.Error("failed to marshal calendar event", "error", err, "calendar_event", appointment)
		return err
	}

	_, err = r.db.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: calendarAppointmentsTableName,
		Item:      item,
	})
	return err
}

func (r *CalendarAppointmentsRepository) GetByCustomerPhoneNumber(customerPhoneNumber string) ([]models.CalendarEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.db.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              calendarAppointmentsTableName,
		KeyConditionExpression: aws.String("customer_phone_number = :cid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":cid": &types.AttributeValueMemberS{Value: customerPhoneNumber},
		},
		Limit:            aws.Int32(20),
		ScanIndexForward: aws.Bool(false),
	})
	if err != nil {
		slog.Error("failed to query calendar appointments", "error", err, "customer_phone_number", customerPhoneNumber)
		return nil, err
	}

	var appointments []models.CalendarEvent
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &appointments); err != nil {
		slog.Error("failed to unmarshal calendar appointments", "error", err, "customer_phone_number", customerPhoneNumber)
		return nil, err
	}

	for i, j := 0, len(appointments)-1; i < j; i, j = i+1, j-1 {
		appointments[i], appointments[j] = appointments[j], appointments[i]
	}

	return appointments, nil
}

func (r *CalendarAppointmentsRepository) Delete(customerPhoneNumber, eventID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.db.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: calendarAppointmentsTableName,
		Key: map[string]types.AttributeValue{
			"customer_phone_number": &types.AttributeValueMemberS{Value: customerPhoneNumber},
			"event_id":              &types.AttributeValueMemberS{Value: eventID},
		},
	})
	if err != nil {
		slog.Error("failed to delete calendar appointment", "error", err, "customer_phone_number", customerPhoneNumber, "event_id", eventID)
		return err
	}

	return nil
}
