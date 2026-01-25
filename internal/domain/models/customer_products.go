package models

import "time"

type BusinessProduct struct {
	BusinessPhoneNumber string    `json:"business_phone_number"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	Category            string    `json:"category"` // pdf, image, video
	Price               float64   `json:"price"`
	File                string    `json:"file"`
	Quantity            int       `json:"quantity"` // only for products if category is product
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
