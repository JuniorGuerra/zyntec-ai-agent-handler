package models

import "time"

type Customer struct {
	BusinessPhoneNumber string    `json:"business_phone_number" dynamodbav:"business_phone_number"`
	CreatedAt           time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" dynamodbav:"updated_at"`
	AgentTimeout        int       `json:"agent_timeout" dynamodbav:"agent_timeout"`
	AIPrompt            string    `json:"ai_prompt" dynamodbav:"ai_prompt"`
	AIModel             string    `json:"ai_model" dynamodbav:"ai_model"`
	IsActive            bool      `json:"is_active" dynamodbav:"is_active"`
	APIKey              string    `json:"api_key" dynamodbav:"api_key"`
	URL                 string    `json:"url" dynamodbav:"url"`
	SessionName         string    `json:"session_name" dynamodbav:"session_name"`
	IsCalendarActive    bool      `json:"is_calendar_active" dynamodbav:"is_calendar_active"`
	Location            Location  `json:"location" dynamodbav:"location"`
	HoursOfOperation    string    `json:"hours_of_operation" dynamodbav:"hours_of_operation"`
}

type Location struct {
	Address   string  `json:"address" dynamodbav:"address"`
	Latitude  float64 `json:"latitude" dynamodbav:"latitude"`
	Longitude float64 `json:"longitude" dynamodbav:"longitude"`
	Title     string  `json:"title" dynamodbav:"title"`
}

func (c *Customer) IsValid() bool {
	return c.IsActive && c.APIKey != ""
}
