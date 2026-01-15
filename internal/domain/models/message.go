package models

import "time"

const (
	CustomerRole  = "user"
	AssistantRole = "model"
	FunctionRole  = "function"
)

type Role string

func (r Role) IsValidRole() bool {
	return r == CustomerRole || r == AssistantRole || r == FunctionRole
}

func (r Role) String() string {
	return string(r)
}

type Message struct {
	ID                  string    `json:"id" dynamodbav:"id"`
	BusinessPhoneNumber string    `json:"business_phone_number" dynamodbav:"business_phone_number"`
	CustomerPhoneNumber string    `json:"customer_phone_number" dynamodbav:"customer_phone_number"`
	SessionID           string    `json:"session_id" dynamodbav:"session_id"`
	Timestamp           string    `json:"timestamp" dynamodbav:"timestamp"`
	Message             string    `json:"message" dynamodbav:"message"`
	Role                Role      `json:"interaction" dynamodbav:"interaction"`
	CreatedAt           time.Time `json:"created_at" dynamodbav:"created_at"`
}
