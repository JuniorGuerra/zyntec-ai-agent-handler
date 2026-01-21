# Zyntec Bot

Chatbot de WhatsApp con inteligencia artificial que integra Google Gemini, soporta transferencia a agente humano y acciones automatizadas mediante function calling.

## Características

- **IA Conversacional**: Respuestas inteligentes usando Google Gemini con historial de conversación
- **Transferencia a Humano**: Cambio automático a modo agente con timeout configurable
- **Function Calling**: Acciones automatizadas (agendar citas, transferir a humano)
- **Multi-tenant**: Configuración independiente por número de negocio
- **Serverless**: Desplegado en AWS Lambda con escalabilidad automática
- **Asíncrono**: Procesamiento de mensajes desacoplado vía SQS

## Arquitectura

El proyecto implementa **Arquitectura Hexagonal** (Ports & Adapters) para separación clara de responsabilidades:

```
┌─────────────────────────────────────────────────────────────────┐
│                         ADAPTERS                                 │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────────────┐    │
│  │ HTTP Handler│  │ SQS Handler  │  │ Gemini │ DynamoDB │   │    │
│  │  (Lambda)   │  │   (Lambda)   │  │  WAHA  │   SQS    │   │    │
│  └──────┬──────┘  └──────┬───────┘  └──────────┬─────────┘   │    │
│         │                │                     │              │    │
│  ───────┴────────────────┴─────────────────────┴──────────    │    │
│                         PORTS                                 │    │
│  ┌─────────────────────────────────────────────────────────┐  │    │
│  │  ChatbotUseCase │ AIPort │ Repositories │ WhatsAppPort  │  │    │
│  └─────────────────────────────────────────────────────────┘  │    │
│                              │                                │    │
│  ────────────────────────────┴────────────────────────────    │    │
│                       APPLICATION                             │    │
│  ┌─────────────────────────────────────────────────────────┐  │    │
│  │           Chatbot Service  │  WAHA Service              │  │    │
│  └─────────────────────────────────────────────────────────┘  │    │
│                              │                                │    │
│  ────────────────────────────┴────────────────────────────    │    │
│                         DOMAIN                                │    │
│  ┌─────────────────────────────────────────────────────────┐  │    │
│  │        Session  │  Message  │  Customer  │  Webhook     │  │    │
│  └─────────────────────────────────────────────────────────┘  │    │
└─────────────────────────────────────────────────────────────────┘
```

## Stack Tecnológico

| Componente | Tecnología |
|------------|------------|
| Lenguaje | Go 1.25 |
| Runtime | AWS Lambda |
| Base de Datos | DynamoDB |
| IA | Google Gemini (gemini-2.0-flash-exp) |
| WhatsApp API | WAHA (WhatsApp HTTP API) |
| Cola de Mensajes | AWS SQS |

## Estructura del Proyecto

```
zyntec-bot/
├── cmd/
│   ├── chatbot-api-handler/       # Lambda: webhook de WhatsApp
│   │   └── main.go
│   ├── whatsapp-sqs-handler/      # Lambda: consumidor SQS
│   │   └── main.go
│   └── config/                    # Configuración por handler
│       ├── chatbot-api-config/
│       └── whatsapp-sqs-config/
│
├── internal/
│   ├── ports/                     # Interfaces (contratos)
│   │   ├── inbound/
│   │   │   └── chatbot.go         # ChatbotUseCase
│   │   └── outbound/
│   │       ├── ai.go              # AIPort, ActionType
│   │       ├── repository.go      # Repositories
│   │       ├── whatsapp.go        # WhatsAppPort
│   │       └── sqs.go             # SQSAdapter
│   │
│   ├── domain/                    # Núcleo (sin dependencias)
│   │   ├── models/
│   │   │   ├── session.go         # Estado de conversación
│   │   │   ├── message.go         # Mensaje y roles
│   │   │   ├── customer.go        # Config por negocio
│   │   │   └── webhook.go         # Payload de WAHA
│   │   └── utils/
│   │       ├── id.go              # Generación UUID
│   │       └── session.go         # ID de sesión
│   │
│   ├── application/               # Casos de uso
│   │   ├── chatbot/
│   │   │   └── handler.go         # Lógica del chatbot
│   │   └── waha/
│   │       └── handler.go         # Envío de mensajes
│   │
│   └── adapters/                  # Implementaciones
│       ├── inbound/
│       │   ├── http/
│       │   │   └── chatbot_handler.go
│       │   └── sqs/
│       │       └── whatsapp_handler.go
│       └── outbound/
│           ├── ai/
│           │   └── gemini.go
│           ├── persistence/
│           │   └── dynamodb.go
│           ├── whatsapp/
│           │   └── waha.go
│           └── sqs/
│               └── sqs.go
│
├── go.mod
├── go.sum
└── README.md
```

