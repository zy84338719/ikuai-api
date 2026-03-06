package ikuaisdk

import (
	"context"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://192.168.1.1", "admin", "password")

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.baseURL != "http://192.168.1.1" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, "http://192.168.1.1")
	}

	if client.username != "admin" {
		t.Errorf("username = %q, want %q", client.username, "admin")
	}

	if client.version != VersionUnknown {
		t.Errorf("version should be unknown before login")
	}

	if client.IsLoggedIn() {
		t.Error("client should not be logged in initially")
	}

	client.Close()
}

func TestNewClientWithNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"192.168.1.1", "http://192.168.1.1"},
		{"192.168.1.1/", "http://192.168.1.1"},
		{"  192.168.1.1  ", "http://192.168.1.1"},
		{"http://192.168.1.1/", "http://192.168.1.1"},
	}

	for _, tt := range tests {
		client := NewClient(tt.input, "admin", "password")
		if client.baseURL != tt.expected {
			t.Errorf("NewClient(%q).baseURL = %q, want %q", tt.input, client.baseURL, tt.expected)
		}
		client.Close()
	}
}

func TestClientWithOptions(t *testing.T) {
	timeout := 60 * time.Second
	client := NewClient("http://192.168.1.1", "admin", "password",
		WithTimeout(timeout),
		WithInsecureSkipVerify(true),
	)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	// Verify logger is set
	if client.logger == nil {
		t.Error("logger should be set")
	}

	// Verify metrics is set
	if client.metrics == nil {
		t.Error("metrics should be set")
	}

	client.Close()
}

func TestClientGetVersion(t *testing.T) {
	client := NewClient("http://192.168.1.1", "admin", "password")
	defer client.Close()

	if client.GetVersion() != VersionUnknown {
		t.Error("version should be unknown before login")
	}
}

func TestClientIsLoggedIn(t *testing.T) {
	client := NewClient("http://192.168.1.1", "admin", "password")
	defer client.Close()

	if client.IsLoggedIn() {
		t.Error("client should not be logged in initially")
	}
}

func TestClientCallNotLoggedIn(t *testing.T) {
	client := NewClient("http://192.168.1.1", "admin", "password")
	defer client.Close()

	ctx := context.Background()
	var result interface{}
	err := client.Call(ctx, "test", "show", nil, &result)

	if err == nil {
		t.Error("Call should return error when not logged in")
	}

	if !IsSDKError(err) {
		t.Errorf("Error should be SDKError, got %T", err)
	}

	if GetErrorCode(err) != ErrCodeNotLoggedIn {
		t.Errorf("Error code = %d, want %d", GetErrorCode(err), ErrCodeNotLoggedIn)
	}
}

func TestClientLoginWithoutServer(t *testing.T) {
	client := NewClient("http://127.0.0.1:12345", "admin", "password",
		WithTimeout(1*time.Second),
	)
	defer client.Close()

	ctx := context.Background()
	err := client.Login(ctx)

	if err == nil {
		t.Error("Login should fail when server is not available")
	}

	if !IsSDKError(err) {
		t.Errorf("Error should be SDKError, got %T", err)
	}
}

func TestNewClientWithLoginWithoutServer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := NewClientWithLoginContext(ctx, "http://127.0.0.1:12345", "admin", "password",
		WithTimeout(1*time.Second),
	)

	if err == nil {
		client.Close()
		t.Error("NewClientWithLogin should fail when server is not available")
	}

	if client != nil {
		t.Error("Client should be nil on error")
	}
}

func TestClientClose(t *testing.T) {
	client := NewClient("http://192.168.1.1", "admin", "password")

	// Close should not panic
	client.Close()
	client.Close() // Second close should also be safe
}
