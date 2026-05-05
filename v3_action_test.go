package ikuaisdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestV3ActionClientUnwrapsData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Action/login":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"Result": 10000})
		case "/Action/call":
			var req V3ActionRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if req.FuncName != "monitor_lanip" || req.Action != "show" {
				t.Fatalf("request = %#v, want monitor_lanip/show", req)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"Result": 10000,
				"Data": map[string]interface{}{
					"data": []map[string]interface{}{
						{"ip_addr": "192.168.1.10"},
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewV3ClientWithLogin(server.URL, "admin", "password")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	defer client.Close()

	v3 := client.V3Action()
	var result []map[string]string
	if err := v3.Show(context.Background(), "clients-online", nil, &result); err != nil {
		t.Fatalf("Show failed: %v", err)
	}
	if got := result[0]["ip_addr"]; got != "192.168.1.10" {
		t.Fatalf("ip_addr = %q, want 192.168.1.10", got)
	}
}

func TestV3ActionClientDryRun(t *testing.T) {
	client := NewV3Client("http://192.168.1.1", "admin", "password")
	client.loggedIn = true
	defer client.Close()

	raw, err := client.V3Action(WithV3DryRun(true)).AddRaw(context.Background(), "dhcp-static", map[string]string{"ip_addr": "192.168.1.20"})
	if err != nil {
		t.Fatalf("AddRaw dry run failed: %v", err)
	}

	var req V3ActionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("unmarshal dry run: %v", err)
	}
	if req.FuncName != "dhcp_static" || req.Action != "add" {
		t.Fatalf("request = %#v, want dhcp_static/add", req)
	}
}

func TestV3EndpointCatalogLookup(t *testing.T) {
	endpoint, ok := V3EndpointByName("clients-online")
	if !ok {
		t.Fatal("expected clients-online endpoint")
	}
	if endpoint.FuncName != "monitor_lanip" {
		t.Fatalf("FuncName = %q, want monitor_lanip", endpoint.FuncName)
	}

	endpoint, ok = V3EndpointByName("wireguard")
	if !ok {
		t.Fatal("expected wireguard compatibility marker")
	}
	if endpoint.Supported {
		t.Fatal("wireguard should be marked unsupported for v3")
	}
}

func TestV3CompatibilityForV4Catalog(t *testing.T) {
	statuses := V3CompatibilityForV4Catalog()
	if len(statuses) != len(V4EndpointCatalog) {
		t.Fatalf("statuses length = %d, want %d", len(statuses), len(V4EndpointCatalog))
	}

	var foundClientsOnline bool
	for _, status := range statuses {
		if status.V4Endpoint.Name == "clients-online" {
			foundClientsOnline = true
			if !status.Supported || status.V3Endpoint.FuncName != "monitor_lanip" {
				t.Fatalf("clients-online status = %#v, want monitor_lanip supported", status)
			}
		}
	}
	if !foundClientsOnline {
		t.Fatal("expected clients-online in compatibility status")
	}
}
