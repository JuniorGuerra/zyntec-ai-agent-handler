package models

import "time"

type Calendar struct {
	CustomerID   string    `json:"customer_id" dynamodbav:"customer_id"`
	RefreshToken string    `json:"refresh_token" dynamodbav:"refresh_token"`
	CreatedAt    time.Time `json:"created_at" dynamodbav:"created_at"`
	ExpiresAt    time.Time `json:"expires_at" dynamodbav:"expires_at"`
}
