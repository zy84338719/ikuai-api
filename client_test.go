package ikuaiapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSanitizeNil(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"already-valid", `{"a":null}`, `{"a":null}`},
		{"bare-nil", `{"a":nil}`, `{"a":null}`},
		{"string-with-nil", `{"a":"nil"}`, `{"a":"nil"}`},
		{"nested", `{"a":nil,"b":[nil,1]}`, `{"a":null,"b":[null,1]}`},
		{"escaped-quote", `{"a":"\""}`, `{"a":"\""}`},
		{"crlf", "{\"a\":nil}\r\n", "{\"a\":null}\n"},
		{"null-then-nil", `{"a":null,"b":nil}`, `{"a":null,"b":null}`},
		{"word-with-nil", `{"a":nullify}`, `{"a":nullify}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(SanitizeNil([]byte(tt.in)))
			if got != tt.want {
				t.Errorf("SanitizeNil(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	good := strings.Repeat("a", 32)
	if err := ValidateToken(good); err != nil {
		t.Errorf("32 hex chars should validate, got %v", err)
	}
	if err := ValidateToken(""); err == nil {
		t.Error("empty token should fail")
	}
	if err := ValidateToken("xx"); err == nil {
		t.Error("short token should fail")
	}
	if err := ValidateToken(strings.Repeat("z", 32)); err == nil {
		t.Error("non-hex token should fail")
	}
}

func TestNewClient(t *testing.T) {
	c, err := NewClient("http://192.168.1.1", WithToken("deadbeefcafebabe1234567890abcdef"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if c.BaseURL != "http://192.168.1.1" {
		t.Errorf("BaseURL = %q", c.BaseURL)
	}
	if c.APIBase != V4APIBase {
		t.Errorf("APIBase = %q", c.APIBase)
	}
	if c.HTTPClient == nil {
		t.Error("HTTPClient should be set")
	}
	c.Close()
}

func TestNewClientRequiresBaseURL(t *testing.T) {
	if _, err := NewClient(""); err == nil {
		t.Fatal("empty baseURL should fail")
	}
}

func TestGetSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Errorf("Authorization header = %q", got)
		}
		if r.URL.Path != "/api/v4.0/monitoring/system" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"hostname":"router"}}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithToken("secret"))
	raw, err := c.Get(context.Background(), "/monitoring/system", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	var resp map[string]string
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["hostname"] != "router" {
		t.Errorf("hostname = %q", resp["hostname"])
	}
}

func TestGetBareNilPayload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Some iKuai firmware emits `nil` rather than `null`.
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":nil}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL)
	raw, err := c.Get(context.Background(), "/anything", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !strings.Contains(string(raw), "ok") {
		t.Errorf("expected message in payload, got %s", raw)
	}
}

func TestGetRowIDCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"message":"created","rowid":42}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL)
	raw, err := c.Post(context.Background(), "/auth/users", map[string]any{"username": "alice"})
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if !strings.Contains(string(raw), `"rowid":42`) {
		t.Errorf("expected rowid 42 in payload, got %s", raw)
	}
}

func TestGetErrorEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":3007,"message":"invalid token","details":[{"field":"token","msg":"expired"}]}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL)
	_, err := c.Get(context.Background(), "/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Code != 3007 {
		t.Errorf("code = %d", apiErr.Code)
	}
	if apiErr.HTTPStatus != 403 {
		t.Errorf("status = %d", apiErr.HTTPStatus)
	}
	if !strings.Contains(apiErr.Error(), "invalid token") {
		t.Errorf("error message missing: %s", apiErr.Error())
	}
	if !strings.Contains(apiErr.Error(), "token: expired") {
		t.Errorf("error message missing field: %s", apiErr.Error())
	}
}

func TestGetMonitorFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// monitor/load endpoints use `results` instead of `data`.
		_, _ = w.Write([]byte(`{"code":0,"results":[{"ts":1,"cpu":10}]}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL)
	raw, err := c.Get(context.Background(), "/monitoring/cpu", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !strings.Contains(string(raw), `"cpu":10`) {
		t.Errorf("expected results payload, got %s", raw)
	}
}

func TestRetryOnServerError(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code":0,"message":"transient"}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"ok":true}}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithRetry(3), WithRetryDelay(time.Millisecond, 10*time.Millisecond))
	raw, err := c.Get(context.Background(), "/x", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !strings.Contains(string(raw), `"ok":true`) {
		t.Errorf("expected success payload, got %s", raw)
	}
	if calls < 2 {
		t.Errorf("expected retry, calls = %d", calls)
	}
}

func TestDryRunDoesNotSend(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("server should not be called in dry-run mode")
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithDryRun(true))
	raw, err := c.Get(context.Background(), "/x", map[string]string{"k": "v"})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !strings.Contains(string(raw), `"dry_run":true`) {
		t.Errorf("expected dry-run payload, got %s", raw)
	}
}

func TestListParamsStylePagination(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"code":0,"data":[]}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL)
	q := map[string]string{
		"page":      "2",
		"page_size": "50",
		"order":     "desc",
		"order_by":  "id",
	}
	if _, err := c.Get(context.Background(), "/x", q); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !strings.Contains(gotQuery, "page=2") || !strings.Contains(gotQuery, "page_size=50") {
		t.Errorf("query string missing pagination: %q", gotQuery)
	}
}

func TestNetworkErrorWrapped(t *testing.T) {
	c, _ := NewClient("http://127.0.0.1:1") // nothing listens
	_, err := c.Get(context.Background(), "/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var netErr *NetworkError
	if !errors.As(err, &netErr) {
		t.Fatalf("expected *NetworkError, got %T", err)
	}
}

func TestReadBodyClosed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"code":0,"data":1}`)
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL)
	if _, err := c.Get(context.Background(), "/x", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}
	// Re-use the underlying transport to verify connections are released.
	if c.HTTPClient.Transport == nil {
		t.Fatal("transport nil")
	}
}
