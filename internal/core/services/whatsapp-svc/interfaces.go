package whatsappsvc

type WhatsAppService interface {
	SendWhatsAppMessage(phoneNumber string, message string) error
}
