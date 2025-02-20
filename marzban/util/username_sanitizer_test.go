package util

import (
	"regexp"
	"testing"
)

func TestEnsureValidUsername(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid username remains unchanged",
			input:    "john_doe123",
			expected: "john_doe123",
		},
		{
			name:     "Dots are replaced with underscores",
			input:    "john.doe",
			expected: "john_doe",
		},
		{
			name:     "Special characters are replaced",
			input:    "user@123",
			expected: "user_123",
		},
		{
			name:     "Short username is padded",
			input:    "ab",
			expected: "ab____",
		},
		{
			name:     "Mixed invalid characters",
			input:    "user#@.123",
			expected: "user___123",
		},
		{
			name:     "Empty string is padded",
			input:    "",
			expected: "______",
		},
		{
			name:     "Unicode characters are replaced",
			input:    "użer123",
			expected: "u_er123",
		},
		{
			name:     "Spaces are replaced",
			input:    "user name",
			expected: "user_name",
		},
		{
			name:     "Dashes are allowed",
			input:    "user-name",
			expected: "user-name",
		},
		{
			name:     "Multiple consecutive invalid chars",
			input:    "user...name",
			expected: "user___name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeUsername(tt.input)
			if result != tt.expected {
				t.Errorf("ensureValidUsername(%q) = %q; want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

// TestUsernameLength ensures the function handles various length scenarios correctly
func TestUsernameLength(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		minLen int
		maxLen int
	}{
		{
			name:   "Exactly 6 characters",
			input:  "user12",
			minLen: 6,
			maxLen: 6,
		},
		{
			name:   "Long username",
			input:  "verylongusername123",
			minLen: 6,
			maxLen: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeUsername(tt.input)
			if len(result) < tt.minLen {
				t.Errorf("Username too short: got length %d, want minimum %d",
					len(result), tt.minLen)
			}
			if len(result) > tt.maxLen {
				t.Errorf("Username too long: got length %d, want maximum %d",
					len(result), tt.maxLen)
			}
		})
	}
}

// TestRegexCompliance ensures all characters in the output comply with the regex pattern
func TestRegexCompliance(t *testing.T) {
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	tests := []string{
		"user@123",
		"john.doe",
		"użer123",
		"user name",
		"user#$%^&*()",
	}

	for _, input := range tests {
		result := SanitizeUsername(input)
		if !validPattern.MatchString(result) {
			t.Errorf("Output contains invalid characters: %q", result)
		}
	}
}
