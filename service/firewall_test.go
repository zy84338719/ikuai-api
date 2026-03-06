package service

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestFirewallService_GetACL(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"results": map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":       1,
						"src_addr": "192.168.1.0/24",
						"dst_addr": "any",
						"protocol": "all",
						"enabled":  "yes",
						"comment":  "Allow LAN to WAN",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	client, server := setupTestClient(handler)
	defer server.Close()

	svc := NewFirewallService(client)
	ctx := context.Background()

	acl, err := svc.GetACL(ctx)
	if err != nil {
		t.Fatalf("GetACL() error = %v", err)
	}

	if len(acl) != 1 {
		t.Fatalf("GetACL() returned %d items, want 1", len(acl))
	}

	if acl[0].Comment != "Allow LAN to WAN" {
		t.Errorf("GetACL() Comment = %s, want 'Allow LAN to WAN'", acl[0].Comment)
	}
}

func TestFirewallService_GetDNAT(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"results": map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":       1,
						"protocol": "tcp",
						"wan_port": "8080",
						"lan_addr": "192.168.1.100",
						"lan_port": "80",
						"comment":  "Web server port forward",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	client, server := setupTestClient(handler)
	defer server.Close()

	svc := NewFirewallService(client)
	ctx := context.Background()

	dnat, err := svc.GetDNAT(ctx)
	if err != nil {
		t.Fatalf("GetDNAT() error = %v", err)
	}

	if len(dnat) != 1 {
		t.Fatalf("GetDNAT() returned %d items, want 1", len(dnat))
	}

	if dnat[0].Protocol != "tcp" {
		t.Errorf("GetDNAT() Protocol = %s, want tcp", dnat[0].Protocol)
	}
}

func TestFirewallService_GetConnLimit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"results": map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":       1,
						"src_addr": "192.168.1.100",
						"conn_num": 100,
						"new_conn": 10,
						"enabled":  "yes",
						"comment":  "Limit connections",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	client, server := setupTestClient(handler)
	defer server.Close()

	svc := NewFirewallService(client)
	ctx := context.Background()

	connLimit, err := svc.GetConnLimit(ctx)
	if err != nil {
		t.Fatalf("GetConnLimit() error = %v", err)
	}

	if len(connLimit) != 1 {
		t.Fatalf("GetConnLimit() returned %d items, want 1", len(connLimit))
	}

	if connLimit[0].ConnNum != 100 {
		t.Errorf("GetConnLimit() ConnNum = %d, want 100", connLimit[0].ConnNum)
	}
}
