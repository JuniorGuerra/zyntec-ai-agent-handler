package inbound

import "github.com/aws/aws-lambda-go/events"

type ChatbotUseCase interface {
	HandleWebhook(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
}
