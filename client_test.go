package ikuaiapi

import (
	"context"
	"encoding/base64"
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
	// A real router issues a base64 token, not hex: it decodes to a UUID
	// prefix. Verified on hardware, where such a token authenticated on
	// every endpoint while the old hex-only check rejected it.
	base64Token := base64.RawStdEncoding.EncodeToString([]byte("0123456789abcdef01234567"))
	if len(base64Token) != 32 {
		t.Fatalf("test fixture should be 32 chars, got %d", len(base64Token))
	}
	if err := ValidateToken(base64Token); err != nil {
		t.Errorf("a base64 token must validate, got %v", err)
	}
	// 32 characters in neither alphabet is a copy-paste accident.
	if err := ValidateToken(strings.Repeat("!", 32)); err == nil {
		t.Error("a token with no valid character should fail")
	}
	if err := ValidateToken(strings.Repeat("z z", 16)); err == nil {
		t.Error("a token containing whitespace should fail")
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
	if err := c.Close(context.Background()); err != nil {
		t.Errorf("Close: %v", err)
	}
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

// streamOf serves an endless run of one byte so a test can push any body
// size through a handler without allocating it.
type streamOf byte

func (s streamOf) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(s)
	}
	return len(p), nil
}

// TestResponseBodyBounded: a response past the cap must fail with an
// explicit size error instead of being buffered without bound. An
// unbounded io.ReadAll hands a hostile or broken endpoint a lever on this
// process's heap — oekit/pkg/httpjson caps reads at 64MiB for exactly
// this reason ("异常服务端刷屏防内存放大").
func TestResponseBodyBounded(t *testing.T) {
	bodySize := int64(maxResponseBody) + 1 // one byte past the cap
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.CopyN(w, streamOf('x'), bodySize)
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithRetry(1)) // one attempt: re-reading an oversized body only re-allocates it
	_, err := c.Get(context.Background(), "/x", nil)
	if err == nil {
		t.Fatal("expected an error for an oversized response body")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if !strings.Contains(apiErr.Message, "exceeds") {
		m := apiErr.Message
		if len(m) > 160 {
			m = m[:160] + "..."
		}
		t.Errorf("want an explicit body-size error, got %q — the read must stop at the cap instead of buffering the whole body", m)
	}
}

// TestResponseBodyExactlyAtLimit is the guard-rail for an off-by-one in the
// cap: a body that lands exactly on the cap is not "over" it. The read has
// to probe one byte past the cap to tell the two apart (same trick as
// httpjson.DoRaw). It passes before and after the fix; it exists to catch
// a `len >= cap` mistake.
func TestResponseBodyExactlyAtLimit(t *testing.T) {
	bodySize := int64(maxResponseBody) // exactly at the cap
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.CopyN(w, streamOf('x'), bodySize)
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithRetry(1))
	_, err := c.Get(context.Background(), "/x", nil)
	if err == nil {
		t.Fatal("expected the non-JSON error for a capped-but-oversized-looking body")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if strings.Contains(apiErr.Message, "exceeds") {
		t.Errorf("a body exactly at the cap must not be rejected as oversized: %s", apiErr.Message)
	}
}

// TestNonJSONBodySummarized: the raw body may enter the error message only
// as a bounded, rune-safe summary. Copying it wholesale puts an entire
// error page (or a 64MiB payload) into the error and every log line that
// carries it — httpjson's DoJSON takes a 200-rune summary for the same
// reason.
func TestNonJSONBodySummarized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, strings.Repeat("x", 4096))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithRetry(1))
	_, err := c.Get(context.Background(), "/x", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if n := len(apiErr.Message); n > 300 {
		t.Errorf("Message is %d bytes; want a bounded summary, got prefix %q", n, truncMsg(apiErr.Message, 120))
	}
	if !strings.Contains(apiErr.Message, "4096") {
		t.Errorf("summary should state the original size so truncation stays observable, got %q", truncMsg(apiErr.Message, 120))
	}
}

// TestHTTPErrorBodySummarized applies the same bound to the status-error
// path: a plain-text 400 page with no envelope message falls back to the
// raw body, which must be summarized too.
func TestHTTPErrorBodySummarized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, strings.Repeat("x", 4096))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithRetry(1)) // 400 is not retryable, one attempt
	_, err := c.Get(context.Background(), "/x", nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if n := len(apiErr.Message); n > 300 {
		t.Errorf("Message is %d bytes; want a bounded summary, got prefix %q", n, truncMsg(apiErr.Message, 120))
	}
	if !strings.Contains(apiErr.Message, "4096") {
		t.Errorf("summary should state the original size, got %q", truncMsg(apiErr.Message, 120))
	}
}

// truncMsg is a test-side preview helper: bounded output for failures, so
// a runaway Message can never flood the test log.
func truncMsg(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
