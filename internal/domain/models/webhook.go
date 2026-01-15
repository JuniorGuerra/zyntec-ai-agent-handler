package models

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

type From string

func (ft From) IsValidFromType() bool {
	return ft != "status@broadcast"
}

func (ft From) String() string {
	return string(ft)
}
