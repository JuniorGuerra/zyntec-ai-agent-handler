package utils

import (
	"fmt"
	"strings"
)

func GenerateSessionID(businessPhone, customerPhone string) string {
	return fmt.Sprintf("%s:%s", businessPhone, customerPhone)
}

func ExtractBusinessPhone(sessionID string) string {
	parts := strings.Split(sessionID, ":")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}
