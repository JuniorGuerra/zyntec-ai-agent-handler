package models

import "time"

type Calendar struct {
	ID           string    `json:"id"`
	CustomerID   string    `json:"customer_id"`
	RefreshToken string    `json:"refresh_token"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}
