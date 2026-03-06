package ikuaisdk

import (
	"errors"
	"testing"
)

func TestSDKError(t *testing.T) {
	tests := []struct {
		name     string
		code     ErrorCode
		message  string
		cause    error
		expected string
	}{
		{
			name:     "error with cause",
			code:     ErrCodeRequestFailed,
			message:  "request failed",
			cause:    errors.New("connection refused"),
			expected: "[SDK Error 3] request failed: connection refused",
		},
		{
			name:     "error without cause",
			code:     ErrCodeLoginFailed,
			message:  "invalid credentials",
			cause:    nil,
			expected: "[SDK Error 1] invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewSDKError(tt.code, tt.message, tt.cause)
			if err.Error() != tt.expected {
				t.Errorf("Error() = %q, want %q", err.Error(), tt.expected)
			}
		})
	}
}

func TestSDKErrorUnwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewSDKError(ErrCodeRequestFailed, "failed", cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}

	// Test with no cause
	err2 := NewSDKError(ErrCodeUnknown, "unknown", nil)
	if err2.Unwrap() != nil {
		t.Errorf("Unwrap() should return nil for no cause")
	}
}

func TestIsSDKError(t *testing.T) {
	sdkErr := NewSDKError(ErrCodeLoginFailed, "login failed", nil)
	normalErr := errors.New("normal error")

	if !IsSDKError(sdkErr) {
		t.Error("IsSDKError should return true for SDKError")
	}
	if IsSDKError(normalErr) {
		t.Error("IsSDKError should return false for normal error")
	}
}

func TestGetErrorCode(t *testing.T) {
	sdkErr := NewSDKError(ErrCodeNotLoggedIn, "not logged in", nil)
	normalErr := errors.New("normal error")

	if code := GetErrorCode(sdkErr); code != ErrCodeNotLoggedIn {
		t.Errorf("GetErrorCode() = %d, want %d", code, ErrCodeNotLoggedIn)
	}
	if code := GetErrorCode(normalErr); code != ErrCodeUnknown {
		t.Errorf("GetErrorCode() = %d, want %d for normal error", code, ErrCodeUnknown)
	}
}

func TestErrorCodeValues(t *testing.T) {
	// Test that error codes have expected values
	codes := map[ErrorCode]string{
		ErrCodeUnknown:            "Unknown",
		ErrCodeLoginFailed:        "LoginFailed",
		ErrCodeNotLoggedIn:        "NotLoggedIn",
		ErrCodeRequestFailed:      "RequestFailed",
		ErrCodeInvalidResponse:    "InvalidResponse",
		ErrCodeVersionNotSupported: "VersionNotSupported",
		ErrCodeValidationFailed:   "ValidationFailed",
		ErrCodeRateLimited:        "RateLimited",
		ErrCodeTimeout:            "Timeout",
		ErrCodeConnectionLost:     "ConnectionLost",
		ErrCodeUnauthorized:       "Unauthorized",
		ErrCodeForbidden:          "Forbidden",
	}

	for code := range codes {
		if code < 0 || code > 100 {
			t.Errorf("ErrorCode %s has unexpected value %d", codes[code], code)
		}
	}
}
