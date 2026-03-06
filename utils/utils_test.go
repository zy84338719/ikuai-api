package utils

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestIsValidMAC(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid with colon", "00:11:22:33:44:55", true},
		{"valid with dash", "00-11-22-33-44-55", true},
		{"invalid format", "00:11:22:33:44", false},
		{"invalid chars", "GG:11:22:33:44:55", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidMAC(tt.input); got != tt.want {
				t.Errorf("IsValidMAC(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidIPv4(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid localhost", "127.0.0.1", true},
		{"valid private", "192.168.1.1", true},
		{"valid max", "255.255.255.255", true},
		{"invalid - too high", "256.1.1.1", false},
		{"invalid - format", "1.2.3", false},
		{"invalid - empty", "", false},
		{"invalid - text", "abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidIPv4(tt.input); got != tt.want {
				t.Errorf("IsValidIPv4(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidPort(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  bool
	}{
		{"valid low", 0, true},
		{"valid http", 80, true},
		{"valid https", 443, true},
		{"valid high", 65535, true},
		{"invalid negative", -1, false},
		{"invalid too high", 65536, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidPort(tt.input); got != tt.want {
				t.Errorf("IsValidPort(%d) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestRetry(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		attempts := 0
		err := Retry(context.Background(), 3, 10*time.Millisecond, func() error {
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

	t.Run("success on second attempt", func(t *testing.T) {
		attempts := 0
		err := Retry(context.Background(), 3, 10*time.Millisecond, func() error {
			attempts++
			if attempts < 2 {
				return fmt.Errorf("temporary error")
			}
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if attempts != 2 {
			t.Errorf("expected 2 attempts, got %d", attempts)
		}
	})

	t.Run("fail all attempts", func(t *testing.T) {
		attempts := 0
		err := Retry(context.Background(), 3, 10*time.Millisecond, func() error {
			attempts++
			return fmt.Errorf("persistent error")
		})
		if err == nil {
			t.Error("expected error, got nil")
		}
		if attempts != 3 {
			t.Errorf("expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := Retry(ctx, 3, 10*time.Millisecond, func() error {
			return fmt.Errorf("error")
		})
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})
}

func TestRateLimiter(t *testing.T) {
	t.Run("basic rate limiting", func(t *testing.T) {
		rl := NewRateLimiter(100)
		defer rl.Stop()

		start := time.Now()
		for i := 0; i < 5; i++ {
			if err := rl.Wait(context.Background()); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}
		elapsed := time.Since(start)

		if elapsed < 40*time.Millisecond {
			t.Errorf("rate limiting not working: elapsed %v", elapsed)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		rl := NewRateLimiter(1)
		defer rl.Stop()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := rl.Wait(ctx)
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})

	t.Run("stop rate limiter", func(t *testing.T) {
		rl := NewRateLimiter(1)
		rl.Stop()

		err := rl.Wait(context.Background())
		if err == nil {
			t.Error("expected error after stop")
		}
	})
}
