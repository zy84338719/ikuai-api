package ikuaisdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zy84338719/ikuai-api/types"
)

// mockServer creates a test server that responds with iKuai API format
func mockServer(t *testing.T, (handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(t)

	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler != nil {
			handler(w, r)
		} else {
			// Default response
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"Result": 10000,
				"ErrMsg":  "",
			})
		}
	})

	return server
}

func TestClientLoginMock(t *testing.T) {
	server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Action/login" {
			var loginReq types.LoginRequest
			json.NewDecoder(r.Body).Decode(&loginReq)
			if loginReq.Username == "admin" {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"Result": 10000,
				})
            } else {
                w.WriteHeader(http.StatusOK)
                json.NewEncoder(w).Encode(map[string]interface{}{
                   	"Result": 50000,
                    "ErrMsg": "Invalid credentials",
                })
            }
        }
    })
    defer server.Close()

    client := NewClient(server.URL, "admin", "password",
        WithTimeout(5*time.Second),
    )
    defer client.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err := client.Login(ctx)
    if err != nil {
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
            json.NewEncoder(w).Encode(map[string]interface{}{
                "Result": 50000,
                "ErrMsg": "Invalid credentials",
            })
        }
    })
    defer server.Close()

    client := NewClient(server.URL, "wrong", "credentials",
        WithTimeout(5*time.Second),
    )
    defer client.Close()

    ctx := context.Background()
    err := client.Login(ctx)
    if err == nil {
        t.Fatal("Login should fail with wrong credentials")
    }

    if IsSDKError(err) {
        code := GetErrorCode(err)
        if code != ErrCodeLoginFailed {
            t.Errorf("Error code should be ErrCodeLoginFailed, got %d", code)
        }
    }
}

func TestClientCallMock(t *testing.T) {
    server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/Action/login" {
            w.WriteHeader(http.StatusOK)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "Result": 10000,
            })
        } else if r.URL.Path == "/Action/call" {
            var req types.BaseRequest
            json.NewDecoder(r.Body).Decode(&req)

            if req.FuncName == "monitor_lanip" && req.Action == "show" {
                w.WriteHeader(http.StatusOK)
                json.NewEncoder(w).Encode(map[string]interface{}{
                    "Result": 10000,
                    "Data": struct {
                        Data []struct {
                            IP      []string `json:"ip_addr"`
                            Mac     []string `json:"mac"`
                        } `json:"data"`,
                    },
                })
            }
        }
    })
    defer server.Close()

    client, err := NewClientWithLogin(server.URL, "admin", "password",
        WithTimeout(5*time.Second),
    )
    if err != nil {
        t.Fatalf("NewClientWithLogin failed: %v", err)
    }
    defer client.Close()

    ctx := context.Background()
    var resp struct {
        types.BaseResponse
        Data   struct {
            Data []types.MonitorLanIPItem `json:"data"`
        } `json:"Data"`
    }
    err = client.Call(ctx, "monitor_lanip", "show", nil, &resp)
    if err != nil {
        t.Fatalf("Call failed: %v", err)
    }

    if len(resp.Data.Data) == 0 {
        t.Error("Expected empty LAN IP list")
    }
}

func TestClientV4Detection(t *testing.T) {
    server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/Action/login" {
            w.WriteHeader(http.StatusOK)
            json.NewEncoder(w).Encode(map[string]interface{}{
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
        if r.URL.Path == "/Action/login" {
            w.WriteHeader(http.StatusOK)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "Result": 10000,
            })
        } else if r.URL.Path == "/Action/logout" {
            w.WriteHeader(http.StatusOK)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "Result": 10000,
            })
        }
    })
    defer server.Close()

    client, err := NewClientWithLogin(server.URL, "admin", "password",
    if err != nil {
        t.Fatalf("NewClientWithLogin failed: %v", err)
    }
    defer client.Close()

    if !client.IsLoggedIn() {
        t.Error("Client should be logged in")
    }

    ctx := context.Background()
    err := client.Logout(ctx)
    if err != nil {
        t.Fatalf("Logout failed: %v", err)
    }

    if client.IsLoggedIn() {
        t.Error("Client should not be logged in after logout")
    }
}

func TestClientContextCancellation(t *testing.T) {
    server := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
        // Slow response
        time.Sleep(2 * time.Second)
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "Result": 10000,
        })
    })
    defer server.Close()

    client := NewClient(server.URL, "admin", "password",
        WithTimeout(5*time.Second),
    )
    defer client.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()

    errChan := make(chan error, 1)
    go func() {
        err := client.Login(ctx)
        errChan <- err
    }()

    select {
    case <-err:
        t.Error("Login should have been cancelled")
    case <-ctx.Done():
        t.Error("Context was not cancelled")
    case err == nil:
        if err == context.Canceled {
            t.Errorf("Error should be context.Canceled, got %v", err)
        }
    }
}
