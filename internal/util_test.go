package internal

import (
	"testing"
)

func TestMD5Hash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"password", "5f4dcc3b5aa765d61d8327deb882cf99"},
		{"admin", "21232f297a57a5a743894a0e4a801fc3"},
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := MD5Hash(tt.input)
			if result != tt.expected {
				t.Errorf("MD5Hash(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBase64Password(t *testing.T) {
	// The function uses salt_11 prefix
	result := Base64Password("test")
	if result == "" {
		t.Error("Base64Password should not return empty string")
	}
	// Should be deterministic
	result2 := Base64Password("test")
	if result != result2 {
		t.Error("Base64Password should be deterministic")
	}
}

func TestBase64Encode(t *testing.T) {
	tests := []struct {
		input    string
		contains string // Check that result contains expected chars
	}{
		{"hello", ""},
		{"test123", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Base64Encode(tt.input)
			// Just verify it doesn't panic and returns something for non-empty input
			if tt.input != "" && result == "" {
				t.Error("Base64Encode should not return empty for non-empty input")
			}
		})
	}
}

func TestNormalizeAddr(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"192.168.1.1", "http://192.168.1.1"},
		{"http://192.168.1.1", "http://192.168.1.1"},
		{"https://192.168.1.1", "https://192.168.1.1"},
		{"192.168.1.1/", "http://192.168.1.1"},
		{" 192.168.1.1 ", "http://192.168.1.1"},
		{"http://192.168.1.1/", "http://192.168.1.1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeAddr(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeAddr(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
