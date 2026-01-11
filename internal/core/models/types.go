package models

import "time"

type Session struct {
	ID              string    `json:"id" dynamodbav:"id"`
	CreatedAt       time.Time `json:"created_at" dynamodbav:"created_at"`
	SessionExpiryAt time.Time `json:"session_expiry_at" dynamodbav:"session_expiry_at"`
}

type Message struct {
	ID        string    `json:"id" dynamodbav:"id"`
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
	SessionID string    `json:"session_id" dynamodbav:"session_id"`
	Body      string    `json:"body" dynamodbav:"body"`
}
