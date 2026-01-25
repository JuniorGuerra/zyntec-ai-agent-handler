package whatsapp

import (
	"app/internal/domain/models"
	"app/internal/ports/outbound"
	"bytes"
	"encoding/json"
	"net/http"
)

type WahaAdapter struct {
	client     *http.Client
	defaultURL string
}

func NewWahaAdapter(url string) outbound.WhatsAppPort {
	return &WahaAdapter{client: &http.Client{}, defaultURL: url}
}

func (a *WahaAdapter) SendMessage(input models.SendMessageInput) error {
	body := map[string]any{
		"chatId":                 input.PhoneNumber,
		"text":                   input.Message,
		"session":                input.SessionName,
		"linkPreview":            true,
		"linkPreviewHighQuality": true,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := a.defaultURL + "/api/sendText"
	if input.URL != "" {
		url = input.URL + "/api/sendText"
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", input.APIKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (a *WahaAdapter) SendLocation(input models.SendLocationInput) error {
	body := map[string]any{
		"chatId":    input.ChatID,
		"latitude":  input.Latitude,
		"longitude": input.Longitude,
		"title":     input.Title,
		"session":   input.SessionName,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := a.defaultURL + "/api/sendLocation"
	if input.URL != "" {
		url = input.URL + "/api/sendLocation"
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", input.APIKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
