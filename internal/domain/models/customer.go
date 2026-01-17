package models

import "time"

type Customer struct {
	BusinessPhoneNumber string    `json:"business_phone_number" dynamodbav:"business_phone_number"`
	CreatedAt           time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" dynamodbav:"updated_at"`
	AgentTimeout        int       `json:"agent_timeout" dynamodbav:"agent_timeout"`
	AIPrompt            string    `json:"ai_prompt" dynamodbav:"ai_prompt"`
	ProductsInfo        string    `json:"products_info" dynamodbav:"products_info"`
	AIModel             string    `json:"ai_model" dynamodbav:"ai_model"`
	IsActive            bool      `json:"is_active" dynamodbav:"is_active"`
	APIKey              string    `json:"api_key" dynamodbav:"api_key"`
	URL                 string    `json:"url" dynamodbav:"url"`
	SessionName         string    `json:"session_name" dynamodbav:"session_name"` // used to identify the session in waha
}

func (c *Customer) IsValid() bool {
	return c.IsActive && c.APIKey != ""
}
