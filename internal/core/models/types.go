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
	ClientName          string    `json:"client_name" dynamodbav:"client_name"`
	IsHumanAgent        bool      `json:"is_human_agent" dynamodbav:"is_human_agent"`
	CreatedAt           time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" dynamodbav:"updated_at"`
	SessionExpiryAt     time.Time `json:"session_expiry_at" dynamodbav:"session_expiry_at"`
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

type Customer struct {
	BusinessPhoneNumber string    `json:"business_phone_number" dynamodbav:"business_phone_number"`
	CreatedAt           time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" dynamodbav:"updated_at"`
	// AgentTimeout is the time in minutes that the agent has to respond to the customer
	AgentTimeout int    `json:"agent_timeout" dynamodbav:"agent_timeout"`
	AIPrompt     string `json:"ai_prompt" dynamodbav:"ai_prompt"`
	ProductsInfo string `json:"products_info" dynamodbav:"products_info"`
	AIModel      string `json:"ai_model" dynamodbav:"ai_model"`
	IsActive     bool   `json:"is_active" dynamodbav:"is_active"`
}

type WebhookRequest struct {
	Me      WebhookMe      `json:"me"`
	Payload WebhookPayload `json:"payload"`
}

type WebhookMe struct {
	ID string `json:"id"`
}

type WebhookPayload struct {
	From        From   `json:"from"`
	To          string `json:"to"`
	Body        string `json:"body"`
	Timestamp   int64  `json:"timestamp"`
	FromMe      bool   `json:"fromMe"`
	Participant string `json:"participant,omitempty"`
}

type Role string
type From string

func (r Role) IsValidRole() bool {
	return r == CustomerRole || r == AssistantRole || r == FunctionRole
}

func (r Role) String() string {
	return string(r)
}

func (ft From) IsValidFromType() bool {
	return ft == "status@broadcast"
}

func (ft From) String() string {
	return string(ft)
}
