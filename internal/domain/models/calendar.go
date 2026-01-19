package models

import "time"

type Calendar struct {
	CustomerID   string    `json:"customer_id" dynamodbav:"customer_id"`
	RefreshToken string    `json:"refresh_token" dynamodbav:"refresh_token"`
	Timezone     string    `json:"timezone" dynamodbav:"timezone"`
	CreatedAt    time.Time `json:"created_at" dynamodbav:"created_at"`
	ExpiresAt    time.Time `json:"expires_at" dynamodbav:"expires_at"`
}

type CalendarEventRequest struct {
	CustomerID  string   `json:"customer_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	StartTime   string   `json:"start_time"`
	EndTime     string   `json:"end_time"`
	Attendees   []string `json:"attendees"`
}
