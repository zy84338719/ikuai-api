package service

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestNetworkService_GetWan(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"results": map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":           1,
						"name":         "wan1",
						"bandeth":      "eth0",
						"internet":     1,
						"dhcp_status":  1,
						"dhcp_ip_addr": "192.168.1.100",
						"dhcp_gateway": "192.168.1.1",
						"pppoe_status": 0,
						"upload":       10240,
						"download":     20480,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	client, server := setupTestClient(handler)
	defer server.Close()

	svc := NewNetworkService(client)
	ctx := context.Background()

	wan, err := svc.GetWan(ctx)
	if err != nil {
		t.Fatalf("GetWan() error = %v", err)
	}

	if len(wan) != 1 {
		t.Fatalf("GetWan() returned %d items, want 1", len(wan))
	}

	if wan[0].Name != "wan1" {
		t.Errorf("GetWan() Name = %s, want wan1", wan[0].Name)
	}
}

func TestNetworkService_GetLan(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"results": map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":          1,
						"name":        "lan1",
						"ip_mask":     "192.168.1.1/24",
						"bandeth":     "eth1",
						"dhcp_server": 1,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	client, server := setupTestClient(handler)
	defer server.Close()

	svc := NewNetworkService(client)
	ctx := context.Background()

	lan, err := svc.GetLan(ctx)
	if err != nil {
		t.Fatalf("GetLan() error = %v", err)
	}

	if len(lan) != 1 {
		t.Fatalf("GetLan() returned %d items, want 1", len(lan))
	}

	if lan[0].Name != "lan1" {
		t.Errorf("GetLan() Name = %s, want lan1", lan[0].Name)
	}
}

func TestNetworkService_GetDDNS(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"results": map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":        1,
						"tagname":   "myddns",
						"enabled":   "yes",
						"interface": "wan1",
						"server":    "ddns.oray.com",
						"domain":    "myrouter.ddns.net",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	client, server := setupTestClient(handler)
	defer server.Close()

	svc := NewNetworkService(client)
	ctx := context.Background()

	ddns, err := svc.GetDDNS(ctx)
	if err != nil {
		t.Fatalf("GetDDNS() error = %v", err)
	}

	if len(ddns) != 1 {
		t.Fatalf("GetDDNS() returned %d items, want 1", len(ddns))
	}

	if ddns[0].Domain != "myrouter.ddns.net" {
		t.Errorf("GetDDNS() Domain = %s, want myrouter.ddns.net", ddns[0].Domain)
	}
}
