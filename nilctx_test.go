package ikuaiapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNilContextTreatedAsBackground pins the oekit matrix contract
// nil_ctx_treated_as_background (contextx.OrBackground, pkg#529): a nil ctx
// behaves like context.Background instead of panicking deep in
// context.WithTimeout / ctx.Done with a stack far from the mistake.
func TestNilContextTreatedAsBackground(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"ok":true}}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL)
	raw, err := c.Get(nil, "/x", nil) //nolint:staticcheck // nil ctx is the contract under test
	if err != nil {
		t.Fatalf("Get with nil ctx: %v", err)
	}
	if !strings.Contains(string(raw), "ok") {
		t.Errorf("payload = %s", raw)
	}
}

// TestCloseNilContext: Close dereferenced ctx.Err directly, so Close(nil)
// panicked. It must instead treat nil as Background and return nil.
func TestCloseNilContext(t *testing.T) {
	c, _ := NewClient("http://127.0.0.1:1")
	if err := c.Close(nil); err != nil { //nolint:staticcheck // nil ctx is the contract under test
		t.Fatalf("Close with nil ctx: %v", err)
	}
}