## Variables de Entorno

### chatbot-api-handler

| Variable | Requerida | Descripción | Default |
|----------|-----------|-------------|---------|
| `GEMINI_API_KEY` | Sí | API key de Google Gemini | - |
| `AWS_REGION` | No | Región de AWS | `us-east-1` |
| `WHATSAPP_SQS_URL` | No | URL de cola SQS | Hardcoded |

### whatsapp-sqs-handler

| Variable | Requerida | Descripción | Default |
|----------|-----------|-------------|---------|
| `WAHA_URL` | No | Endpoint de WAHA | `https://api.waha.ai/v1/send-message` |

## Tablas DynamoDB

### zyntec_agent_sessions
| Campo | Tipo | Descripción |
|-------|------|-------------|
| `id` (PK) | String | `businessPhone:customerPhone` |
| `is_human_agent` | Boolean | Modo agente activo |
| `updated_at` | Timestamp | Para calcular timeout |
| `session_expiry_at` | Timestamp | Expiración de sesión |

### zyntec_agent_messages
| Campo | Tipo | Descripción |
|-------|------|-------------|
| `session_id` (PK) | String | ID de sesión |
| `timestamp` (SK) | String | Ordenamiento cronológico |
| `message` | String | Contenido del mensaje |
| `role` | String | `user`, `model`, `function` |

### zyntec_agent_customers
| Campo | Tipo | Descripción |
|-------|------|-------------|
| `business_phone_number` (PK) | String | Número del negocio |
| `ai_prompt` | String | System instruction para Gemini |
| `agent_timeout` | Number | Minutos antes de volver a bot |
| `api_key` | String | API key de WAHA |

## Flujo de Mensajes

```
Usuario WhatsApp
      │
      ▼
   [WAHA] ──webhook──► chatbot-api-handler (Lambda)
                              │
                              ▼
                    ┌─────────────────┐
                    │ ¿Modo Humano?   │
                    └────────┬────────┘
                             │
              ┌──────────────┴──────────────┐
              ▼                             ▼
        [Sí: Guardar]               [No: Procesar]
        [solo mensaje]                     │
              │                            ▼
              │                   ┌─────────────────┐
              │                   │ Gemini AI       │
              │                   │ + Function Call │
              │                   └────────┬────────┘
              │                            │
              └────────────────────────────┤
                                           ▼
                                     [SQS Queue]
                                           │
                                           ▼
                               whatsapp-sqs-handler
                                           │
                                           ▼
                                     [WAHA API]
                                           │
                                           ▼
                                   Usuario WhatsApp
```

## Acciones de IA (Function Calling)

El modelo Gemini puede invocar acciones automatizadas. Roadmap de funcionalidades:

### Core
- [x] `transfer_to_human` - Transferir conversación a agente humano
- [x] `schedule_appointment` - Agendar cita

### Comunicación
- [ ] `send_location` - Enviar ubicación del negocio/sucursal
- [ ] `send_catalog` - Enviar lista de productos/servicios con precios
- [ ] `send_document` - Enviar PDF (menú, brochure, contrato)
- [ ] `send_reminder` - Programar recordatorio para el cliente
- [ ] `notify_business` - Alertar al negocio de algo urgente
- [ ] `send_image` - Enviar imagen de producto

### Pagos
- [ ] `generate_payment_link` - Crear link de pago (Stripe, MercadoPago)
- [ ] `check_payment_status` - Verificar si un pago fue completado
- [ ] `send_invoice` - Generar y enviar factura
- [ ] `calculate_quote` - Calcular cotización con desglose
- [ ] `apply_discount` - Validar y aplicar cupón de descuento

