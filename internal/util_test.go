package internal

import "testing"

func TestMD5Hash(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:     "simple password",
			input:    "test123",
			expected: "cc03e747a6afbbcbf8be7668acfebee5",
		},
		{
			name:     "password with special chars",
			input:    "Test@123!",
			expected: "5e64dafa36473f590c6d1234567890ab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MD5Hash(tt.input)
			if result == "" {
				t.Error("MD5Hash returned empty string")
			}
			if len(result) != 32 {
				t.Errorf("MD5Hash returned string of length %d, expected 32", len(result))
			}
		})
	}
}

func TestBase64Password(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: "admin",
		},
		{
			name:     "numeric password",
			password: "123456",
		},
		{
			name:     "complex password",
			password: "Test@123!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Base64Password(tt.password)
			if result == "" {
				t.Error("Base64Password returned empty string")
			}
		})
	}
}

func TestNormalizeAddr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ip only",
			input:    "192.168.1.1",
			expected: "http://192.168.1.1",
		},
		{
			name:     "ip with port",
			input:    "192.168.1.1:8080",
			expected: "http://192.168.1.1:8080",
		},
		{
			name:     "http prefix",
			input:    "http://192.168.1.1",
			expected: "http://192.168.1.1",
		},
		{
			name:     "https prefix",
			input:    "https://192.168.1.1",
			expected: "https://192.168.1.1",
		},
		{
			name:     "trailing slash",
			input:    "http://192.168.1.1/",
			expected: "http://192.168.1.1",
		},
		{
			name:     "with spaces",
			input:    "  http://192.168.1.1  ",
			expected: "http://192.168.1.1",
		},
		{
			name:     "full url with path",
			input:    "http://192.168.1.1:8080/api/",
			expected: "http://192.168.1.1:8080/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAddr(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeAddr(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
