package models

import "time"

type Session struct {
	ID                  string    `json:"id" dynamodbav:"id"`
	BusinessPhoneNumber string    `json:"business_phone_number" dynamodbav:"business_phone_number"`
	CustomerPhoneNumber string    `json:"customer_phone_number" dynamodbav:"customer_phone_number"`
	ClientName          string    `json:"client_name" dynamodbav:"client_name"`
	IsHumanAgent        bool      `json:"is_human_agent" dynamodbav:"is_human_agent"`
	CreatedAt           time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" dynamodbav:"updated_at"`
	SessionExpiryAt     time.Time `json:"session_expiry_at" dynamodbav:"session_expiry_at"`
}
