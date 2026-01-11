# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

WhatsApp chatbot backend built as an AWS Lambda function in Go. The application is designed to handle chatbot API requests via API Gateway and persist data to DynamoDB.

## Architecture

The project follows **Hexagonal Architecture** (Ports & Adapters):

- **cmd/chatbot-api-handler**: Application entry point for the Lambda function
- **internal/adapters/handlers**: HTTP/Lambda event handlers (inbound adapters)
- **internal/adapters/repositories**: Database implementations (outbound adapters)
  - Repository interface defined in `interfaces.go`
  - DynamoDB implementation in `dynamodb.go`
- **internal/core**: Domain logic and models (when implemented)
- **pkg**: Shared utilities accessible to external packages (currently empty)

The main handler (`handlers.Handler`) is invoked by AWS Lambda via API Gateway proxy events.

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
All database operations go through the `Repository` interface. The factory function `NewRepository()` returns a DynamoDB implementation. When adding new data operations:
1. Add method to `Repository` interface in `internal/adapters/repositories/interfaces.go`
2. Implement in `DynamoDBRepository` in `internal/adapters/repositories/dynamodb.go`

### Lambda Handler
The handler receives `events.APIGatewayProxyRequest` and returns `events.APIGatewayProxyResponse`. The response must include `StatusCode` and `Body` fields.

## AWS Dependencies

- Uses AWS Lambda Go SDK (`github.com/aws/aws-lambda-go`)
- Uses AWS SDK for Go v2 for DynamoDB operations
- Designed to run as a Lambda function behind API Gateway
