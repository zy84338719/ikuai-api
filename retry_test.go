package ikuaiapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestWriteNetworkErrorNotRetried verifies the regression fix: a POST that
// fails at the transport layer must NOT be retried, because the request may
// have reached the router and a retry could create a duplicate.
func TestWriteNetworkErrorNotRetried(t *testing.T) {
	c, _ := NewClient("http://127.0.0.1:1", // nothing listens → NetworkError
		WithRetry(3), WithRetryDelay(time.Millisecond, time.Millisecond))
	var calls int32
	_, err := c.Post(context.Background(), "/auth/users", map[string]any{"x": 1})
	// Wrap to count attempts indirectly: Post goes through doWithRetry which
	// calls doOnce; since the host is unreachable each attempt would increment.
	// We can't observe calls to the unreachable host, so assert via behaviour:
	// the error must be a *NetworkError and it must come back quickly (no 3x
	// backoff sleep). A 3-attempt retry with min base would still take >0 even
	// with tiny delays; here we mainly assert the error type.
	if err == nil {
		t.Fatal("expected error")
	}
	var netErr *NetworkError
	if !errors.As(err, &netErr) {
		t.Fatalf("expected *NetworkError, got %T", err)
	}
	// shouldRetry must report false for a non-idempotent verb.
	if shouldRetry(err, false) {
		t.Error("network error on a write must not be retryable")
	}
	_ = atomic.LoadInt32(&calls)
}

// TestGetNetworkErrorRetried verifies GET transport failures are still retried.
func TestGetNetworkErrorRetried(t *testing.T) {
	c, _ := NewClient("http://127.0.0.1:1",
		WithRetry(2), WithRetryDelay(time.Millisecond, time.Millisecond))
	start := time.Now()
	_, err := c.Get(context.Background(), "/x", nil)
	dur := time.Since(start)
	if err == nil {
		t.Fatal("expected error")
	}
	// With retry(2) there is one backoff sleep between the two attempts.
	// Transport failure to a closed port is near-instant, so total time is
	// dominated by the single retry delay (<100ms with these tiny values).
	if dur > 500*time.Millisecond {
		t.Errorf("GET retry took too long: %v (retry loop misbehaving?)", dur)
	}
	if !shouldRetry(err, true) {
		t.Error("network error on a GET should be retryable")
	}
}

// TestRetryOn429 verifies 429 responses are retried (previously they were not).
func TestRetryOn429(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 2 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"code":0,"message":"rate limited"}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":0,"data":{"ok":true}}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithRetry(3), WithRetryDelay(time.Millisecond, time.Millisecond))
	raw, err := c.Get(context.Background(), "/x", nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !strings.Contains(string(raw), `"ok":true`) {
		t.Errorf("expected success after 429 retry, got %s", raw)
	}
	if calls < 2 {
		t.Errorf("expected at least 2 calls, got %d", calls)
	}
}

// Test4xxNotRetried verifies a plain 4xx (not 429) is not retried.
func Test4xxNotRetried(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":0,"message":"bad request"}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithRetry(3), WithRetryDelay(time.Millisecond, time.Millisecond))
	_, err := c.Get(context.Background(), "/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("400 should not be retried, calls = %d", calls)
	}
}

// TestIsRetryable covers the public IsRetryable methods on both error types.
func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"api 500", &APIError{HTTPStatus: 500}, true},
		{"api 429", &APIError{HTTPStatus: 429}, true},
		{"api 503", &APIError{HTTPStatus: 503}, true},
		{"api 400", &APIError{HTTPStatus: 400}, false},
		{"api 403", &APIError{HTTPStatus: 403}, false},
		{"api 404", &APIError{HTTPStatus: 404}, false},
		{"api nil", (*APIError)(nil), false},
		{"network", &NetworkError{Message: "x"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch e := tt.err.(type) {
			case *APIError:
				if got := e.IsRetryable(); got != tt.want {
					t.Errorf("APIError.IsRetryable() = %v, want %v", got, tt.want)
				}
			case *NetworkError:
				if got := e.IsRetryable(); got != tt.want {
					t.Errorf("NetworkError.IsRetryable() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// TestMetricsRecorded verifies the Metrics collector is wired into doOnce.
func TestMetricsRecorded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":1}`))
	}))
	defer srv.Close()

	m := NewMetrics()
	c, _ := NewClient(srv.URL, WithMetrics(m))
	if _, err := c.Get(context.Background(), "/x", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}
	count, errs, _ := m.GetStats()
	if count != 1 {
		t.Errorf("requestCount = %d, want 1", count)
	}
	if errs != 0 {
		t.Errorf("errorCount = %d, want 0", errs)
	}

	// Now an erroring request.
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code":0,"message":"boom"}`))
	}))
	defer srv2.Close()
	m2 := NewMetrics()
	c2, _ := NewClient(srv2.URL, WithMetrics(m2), WithRetry(1))
	_, _ = c2.Get(context.Background(), "/x", nil)
	count2, errs2, _ := m2.GetStats()
	if count2 != 1 {
		t.Errorf("erroring requestCount = %d, want 1", count2)
	}
	if errs2 != 1 {
		t.Errorf("errorCount = %d, want 1", errs2)
	}
}

// TestRetryAfterHeaderParsed verifies Retry-After is parsed off a 429.
func TestRetryAfterHeaderParsed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "5")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"code":0,"message":"slow down"}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithRetry(1)) // no retry, so we get the raw error
	_, err := c.Get(context.Background(), "/x", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.RetryAfter != 5*time.Second {
		t.Errorf("RetryAfter = %v, want 5s", apiErr.RetryAfter)
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		in   string
		want time.Duration
	}{
		{"", 0},
		{"0", 0},
		{"-1", 0},
		{"3", 3 * time.Second},
		{"garbage", 0},
	}
	for _, tt := range tests {
		got := parseRetryAfter(tt.in)
		if got != tt.want {
			t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
