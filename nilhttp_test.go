package ikuaiapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewClientRejectsNilHTTPClient: WithHTTPClient(nil) overwrites the
// default client, then NewClient dereferenced c.HTTPClient.Transport —
// a nil-pointer panic deep in the constructor, far from the option that
// caused it. Constructor-time validation must fail fast with an error
// (oekit 构造期 fail-fast, pkg#356 style), not panic.
func TestNewClientRejectsNilHTTPClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("NewClient panicked on WithHTTPClient(nil): %v", r)
		}
	}()
	c, err := NewClient(srv.URL, WithHTTPClient(nil))
	if err == nil {
		t.Fatalf("WithHTTPClient(nil) should fail fast, got client %#v", c)
	}
	if c != nil {
		t.Errorf("client should be nil on error, got %#v", c)
	}
}
