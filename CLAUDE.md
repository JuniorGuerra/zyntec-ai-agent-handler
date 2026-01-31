# CLAUDE.md - AI Assistant Guide for Zyntec Bot

This document provides comprehensive guidance for AI assistants working with the Zyntec Bot codebase.

## Project Overview

**Zyntec Bot** is a serverless WhatsApp chatbot powered by Google Gemini that provides intelligent conversational AI, calendar appointment management, human agent transfers, and automated business actions through function calling.

**Primary Use Cases:**
- Intelligent customer service via WhatsApp
- Calendar appointment scheduling, cancellation, and rescheduling
- Seamless transfer to human agents with configurable timeouts
- Multi-tenant support (multiple businesses on same infrastructure)

## Technology Stack

| Component | Technology | Version/Details |
|-----------|------------|-----------------|
| Language | Go | 1.25 |
| Runtime | AWS Lambda | Serverless functions |
| Database | DynamoDB | 6 tables for persistence |
| AI Model | Google Gemini | gemini-2.0-flash-exp |
| Message Queue | AWS SQS | Async message processing |
| WhatsApp API | WAHA | WhatsApp HTTP API |
| Calendar | Google Calendar | OAuth2 integration |
| Auth | OAuth2 | For Google Calendar access |

## Architecture

The project implements **Hexagonal Architecture** (Ports & Adapters):

```
Adapters (External) → Ports (Interfaces) → Application (Use Cases) → Domain (Core)
```

### Layer Responsibilities

- **Domain** (`internal/domain/`): Pure business logic, zero external dependencies
  - `models/`: Entity definitions (Session, Message, Customer, Calendar, etc.)
  - `utils/`: Helper functions (UUID generation, session ID)

- **Ports** (`internal/ports/`): Interface contracts
  - `inbound/`: Use case interfaces (ChatbotUseCase)
  - `outbound/`: External service interfaces (AIPort, Repositories, WhatsAppPort, CalendarPort)

- **Application** (`internal/application/`): Business logic orchestration
  - `chatbot/`: Main chatbot service and action handlers
  - `waha/`: WhatsApp message sending service
  - `calendar/`: Calendar operations service
  - `calendar-starter/`: OAuth callback handler

- **Adapters** (`internal/adapters/`): Concrete implementations
  - `inbound/http/`: HTTP webhook handlers
  - `inbound/sqs/`: SQS queue consumers
  - `outbound/ai/`: Gemini API client
  - `outbound/persistence/`: DynamoDB repositories
  - `outbound/calendar/`: Google Calendar client
  - `outbound/whatsapp/`: WAHA HTTP client
  - `outbound/sqs/`: SQS message sender

## Directory Structure

```
zyntec-ai-agent-handler/
├── cmd/                           # Lambda entry points
│   ├── chatbot-api-handler/       # WhatsApp webhook receiver
│   ├── whatsapp-sqs-handler/      # WhatsApp message sender
│   ├── calendar-api-handler/      # OAuth callback handler
│   ├── calendar-sqs-handler/      # Calendar event processor
│   └── config/                    # Configuration loaders
│       ├── chatbot-api-config/
│       ├── whatsapp-sqs-config/
│       └── calendar-general-config/
│
├── internal/
│   ├── ports/
│   │   ├── inbound/chatbot.go     # ChatbotUseCase interface
│   │   └── outbound/
│   │       ├── ai.go              # AIPort, ActionType definitions
│   │       ├── repository.go      # Repository interfaces
│   │       ├── calendar.go        # CalendarPort interface
│   │       ├── whatsapp.go        # WhatsAppPort interface
│   │       └── sqs.go             # SQSAdapter interface
│   │
│   ├── domain/
│   │   ├── models/                # Domain entities
│   │   │   ├── session.go
│   │   │   ├── message.go
│   │   │   ├── customer.go
│   │   │   ├── calendar.go
│   │   │   ├── webhook.go
│   │   │   └── customer_products.go
│   │   └── utils/
│   │       ├── id.go
│   │       └── session.go
│   │
│   ├── application/
│   │   ├── chatbot/
│   │   │   ├── handler.go         # Core chatbot service (466 lines)
│   │   │   └── calendar.go        # Calendar action helpers
│   │   ├── waha/handler.go
│   │   ├── calendar/service.go
│   │   └── calendar-starter/handler.go
│   │
│   └── adapters/
│       ├── inbound/
│       │   ├── http/chatbot_handler.go
│       │   └── sqs/
│       │       ├── whatsapp_handler.go
│       │       └── calendar_handler.go
│       └── outbound/
│           ├── ai/gemini.go           # Gemini API (270 lines)
│           ├── persistence/dynamodb.go # DynamoDB (386 lines)
│           ├── calendar/google_calendar.go # Google Calendar (279 lines)
│           ├── whatsapp/waha.go
│           └── sqs/sqs.go
│
├── prompts/                       # AI system prompts
│   └── mate-amigo.md              # Example prompt template
│
├── go.mod
├── go.sum
└── README.md
```

