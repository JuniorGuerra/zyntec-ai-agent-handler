package models

import "time"

type CalendarActionType string

const (
	CalendarActionSchedule   CalendarActionType = "schedule"
	CalendarActionCancel     CalendarActionType = "cancel"
	CalendarActionReschedule CalendarActionType = "reschedule"
)

type Calendar struct {
	CustomerID   string    `json:"customer_id" dynamodbav:"customer_id"`
	RefreshToken string    `json:"refresh_token" dynamodbav:"refresh_token"`
	Timezone     string    `json:"timezone" dynamodbav:"timezone"`
	CreatedAt    time.Time `json:"created_at" dynamodbav:"created_at"`
	ExpiresAt    time.Time `json:"expires_at" dynamodbav:"expires_at"`
}

type CalendarEventRequest struct {
	Action      CalendarActionType `json:"action"`
	CustomerID  string             `json:"customer_id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	StartTime   string             `json:"start_time"`
	EndTime     string             `json:"end_time"`
	Attendees   []string           `json:"attendees"`
}

type CancelEventRequest struct {
	Action     CalendarActionType `json:"action"`
	CustomerID string             `json:"customer_id"`
	Date       string             `json:"date"`
	Time       string             `json:"time"`
	Reason     string             `json:"reason"`
}

type RescheduleEventRequest struct {
	Action       CalendarActionType `json:"action"`
	CustomerID   string             `json:"customer_id"`
	OriginalDate string             `json:"original_date"`
	OriginalTime string             `json:"original_time"`
	NewDate      string             `json:"new_date"`
	NewTime      string             `json:"new_time"`
	Reason       string             `json:"reason"`
}
