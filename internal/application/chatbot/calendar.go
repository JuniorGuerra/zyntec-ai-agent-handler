package chatbot

import (
	"app/internal/domain/models"
	"app/internal/ports/outbound"
	"fmt"
	"time"
)

func (s *Service) ProcessAction(customerID string, action *outbound.Action) any {
	if action == nil {
		return nil
	}

	switch action.Type {
	case outbound.ActionScheduleAppointment:
		return s.sendCalendarEvent(customerID, action)
	case outbound.ActionCancelAppointment:
		return s.sendCancelEvent(customerID, action)
	case outbound.ActionRescheduleAppointment:
		return s.sendRescheduleEvent(customerID, action)
	default:
		return nil
	}
}

func (s *Service) sendCalendarEvent(customerID string, action *outbound.Action) *models.CalendarEventRequest {
	calendarReq := &models.CalendarEventRequest{
		Action:      models.CalendarActionSchedule,
		CustomerID:  customerID,
		Title:       getStringArg(action.Args, "service", "Cita"),
		Description: getStringArg(action.Args, "notes", ""),
		StartTime:   buildStartTime(action.Args, "date", "time"),
		EndTime:     buildEndTime(action.Args, "date", "time"),
	}

	return calendarReq
}

func (s *Service) sendCancelEvent(customerID string, action *outbound.Action) *models.CancelEventRequest {
	cancelReq := &models.CancelEventRequest{
		Action:     models.CalendarActionCancel,
		CustomerID: customerID,
		Date:       getStringArg(action.Args, "date", ""),
		Time:       getStringArg(action.Args, "time", ""),
		Reason:     getStringArg(action.Args, "reason", ""),
	}

	return cancelReq
}

func (s *Service) sendRescheduleEvent(customerID string, action *outbound.Action) *models.RescheduleEventRequest {
	rescheduleReq := &models.RescheduleEventRequest{
		Action:       models.CalendarActionReschedule,
		CustomerID:   customerID,
		OriginalDate: getStringArg(action.Args, "original_date", ""),
		OriginalTime: getStringArg(action.Args, "original_time", ""),
		NewDate:      getStringArg(action.Args, "new_date", ""),
		NewTime:      getStringArg(action.Args, "new_time", ""),
		Reason:       getStringArg(action.Args, "reason", ""),
	}

	return rescheduleReq

}

func getStringArg(args map[string]any, key, defaultVal string) string {
	if val, ok := args[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultVal
}

func buildStartTime(args map[string]any, dateKey, timeKey string) string {
	date := getStringArg(args, dateKey, "")
	timeStr := getStringArg(args, timeKey, "09:00")
	if date == "" {
		return ""
	}
	return fmt.Sprintf("%sT%s:00", date, timeStr)
}

func buildEndTime(args map[string]any, dateKey, timeKey string) string {
	date := getStringArg(args, dateKey, "")
	timeStr := getStringArg(args, timeKey, "09:00")
	if date == "" {
		return ""
	}
	endHour := addOneHour(timeStr)
	return fmt.Sprintf("%sT%s:00", date, endHour)
}

func addOneHour(timeStr string) string {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return "10:00"
	}
	return t.Add(time.Hour).Format("15:04")
}
