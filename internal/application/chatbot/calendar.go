package chatbot

import (
	"fmt"
	"time"
)

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