## Key Files Reference

| File | Purpose | Lines |
|------|---------|-------|
| `internal/application/chatbot/handler.go` | Core chatbot processing and action handling | 466 |
| `internal/adapters/outbound/persistence/dynamodb.go` | DynamoDB client and all repositories | 386 |
| `internal/adapters/outbound/calendar/google_calendar.go` | Google Calendar API integration | 279 |
| `internal/adapters/outbound/ai/gemini.go` | Gemini API with function declarations | 270 |
| `internal/adapters/inbound/http/chatbot_handler.go` | HTTP webhook handler | 130 |
| `internal/ports/outbound/ai.go` | AI port interface and action types | 38 |
| `internal/ports/outbound/repository.go` | Repository interfaces | 34 |

## Development Commands

```bash
# Build all handlers
go build ./...

# Run tests (requires GEMINI_API_KEY)
GEMINI_API_KEY=xxx go test ./...

# Deploy with AWS SAM
sam build && sam deploy

# Format code
go fmt ./...

# Vet code
go vet ./...
```

## DynamoDB Tables

| Table | Primary Key | Purpose |
|-------|-------------|---------|
| `zyntec_agent_sessions` | `id` (businessPhone:customerPhone) | Conversation session state |
| `zyntec_agent_messages` | `session_id` + `timestamp` (SK) | Message history |
| `zyntec_agent_customers` | `business_phone_number` | Business configuration |
| `zyntec_agent_calendars` | `customer_id` | OAuth tokens and timezone |
| `zyntec_agent_calendar_appointments` | `customer_phone_number` | Scheduled appointments |
| `zyntec_agent_business_products` | `business_phone_number` | Product catalog |

## Environment Variables

### chatbot-api-handler
| Variable | Required | Description |
|----------|----------|-------------|
| `GEMINI_API_KEY` | Yes | Google Gemini API key |
| `GOOGLE_CLIENT_ID` | Yes | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | Yes | Google OAuth client secret |
| `AWS_REGION` | No | AWS region (default: us-east-1) |
| `WHATSAPP_SQS_URL` | No | WhatsApp SQS queue URL |
| `CALENDAR_SQS_URL` | No | Calendar SQS queue URL |

