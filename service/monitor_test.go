package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ikuaisdk "github.com/zy84338719/ikuai-api"
)

func setupTestClient(handler http.HandlerFunc) (*ikuaisdk.Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	client := ikuaisdk.NewClient(server.URL, "admin", "password")
	client.Login(context.Background())
	return client, server
}

func TestMonitorService_GetLanIP(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"results": map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":          1,
						"ip_addr":     "192.168.1.100",
						"mac":         "00:11:22:33:44:55",
						"hostname":    "test-device",
						"interface":   "lan1",
						"upload":      1024,
						"download":    2048,
						"connect_num": 10,
						"client_type": "PC",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	client, server := setupTestClient(handler)
	defer server.Close()

	svc := NewMonitorService(client)
	ctx := context.Background()

	devices, err := svc.GetLanIP(ctx)
	if err != nil {
		t.Fatalf("GetLanIP() error = %v", err)
	}

	if len(devices) != 1 {
		t.Fatalf("GetLanIP() returned %d devices, want 1", len(devices))
	}

	if devices[0].IPAddr != "192.168.1.100" {
		t.Errorf("GetLanIP() IP = %s, want 192.168.1.100", devices[0].IPAddr)
	}

	if devices[0].Hostname != "test-device" {
		t.Errorf("GetLanIP() Hostname = %s, want test-device", devices[0].Hostname)
	}
}

func TestMonitorService_GetInterfaces(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"results": map[string]interface{}{
				"iface_check": []map[string]interface{}{
					{
						"id":         1,
						"interface":  "wan1",
						"ip_addr":    "192.168.1.1",
						"gateway":    "192.168.1.254",
						"internet":   "1",
						"updatetime": "2024-01-01 12:00:00",
					},
				},
				"iface_stream": []map[string]interface{}{
					{
						"interface":   "wan1",
						"ip_addr":     "192.168.1.1",
						"connect_num": "50",
						"upload":      10240,
						"download":    20480,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	client, server := setupTestClient(handler)
	defer server.Close()

	svc := NewMonitorService(client)
	ctx := context.Background()

	ifaces, err := svc.GetInterfaces(ctx)
	if err != nil {
		t.Fatalf("GetInterfaces() error = %v", err)
	}

	checks := ifaces.GetIFaceCheck()
	if len(checks) != 1 {
		t.Fatalf("GetInterfaces() returned %d iface_check, want 1", len(checks))
	}

	if checks[0].Interface != "wan1" {
		t.Errorf("GetInterfaces() Interface = %s, want wan1", checks[0].Interface)
	}
}

func TestMonitorService_GetSystem(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"results": map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":                1,
						"timestamp":         1704067200,
						"cpu":               25.5,
						"memory":            1024000,
						"memory_use":        50,
						"conn_num":          100,
						"wired_terminal":    10,
						"wireless_terminal": 5,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	client, server := setupTestClient(handler)
	defer server.Close()

	svc := NewMonitorService(client)
	ctx := context.Background()

	system, err := svc.GetSystem(ctx)
	if err != nil {
		t.Fatalf("GetSystem() error = %v", err)
	}

	if len(system) != 1 {
		t.Fatalf("GetSystem() returned %d items, want 1", len(system))
	}

	if system[0].CPU != 25.5 {
		t.Errorf("GetSystem() CPU = %f, want 25.5", system[0].CPU)
	}
}

func TestMonitorService_GetARP(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"results": map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":        1,
						"interface": "lan1",
						"ip_addr":   "192.168.1.100",
						"mac":       "00:11:22:33:44:55",
						"comment":   "Test device",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	client, server := setupTestClient(handler)
	defer server.Close()

	svc := NewMonitorService(client)
	ctx := context.Background()

	arp, err := svc.GetARP(ctx)
	if err != nil {
		t.Fatalf("GetARP() error = %v", err)
	}

	if len(arp) != 1 {
		t.Fatalf("GetARP() returned %d items, want 1", len(arp))
	}

	if arp[0].Mac != "00:11:22:33:44:55" {
		t.Errorf("GetARP() Mac = %s, want 00:11:22:33:44:55", arp[0].Mac)
	}
}
