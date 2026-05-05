package ikuaisdk

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zy84338719/ikuai-api/types"
)

func mockServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler != nil {
			handler(w, r)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"Result": 10000,
			"ErrMsg": "",
		})
	}))
}

func TestClientLoginMock(t *testing.T) {
	server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Action/login" {
			http.NotFound(w, r)
			return
		}

		var loginReq types.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
			t.Fatalf("decode login request: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		if loginReq.Username == "admin" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"Result": 10000})
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"Result": 50000,
			"ErrMsg": "Invalid credentials",
		})
	})
	defer server.Close()

	client := NewClient(server.URL, "admin", "password", WithTimeout(5*time.Second))
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Login(ctx); err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if !client.IsLoggedIn() {
		t.Error("Client should be logged in after successful login")
	}
	if client.GetVersion() != VersionV3 {
		t.Errorf("Version should be v3, got %v", client.GetVersion())
	}
}

func TestClientLoginFailMock(t *testing.T) {
	server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Action/login" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"Result": 50000,
				"ErrMsg": "Invalid credentials",
			})
		}
	})
	defer server.Close()

	client := NewClient(server.URL, "wrong", "credentials", WithTimeout(5*time.Second))
	defer client.Close()

	err := client.Login(context.Background())
	if err == nil {
		t.Fatal("Login should fail with wrong credentials")
	}

	if code := GetErrorCode(err); code != ErrCodeLoginFailed {
		t.Errorf("Error code should be ErrCodeLoginFailed, got %d", code)
	}
}

func TestClientCallMock(t *testing.T) {
	server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Action/login":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"Result": 10000})
		case "/Action/call":
			var req types.BaseRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode call request: %v", err)
			}

			if req.FuncName == "monitor_lanip" && req.Action == "show" {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"Result": 10000,
					"Data": map[string]interface{}{
						"data": []map[string]interface{}{
							{"ip_addr": "192.168.1.10", "mac": "aa:bb:cc:dd:ee:ff"},
						},
					},
				})
				return
			}
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	})
	defer server.Close()

	client, err := NewClientWithLogin(server.URL, "admin", "password", WithTimeout(5*time.Second))
	if err != nil {
		t.Fatalf("NewClientWithLogin failed: %v", err)
	}
	defer client.Close()

	var resp struct {
		types.BaseResponse
		Data struct {
			Data []types.MonitorLanIPItem `json:"data"`
		} `json:"Data"`
	}
	if err := client.Call(context.Background(), "monitor_lanip", "show", nil, &resp); err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if len(resp.Data.Data) != 1 {
		t.Fatalf("LAN IP list length = %d, want 1", len(resp.Data.Data))
	}
}

func TestClientV4Detection(t *testing.T) {
	server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Action/login" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    0,
				"message": "success",
			})
		}
	})
	defer server.Close()

	client, err := NewClientWithLogin(server.URL, "admin", "password")
	if err != nil {
		t.Fatalf("NewClientWithLogin failed: %v", err)
	}
	defer client.Close()

	if client.GetVersion() != VersionV4 {
		t.Errorf("Version should be v4 for this response, got %v", client.GetVersion())
	}
}

func TestClientLogout(t *testing.T) {
	server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Action/login":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"Result": 10000})
		case "/Action/logout":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"Result": 10000})
		default:
			http.NotFound(w, r)
		}
	})
	defer server.Close()

	client, err := NewClientWithLogin(server.URL, "admin", "password")
	if err != nil {
		t.Fatalf("NewClientWithLogin failed: %v", err)
	}
	defer client.Close()

	if !client.IsLoggedIn() {
		t.Error("Client should be logged in")
	}

	if err := client.Logout(context.Background()); err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	if client.IsLoggedIn() {
		t.Error("Client should not be logged in after logout")
	}
}

func TestClientContextCancellation(t *testing.T) {
	server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"Result": 10000})
	})
	defer server.Close()

	client := NewClient(server.URL, "admin", "password", WithTimeout(5*time.Second))
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := client.Login(ctx)
	if err == nil {
		t.Fatal("Login should have been cancelled")
	}
	if !errors.Is(err, context.DeadlineExceeded) && GetErrorCode(err) != ErrCodeRequestFailed {
		t.Errorf("expected cancellation/request error, got %v", err)
	}
}
