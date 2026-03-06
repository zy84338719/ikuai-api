package types

import "testing"

func TestBaseResponseIsV4(t *testing.T) {
	tests := []struct {
		name     string
		resp     BaseResponse
		expected bool
	}{
		{
			name:     "v3 response with Result",
			resp:     BaseResponse{Result: 10000, ErrMsg: ""},
			expected: false,
		},
		{
			name:     "v4 response with message",
			resp:     BaseResponse{Code: 0, Message: "success"},
			expected: true,
		},
		{
			name:     "v3 with error",
			resp:     BaseResponse{Result: 50000, ErrMsg: "error"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.IsV4(); got != tt.expected {
				t.Errorf("IsV4() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBaseResponseIsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		resp     BaseResponse
		expected bool
	}{
		{
			name:     "v3 success 10000",
			resp:     BaseResponse{Result: 10000},
			expected: true,
		},
		{
			name:     "v3 success 30000",
			resp:     BaseResponse{Result: 30000},
			expected: true,
		},
		{
			name:     "v3 failure",
			resp:     BaseResponse{Result: 50000},
			expected: false,
		},
		{
			name:     "v4 success",
			resp:     BaseResponse{Code: 0, Message: "success"},
			expected: true,
		},
		{
			name:     "v4 failure",
			resp:     BaseResponse{Code: 1, Message: "error"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.IsSuccess(); got != tt.expected {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBaseResponseGetErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		resp     BaseResponse
		expected string
	}{
		{
			name:     "v3 error message",
			resp:     BaseResponse{ErrMsg: "v3 error"},
			expected: "v3 error",
		},
		{
			name:     "v4 error message",
			resp:     BaseResponse{Message: "v4 error"},
			expected: "v4 error",
		},
		{
			name:     "v3 message takes precedence",
			resp:     BaseResponse{ErrMsg: "v3", Message: "v4"},
			expected: "v3",
		},
		{
			name:     "no error message",
			resp:     BaseResponse{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resp.GetErrorMessage(); got != tt.expected {
				t.Errorf("GetErrorMessage() = %q, want %q", got, tt.expected)
			}
		})
	}
}
