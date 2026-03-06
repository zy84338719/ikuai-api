package utils

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestValidateIPRange(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{"valid localhost", "127.0.0.1", false},
		{"valid private", "192.168.1.1", false},
		{"invalid", "256.1.1.1", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPRange(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIPRange(%q) error = %v, wantErr %v", tt.ip, err, tt.wantErr)
			}
		})
	}
}

func TestValidateMACAddress(t *testing.T) {
	tests := []struct {
		name    string
		mac     string
		wantErr bool
	}{
		{"valid colon", "00:11:22:33:44:55", false},
		{"valid dash", "00-11-22-33-44-55", false},
		{"invalid format", "invalid", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMACAddress(tt.mac)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMACAddress(%q) error = %v, wantErr %v", tt.mac, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"valid http", 80, false},
		{"valid https", 443, false},
		{"invalid negative", -1, true},
		{"invalid too high", 65536, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePort(%d) error = %v, wantErr %v", tt.port, err, tt.wantErr)
			}
		})
	}
}

func TestRetryWithBackoff(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		attempts := 0
		err := RetryWithBackoff(context.Background(), 3, 10*time.Millisecond, func() error {
			attempts++
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("fail all attempts with backoff", func(t *testing.T) {
		attempts := 0
		err := RetryWithBackoff(context.Background(), 3, 5*time.Millisecond, func() error {
			attempts++
			return errors.New("persistent error")
		})
		if err == nil {
			t.Error("expected error, got nil")
		}
		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
	})
}
