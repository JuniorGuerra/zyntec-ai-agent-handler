package whatsappsvc

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type WahaService struct {
	client *http.Client
}

func NewWahaService() WhatsAppService {
	return &WahaService{
		client: &http.Client{},
	}
}

func (s *WahaService) SendWhatsAppMessage(phoneNumber string, message string) error {

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

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

/*
curl -X 'POST' \
  'http://localhost:3000/api/sendText' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "chatId": "573234574967@c.us",
  "reply_to": null,
  "text": "Hi there!",
  "linkPreview": true,
  "linkPreviewHighQuality": false,
  "session": "default"
}'
*/
