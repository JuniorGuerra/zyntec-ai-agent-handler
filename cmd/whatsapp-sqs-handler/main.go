package whatsappsqshandler

import (
	"app/internal/adapters/inbound/sqs"
	sqsAdapter "app/internal/adapters/outbound/sqs"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func main() {
	adapter := sqsAdapter.NewSQSAdapter()
	handler := sqs.NewWhatsAppHandler(adapter)

	lambda.Start(handler.HandleMessage)
}