### whatsapp-sqs-handler
| Variable | Required | Description |
|----------|----------|-------------|
| `WAHA_URL` | No | WAHA endpoint (default: https://api.waha.ai/v1/send-message) |

### calendar-api-handler / calendar-sqs-handler
| Variable | Required | Description |
|----------|----------|-------------|
| `CLIENT_ID` | Yes | Google OAuth client ID |
| `CLIENT_SECRET` | Yes | Google OAuth client secret |
| `REDIRECT_URL` | Yes | OAuth redirect URL |

## Code Conventions

### Error Handling
```go
// Always wrap errors with context
if err != nil {
    return nil, fmt.Errorf("failed to get customer: %w", err)
}
```

### Logging
```go
// Use structured logging with log/slog
slog.Info("transferred to human agent", "session_id", session.ID, "reason", action.Args["reason"])
slog.Warn("action validation failed", "error", err, "action", aiResponse.Action.Type)
slog.Error("Error saving session", "session", session, "error", err)
```

### Dependency Injection
```go
// Constructor-based injection
func NewService(
    sessionRepo outbound.SessionRepository,
    messageRepo outbound.MessageRepository,
    // ... other dependencies
) *Service {
    return &Service{
        sessionRepo: sessionRepo,
        messageRepo: messageRepo,
        // ...
    }
}
```

### Interface Definitions
- Define interfaces in `internal/ports/` (not with implementations)
- Inbound ports in `ports/inbound/`
- Outbound ports in `ports/outbound/`

### Context Timeouts
```go
// Use context with timeouts for external calls
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

## Adding New AI Actions (Function Calling)

### Step 1: Define Action Type
In `internal/ports/outbound/ai.go`:
```go
const ActionMyNewAction ActionType = "my_new_action"
```

### Step 2: Add Gemini Function Declaration
In `internal/adapters/outbound/ai/gemini.go`, add to function declarations:
```go
{
    Name:        "my_new_action",
    Description: "Description of what this action does",
    Parameters: &genai.Schema{
        Type: genai.TypeObject,
        Properties: map[string]*genai.Schema{
            "param1": {Type: genai.TypeString, Description: "Parameter description"},
            "param2": {Type: genai.TypeNumber, Description: "Another parameter"},
        },
        Required: []string{"param1"},
    },
}
```

### Step 3: Implement Action Handler
In `internal/application/chatbot/handler.go`, add case to `handleAction`:
```go
case outbound.ActionMyNewAction:
    return s.handleMyNewAction(session, action)
```

Then implement the handler method:
```go
func (s *Service) handleMyNewAction(session *models.Session, action *outbound.Action) (any, error) {
    param1 := getStringArg(action.Args, "param1", "")
    // Implementation logic
    return result, nil
}
```

### Step 4: Add Response Message (Optional)
In `generateActionMessage`:
```go
case outbound.ActionMyNewAction:
    return "Action completed successfully!"
```

## Message Flow

```
User WhatsApp → WAHA webhook → chatbot-api-handler (Lambda)
                                      │
                                      ├─ Human mode? → Save message only
                                      │
                                      └─ AI mode → Gemini + Function Calling
                                                          │
                                                          ▼
                                                    Response → SQS
                                                          │
                                                          ▼
                                              whatsapp-sqs-handler
                                                          │
                                                          ▼
                                                    WAHA API → User
```

## Implemented Actions

| Action | Description | Status |
|--------|-------------|--------|
| `transfer_to_human` | Transfer to human agent | ✅ |
| `schedule_appointment` | Create calendar event | ✅ |
| `cancel_appointment` | Cancel existing appointment | ✅ |
| `reschedule_appointment` | Move appointment | ✅ |
| `send_location` | Send business location | ✅ |

## Common Patterns

### Session ID Generation
```go
// Format: businessPhone:customerPhone
sessionID := utils.GenerateSessionID(businessID, customerPhone)
```

### Human Agent Mode
- Sessions can toggle between AI and human agent mode
- `IsHumanAgent` flag in session determines mode
- `AgentTimeout` (minutes) controls auto-return to AI mode

### Repository Pattern
All data access goes through repository interfaces:
```go
type SessionRepository interface {
    Save(session models.Session) error
    Get(sessionID string) (*models.Session, error)
}
```

### Concurrent Initialization
In `main.go`, services are initialized concurrently using `errgroup`:
```go
g, gCtx := errgroup.WithContext(ctx)
g.Go(func() error { /* init service 1 */ })
g.Go(func() error { /* init service 2 */ })
if err := g.Wait(); err != nil { /* handle */ }
```

## Language Notes

- **Primary language**: Spanish (error messages, user-facing text)
- **Code/comments**: English
- **AI Prompts**: Spanish (see `prompts/mate-amigo.md`)

## Testing Notes

- No test files currently exist in the repository
- Tests require `GEMINI_API_KEY` environment variable
- Use `go test ./...` to run tests

## Important Considerations

1. **Multi-tenancy**: Each business has its own configuration in `zyntec_agent_customers`
2. **Stateless Lambda**: All state is in DynamoDB; Lambda functions are stateless
3. **Async Processing**: Message sending is decoupled via SQS for reliability
4. **Calendar OAuth**: Refresh tokens are stored in `zyntec_agent_calendars`
5. **Context Window**: AI conversation history is loaded from `zyntec_agent_messages`

## Git Workflow

- Main development happens on feature branches
- Commit messages should be descriptive
- Follow Go conventions for code formatting
