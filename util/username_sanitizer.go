package util

import (
	"regexp"
	"strings"
)

func SanitizeUsername(username string) string {
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	var sanitized strings.Builder

	for _, char := range username {
		if validPattern.MatchString(string(char)) {
			sanitized.WriteRune(char)
		} else {
			sanitized.WriteRune('_')
		}
	}

	result := sanitized.String()

	if len(result) < 6 {
		result = result + strings.Repeat("_", 6-len(result))
	}

	return result
}
