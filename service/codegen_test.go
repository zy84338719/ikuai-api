package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ikuaiapi "github.com/zy84338719/ikuai-api"
)

// TestEveryEndpointReachable verifies that every catalog entry can be
// resolved via a mock server, catching typos in generated paths.
func TestEveryEndpointReachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"data":[],"total":0}}`))
	}))
	defer srv.Close()

	c, err := ikuaiapi.NewClient(srv.URL, ikuaiapi.WithToken("x"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	api := NewAPIClient(c)

	for _, ep := range ikuaiapi.V4EndpointCatalog {
		raw, err := api.Call(context.Background(), ep.Group, ep.Name, "GET", nil, nil)
		if err != nil {
			t.Errorf("Call %s/%s: %v", ep.Group, ep.Name, err)
			continue
		}
		if len(raw) == 0 {
			t.Errorf("Call %s/%s: empty payload", ep.Group, ep.Name)
		}
	}
}

func TestAPIClientGroups(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{}}`))
	}))
	defer srv.Close()

	c, _ := ikuaiapi.NewClient(srv.URL, ikuaiapi.WithToken("x"))
	api := NewAPIClient(c)

	if api.Network() == nil {
		t.Fatal("Network() nil")
	}
	if api.Monitoring() == nil {
		t.Fatal("Monitoring() nil")
	}
	if api.System() == nil {
		t.Fatal("System() nil")
	}
	if api.Client() != c {
		t.Error("Client() should return underlying ikuaiapi.Client")
	}
}

func TestListWithPagination(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"code":0,"data":[{"id":1}]}`))
	}))
	defer srv.Close()

	c, _ := ikuaiapi.NewClient(srv.URL, ikuaiapi.WithToken("x"))
	api := NewAPIClient(c)
	raw, err := api.Network().ListNetworkDhcpServices(context.Background(), &NetworkDhcpServicesListOptions{
		Page:     1,
		PageSize: 50,
		Order:    "desc",
		OrderBy:  "id",
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotQuery == "" {
		t.Error("expected non-empty query")
	}
	for _, want := range []string{"page=1", "page_size=50", "order=desc", "order_by=id"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query %q missing %q", gotQuery, want)
		}
	}
	var list []map[string]any
	if err := json.Unmarshal(raw, &list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 item, got %d", len(list))
	}
}

func TestCreateReturnsRowID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"message":"created","rowid":99}`))
	}))
	defer srv.Close()

	c, _ := ikuaiapi.NewClient(srv.URL, ikuaiapi.WithToken("x"))
	api := NewAPIClient(c)
	id, err := api.Auth().CreateAuthUsers(context.Background(), map[string]any{"username": "alice"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id != 99 {
		t.Errorf("id = %d, want 99", id)
	}
}

func TestDeleteUsesIDQuery(t *testing.T) {
	var gotQuery, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		gotMethod = r.Method
		_, _ = w.Write([]byte(`{"code":0,"message":"ok"}`))
	}))
	defer srv.Close()

	c, _ := ikuaiapi.NewClient(srv.URL, ikuaiapi.WithToken("x"))
	api := NewAPIClient(c)
	if err := api.Auth().DeleteAuthUsers(context.Background(), 7); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if gotMethod != "DELETE" {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if gotQuery != "id=7" {
		t.Errorf("query = %q, want id=7", gotQuery)
	}
}

func TestCatalogGroupingConsistent(t *testing.T) {
	for _, ep := range ikuaiapi.V4EndpointCatalog {
		if ep.Group == "" || ep.Name == "" || ep.Path == "" {
			t.Errorf("incomplete entry: %+v", ep)
		}
		if len(ep.Methods) == 0 {
			t.Errorf("no methods for %s/%s", ep.Group, ep.Name)
		}
	}
}

func bodyToString(r *http.Request) string {
	buf := make([]byte, 0, 1024)
	tmp := make([]byte, 512)
	for {
		n, err := r.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	return string(buf)
}
