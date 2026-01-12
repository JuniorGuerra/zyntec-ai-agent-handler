# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

WhatsApp chatbot backend built as an AWS Lambda function in Go. The application receives webhook requests from WhatsApp via API Gateway, processes messages using Google Gemini AI, persists conversation data to DynamoDB, and can send responses back via the WAHA (WhatsApp HTTP API) service.

## Architecture

The project follows **Hexagonal Architecture** (Ports & Adapters):

- **cmd/chatbot-api-handler**: Application entry point for the Lambda function
  - Initializes JSON logging via `slog` package
  - Uses `errgroup` to initialize services concurrently: DynamoDB repository, WAHA service, and Gemini AI service
  - Requires `GEMINI_API_KEY` environment variable
- **internal/adapters/handlers**: HTTP/Lambda event handlers (inbound adapters)
  - `ChatbotAPIHandler` receives API Gateway proxy events containing WhatsApp webhook payloads
  - Manages conversation flow: session creation/retrieval, message persistence, AI response generation
- **internal/adapters/repositories**: Database implementations (outbound adapters)
  - Repository interface defined in `interfaces.go`
  - DynamoDB implementation in `dynamodb.go`
  - Uses AWS SDK v2 with 10-second context timeouts for all operations
  - Tables: `sessions`, `messages`, `customers`
- **internal/core/models**: Domain models with JSON and DynamoDB tags
  - `Session`: Tracks conversations between business and customer phone numbers with expiry
  - `Message`: Individual messages with role (user/model/function)
  - `Customer`: Business configuration including AI prompts and product info
  - `WebhookRequest`: Incoming WhatsApp webhook payload structure
- **internal/core/services/ai**: AI service implementations
  - `AIModels` interface with `GenerateResponse` and `SetSystemInstruction` methods
  - Gemini implementation uses chat sessions with conversation history
  - 10-second timeout for AI generation
- **internal/core/services/whatsapp-svc**: WhatsApp messaging service
  - WAHA (WhatsApp HTTP API) integration for sending messages
  - Currently commented out in handler (line 126 in chatbot-api-handler.go)

## Key Data Flow

1. WhatsApp webhook arrives at Lambda via API Gateway
2. Handler unmarshals `WebhookRequest` from request body
3. Session ID generated from business and customer phone numbers
4. Session retrieved from DynamoDB or created if new (1-hour expiry)
5. Customer message saved to DynamoDB
6. Message history retrieved and passed to Gemini as context
7. Gemini generates AI response based on system instruction and history
8. AI response saved as assistant message
9. (Optional) Response sent back via WAHA service

## Development Commands

### Build
```bash
go build -o bootstrap ./cmd/chatbot-api-handler/
```

The output binary must be named `bootstrap` for AWS Lambda custom runtime compatibility.

### Run Tests
```bash
go test ./...
```

Run a specific test:
```bash
go test -v -run TestGenerateResponse ./internal/core/services/ai
```

Run tests with verbose output:
```bash
go test -v ./...
```

Note: AI tests require `GEMINI_API_KEY` environment variable.

### Dependency Management
```bash
go mod tidy       # Clean up dependencies
go mod download   # Download dependencies
```

## Key Patterns

### Repository Pattern
All database operations go through the `Repository` interface. The factory function `NewDynamoDBRepository()` returns a DynamoDB implementation. When adding new data operations:
1. Add method to `Repository` interface in `internal/adapters/repositories/interfaces.go`
2. Implement in `DynamoDBRepository` in `internal/adapters/repositories/dynamodb.go`
3. Use `attributevalue.MarshalMap()` to convert Go structs to DynamoDB items
4. All DynamoDB operations use 10-second context timeouts

### AI Service Pattern
AI providers implement the `AIModels` interface. The Gemini implementation:
- Maintains chat session state with conversation history
- Requires `SetSystemInstruction()` call before `GenerateResponse()`
- Converts `[]models.Message` history to Gemini format
- Filters messages by valid roles (user/model/function)

### Service Initialization
Services are initialized concurrently using `errgroup` in main.go. This pattern allows parallel setup of independent services (repository, WhatsApp, AI) while collecting any initialization errors.

### Lambda Handler
The handler receives `events.APIGatewayProxyRequest` and returns `events.APIGatewayProxyResponse`. The response must include `StatusCode` and `Body` fields. Error handling returns HTTP 500 with error description in the body.

### Logging
The application uses structured JSON logging via `log/slog`:
- Configured in `cmd/chatbot-api-handler/main.go` init function
- Use `slog.Info()`, `slog.Error()` for log statements
- Logs are output to stdout in JSON format for CloudWatch integration

## AWS Dependencies

- Uses AWS Lambda Go SDK (`github.com/aws/aws-lambda-go`)
- Uses AWS SDK for Go v2 for DynamoDB operations
- DynamoDB tables: `sessions` (primary key: id), `messages` (scanned by session_id), `customers` (primary key: business_phone_number)
- Designed to run as a Lambda function behind API Gateway
- Note: `GetMessageHistory` uses Scan with FilterExpression (consider adding GSI on session_id for production)

## External Services

- **Google Gemini**: AI model "gemini-2.0-flash-exp" for response generation
- **WAHA**: WhatsApp HTTP API service (expects local endpoint at http://localhost:3000/api/sendText)