### Calendario
- [x] `check_availability` - Consultar horarios disponibles
- [x] `cancel_appointment` - Cancelar cita existente
- [x] `reschedule_appointment` - Mover cita a otra fecha/hora
- [x] `list_appointments` - Mostrar citas futuras del cliente
- [x] `send_calendar_invite` - Enviar .ics al email del cliente

### CRM / Clientes
- [ ] `register_lead` - Capturar datos de prospecto
- [ ] `update_customer_info` - Actualizar email, nombre, preferencias
- [ ] `get_purchase_history` - Consultar compras anteriores
- [ ] `add_customer_note` - Agregar nota interna sobre el cliente
- [ ] `tag_customer` - Clasificar cliente (VIP, nuevo, recurrente)

### Pedidos / Logística
- [ ] `create_order` - Crear pedido con productos seleccionados
- [ ] `check_order_status` - Consultar estado del pedido
- [ ] `track_shipment` - Obtener tracking de envío
- [ ] `cancel_order` - Cancelar pedido (si es posible)
- [ ] `reorder_previous` - Repetir un pedido anterior

### Inventario
- [ ] `check_stock` - Verificar disponibilidad de producto
- [ ] `reserve_product` - Reservar producto por X tiempo
- [ ] `search_product` - Buscar por características/filtros
- [ ] `get_product_details` - Info detallada de un producto

### Soporte
- [ ] `create_ticket` - Abrir ticket de soporte
- [ ] `escalate_issue` - Escalar problema a supervisor
- [ ] `search_faq` - Buscar en base de conocimiento
- [ ] `send_tutorial` - Enviar guía/video tutorial

### Flujos Combinados
- [ ] `book_and_pay` - Agendar + generar link de pago + confirmar
- [ ] `quote_and_book` - Cotizar + agendar si acepta
- [ ] `order_and_schedule_delivery` - Crear pedido + agendar entrega
- [ ] `collect_feedback` - Pedir rating + guardar reseña

### Integraciones Externas
- [ ] `find_nearest_location` - Buscar sucursal más cercana (Google Maps)
- [ ] `check_weather` - Clima para fecha de cita/evento
- [ ] `generate_qr` - QR para pago, entrada, etc.

---

### Agregar Nueva Acción

1. Definir constante en `internal/ports/outbound/ai.go`:
```go
const ActionMyNewAction ActionType = "my_new_action"
```

2. Agregar `FunctionDeclaration` en `internal/adapters/outbound/ai/gemini.go`:
```go
{
    Name:        "my_new_action",
    Description: "Descripción de la acción",
    Parameters: &genai.Schema{
        Type: genai.TypeObject,
        Properties: map[string]*genai.Schema{
            "param1": {Type: genai.TypeString, Description: "..."},
        },
    },
}
```

3. Manejar en `internal/application/chatbot/handler.go`:
```go
case outbound.ActionMyNewAction:
    // Implementar lógica
```

## Comandos

```bash
# Compilar
go build ./...

# Tests
GEMINI_API_KEY=xxx go test ./...

# Desplegar con SAM
sam build && sam deploy
```

## Configuración por Negocio

Cada negocio (número de WhatsApp) tiene su propia configuración en la tabla `zyntec_agent_customers`:

- **AIPrompt**: Instrucciones del sistema para personalizar la IA
- **ProductsInfo**: Información de productos/servicios
- **AIModel**: Modelo de Gemini a usar
- **AgentTimeout**: Minutos antes de volver a modo bot
- **IsActive**: Habilitar/deshabilitar el chatbot

## Dependencias Principales

- `github.com/aws/aws-lambda-go` - Runtime de Lambda
- `github.com/aws/aws-sdk-go-v2` - SDK de AWS (DynamoDB, SQS)
- `github.com/google/generative-ai-go` - SDK de Gemini
- `github.com/google/uuid` - Generación de UUIDs
- `golang.org/x/sync` - Utilidades de concurrencia

## Licencia

Propietario - Todos los derechos reservados
