package ikuaisdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestV4RESTClientGetUsesBearerTokenAndUnwrapsData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != V4APIBase+"/monitoring/system" {
			t.Fatalf("path = %q, want %q", got, V4APIBase+"/monitoring/system")
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-123" {
			t.Fatalf("Authorization = %q, want Bearer token-123", got)
		}
		if got := r.URL.Query().Get("page"); got != "2" {
			t.Fatalf("page = %q, want 2", got)
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "success",
			"data": map[string]interface{}{
				"hostname": "ikuai",
			},
		})
	}))
	defer server.Close()

	client := NewV4RESTClient(server.URL, "token-123")
	var result map[string]string
	if err := client.Get(context.Background(), "/monitoring/system", map[string]string{"page": "2"}, &result); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got := result["hostname"]; got != "ikuai" {
		t.Fatalf("hostname = %q, want ikuai", got)
	}
}

func TestV4RESTClientUnwrapsResultsAndRowID(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(V4APIBase+"/monitoring/cpu", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"results": []map[string]interface{}{{"cpu": 10}},
		})
	})
	mux.HandleFunc(V4APIBase+"/network/vlan", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "success",
			"rowid":   42,
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewV4RESTClient(server.URL, "token-123")
	var results []map[string]int
	if err := client.Get(context.Background(), "/monitoring/cpu", nil, &results); err != nil {
		t.Fatalf("Get results failed: %v", err)
	}
	if got := results[0]["cpu"]; got != 10 {
		t.Fatalf("cpu = %d, want 10", got)
	}

	var created map[string]interface{}
	if err := client.Post(context.Background(), "/network/vlan", map[string]string{"name": "IoT"}, &created); err != nil {
		t.Fatalf("Post rowid failed: %v", err)
	}
	if got := created["rowid"].(float64); got != 42 {
		t.Fatalf("rowid = %v, want 42", got)
	}
}

func TestV4RESTClientSanitizesBareNil(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"data":{"value":nil,"text":"nil stays"}}`))
	}))
	defer server.Close()

	client := NewV4RESTClient(server.URL, "")
	var result map[string]interface{}
	if err := client.Get(context.Background(), "/system/basic/config", nil, &result); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result["value"] != nil {
		t.Fatalf("value = %v, want nil", result["value"])
	}
	if got := result["text"]; got != "nil stays" {
		t.Fatalf("text = %q, want nil stays", got)
	}
}

func TestV4EndpointCatalogLookup(t *testing.T) {
	endpoint, ok := V4EndpointByName("wireguard")
	if !ok {
		t.Fatal("expected wireguard endpoint in catalog")
	}
	if endpoint.Path != "/vpn/wireguard" {
		t.Fatalf("wireguard path = %q, want /vpn/wireguard", endpoint.Path)
	}
	if got := len(V4EndpointsByGroup("monitoring")); got == 0 {
		t.Fatal("expected monitoring endpoints")
	}
}
