package calendarstarter

import (
	"app/internal/adapters/outbound/persistence"
	"app/internal/domain/models"
	"app/internal/ports/outbound"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

type CalendarStarterHandler struct {
	calendarPort outbound.CalendarPort
	dbClient     *persistence.CalendarRepository
}

func NewCalendarStarterHandler(calendarPort outbound.CalendarPort, dbClient *persistence.CalendarRepository) *CalendarStarterHandler {
	return &CalendarStarterHandler{
		calendarPort: calendarPort,
		dbClient:     dbClient,
	}
}

func (h *CalendarStarterHandler) Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	code := request.QueryStringParameters["code"]
	state := request.QueryStringParameters["state"]

	token, err := h.calendarPort.ExchangeToken(code)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}

	calendar := models.Calendar{
		CustomerID:   state,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
		CreatedAt:    time.Now(),
	}

	err = h.dbClient.Save(calendar)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
			StatusCode: 500,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 201,
	}, nil
}
