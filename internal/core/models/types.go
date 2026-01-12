package models

import "time"

const (
	CustomerRole  = "user"
	AssistantRole = "model"
	FunctionRole  = "function"
)

type Session struct {
	ID                  string    `json:"id" dynamodbav:"id"`
	BusinessPhoneNumber string    `json:"business_phone_number" dynamodbav:"business_phone_number"`
	CustomerPhoneNumber string    `json:"customer_phone_number" dynamodbav:"customer_phone_number"`
	CreatedAt           time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" dynamodbav:"updated_at"`
	SessionExpiryAt     time.Time `json:"session_expiry_at" dynamodbav:"session_expiry_at"`
}

type Message struct {
	ID                  string    `json:"id" dynamodbav:"id"`
	BusinessPhoneNumber string    `json:"business_phone_number" dynamodbav:"business_phone_number"`
	CustomerPhoneNumber string    `json:"customer_phone_number" dynamodbav:"customer_phone_number"`
	SessionID           string    `json:"session_id" dynamodbav:"session_id"`
	Timestamp           int64     `json:"timestamp" dynamodbav:"timestamp"`
	Message             string    `json:"message" dynamodbav:"message"`
	Role                string    `json:"interaction" dynamodbav:"interaction"`
	CreatedAt           time.Time `json:"created_at" dynamodbav:"created_at"`
}

type Customer struct {
	BusinessPhoneNumber string    `json:"business_phone_number" dynamodbav:"business_phone_number"`
	CreatedAt           time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" dynamodbav:"updated_at"`
	AIPrompt            string    `json:"ai_prompt" dynamodbav:"ai_prompt"`
	ProductsInfo        string    `json:"products_info" dynamodbav:"products_info"`
}

type WebhookRequest struct {
	Me      WebhookMe      `json:"me"`
	Payload WebhookPayload `json:"payload"`
}

type WebhookMe struct {
	ID string `json:"id"`
}

type WebhookPayload struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Body        string `json:"body"`
	Timestamp   int64  `json:"timestamp"`
	FromMe      bool   `json:"fromMe"`
	Participant string `json:"participant,omitempty"`
}
