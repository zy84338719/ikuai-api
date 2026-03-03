package types_test

import (
	"encoding/json"
	"testing"

	"github.com/zy84338719/ikuai-api/types"
)

func TestBaseResponseIsV4(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected bool
	}{
		{
			name:     "v3 response",
			json:     `{"Result": 10000, "ErrMsg": ""}`,
			expected: false,
		},
		{
			name:     "v4 response",
			json:     `{"code": 0, "message": "success"}`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp types.BaseResponse
			if err := json.Unmarshal([]byte(tt.json), &resp); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}
			if got := resp.IsV4(); got != tt.expected {
				t.Errorf("BaseResponse.IsV4() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBaseResponseIsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected bool
	}{
		{
			name:     "v3 login success",
			json:     `{"Result": 10000}`,
			expected: true,
		},
		{
			name:     "v3 call success",
			json:     `{"Result": 30000}`,
			expected: true,
		},
		{
			name:     "v3 failure",
			json:     `{"Result": 10014, "ErrMsg": "failed"}`,
			expected: false,
		},
		{
			name:     "v4 success",
			json:     `{"code": 0, "message": "success"}`,
			expected: true,
		},
		{
			name:     "v4 failure",
			json:     `{"code": 10014, "message": "failed"}`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp types.BaseResponse
			if err := json.Unmarshal([]byte(tt.json), &resp); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}
			if got := resp.IsSuccess(); got != tt.expected {
				t.Errorf("BaseResponse.IsSuccess() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBaseResponseGetErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected string
	}{
		{
			name:     "v3 error message",
			json:     `{"Result": 10014, "ErrMsg": "password error"}`,
			expected: "password error",
		},
		{
			name:     "v4 error message",
			json:     `{"code": 10014, "message": "login failed"}`,
			expected: "login failed",
		},
		{
			name:     "no error message",
			json:     `{"Result": 10000}`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp types.BaseResponse
			if err := json.Unmarshal([]byte(tt.json), &resp); err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}
			if got := resp.GetErrorMessage(); got != tt.expected {
				t.Errorf("BaseResponse.GetErrorMessage() = %v, want %v", got, tt.expected)
			}
		})
	}
}
