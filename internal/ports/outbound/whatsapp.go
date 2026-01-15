package outbound

type WhatsAppPort interface {
	SendMessage(phoneNumber string, message string) error
}
