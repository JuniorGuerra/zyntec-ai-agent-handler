# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

WhatsApp chatbot backend built as an AWS Lambda function in Go. The application is designed to handle chatbot API requests via API Gateway and persist data to DynamoDB.

## Architecture

The project follows **Hexagonal Architecture** (Ports & Adapters):

- **cmd/chatbot-api-handler**: Application entry point for the Lambda function
  - Initializes JSON logging via `slog` package
  - Wires dependencies: creates DynamoDBRepository and injects into ChatbotAPIHandler
- **internal/adapters/handlers**: HTTP/Lambda event handlers (inbound adapters)
  - `ChatbotAPIHandler` receives API Gateway proxy events
- **internal/adapters/repositories**: Database implementations (outbound adapters)
  - Repository interface defined in `interfaces.go`
  - DynamoDB implementation in `dynamodb.go`
  - Uses AWS SDK v2 with 10-second context timeouts for all operations
- **internal/core/models**: Domain models (`Session`, `Message`) with JSON and DynamoDB tags
- **internal/core/services/ai**: AI service interfaces (placeholder for future AI integration)

The main handler (`ChatbotAPIHandler.Handle`) receives `events.APIGatewayProxyRequest` and persists session/message data to DynamoDB tables (`sessions` and `messages`).

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

Run tests for a specific package:
```bash
go test ./internal/adapters/handlers
```

Run tests with verbose output:
```bash
go test -v ./...
```

### Dependency Management
```bash
go mod tidy       # Clean up dependencies
go mod download   # Download dependencies
go mod vendor     # Vendor dependencies (if needed)
```

## Key Patterns

### Repository Pattern
All database operations go through the `Repository` interface. The factory function `NewDynamoDBRepository()` returns a DynamoDB implementation. When adding new data operations:
1. Add method to `Repository` interface in `internal/adapters/repositories/interfaces.go`
2. Implement in `DynamoDBRepository` in `internal/adapters/repositories/dynamodb.go`
3. Use `attributevalue.MarshalMap()` to convert Go structs to DynamoDB items
4. All DynamoDB operations use 10-second context timeouts

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
- DynamoDB tables: `sessions` and `messages`
- Designed to run as a Lambda function behind API Gateway
