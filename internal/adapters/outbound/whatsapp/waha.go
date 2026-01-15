package whatsapp

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type WahaAdapter struct {
	client *http.Client
}

func NewWahaAdapter() *WahaAdapter {
	return &WahaAdapter{client: &http.Client{}}
}

func (a *WahaAdapter) SendMessage(phoneNumber string, message string) error {
	body := map[string]interface{}{
		"chatId":                 phoneNumber,
		"text":                   message,
		"session":                "default",
		"linkPreview":            true,
		"linkPreviewHighQuality": false,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "http://localhost:3000/api/sendText", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
