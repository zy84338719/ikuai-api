//go:build integration

package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/service"
)

func getTestConfig() (string, string, string) {
	addr := os.Getenv("IKUAI_TEST_ADDR")
	if addr == "" {
		addr = "10.10.30.254"
	}
	username := os.Getenv("IKUAI_TEST_USERNAME")
	if username == "" {
		username = "zhangyi"
	}
	password := os.Getenv("IKUAI_TEST_PASSWORD")
	if password == "" {
		password = "REDACTED"
	}
	return addr, username, password
}

func TestIntegrationAllAPIs(t *testing.T) {
	addr, username, password := getTestConfig()

	client, err := ikuaisdk.NewClientWithLogin(addr, username, password, ikuaisdk.WithTimeout(10*time.Second))
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer client.Close()

	t.Logf("Connected to iKuai %s", client.GetVersion())
	ctx := context.Background()

	t.Run("Monitor_LanIP", func(t *testing.T) {
		svc := service.NewMonitorService(client)
		items, err := svc.GetLanIP(ctx)
		if err != nil {
			t.Fatalf("GetLanIP failed: %v", err)
		}
		t.Logf("Found %d LAN devices", len(items))
	})

	t.Run("Monitor_LanIPv6", func(t *testing.T) {
		svc := service.NewMonitorService(client)
		items, err := svc.GetLanIPv6(ctx)
		if err != nil {
			t.Fatalf("GetLanIPv6 failed: %v", err)
		}
		t.Logf("Found %d IPv6 devices", len(items))
	})

	t.Run("System_Homepage", func(t *testing.T) {
		svc := service.NewSystemService(client)
		data, err := svc.GetHomepage(ctx)
		if err != nil {
			t.Fatalf("GetHomepage failed: %v", err)
		}
		t.Logf("Hostname: %s, Uptime: %d", data.Hostname, data.Uptime)
	})

	t.Run("Network_Wan", func(t *testing.T) {
		svc := service.NewNetworkService(client)
		items, err := svc.GetWan(ctx)
		if err != nil {
			t.Fatalf("GetWan failed: %v", err)
		}
		t.Logf("Found %d WAN interfaces", len(items))
		for _, item := range items {
			t.Logf("  - %s: internet=%d", item.Name, item.Internet)
		}
	})

	t.Run("Network_Lan", func(t *testing.T) {
		svc := service.NewNetworkService(client)
		items, err := svc.GetLan(ctx)
		if err != nil {
			t.Fatalf("GetLan failed: %v", err)
		}
		t.Logf("Found %d LAN interfaces", len(items))
		for _, item := range items {
			t.Logf("  - %s: %s", item.Name, item.IPMask)
		}
	})

	t.Run("Network_IPv6", func(t *testing.T) {
		svc := service.NewNetworkService(client)
		items, err := svc.GetIPv6(ctx)
		if err != nil {
			t.Fatalf("GetIPv6 failed: %v", err)
		}
		t.Logf("Found %d IPv6 configs", len(items))
	})

	t.Run("Network_DDNS", func(t *testing.T) {
		svc := service.NewNetworkService(client)
		items, err := svc.GetDDNS(ctx)
		if err != nil {
			t.Fatalf("GetDDNS failed: %v", err)
		}
		t.Logf("Found %d DDNS configs", len(items))
		for _, item := range items {
			t.Logf("  - %s: %s -> %s (%s)", item.TagName, item.Domain, item.IPAddress, item.Result)
		}
	})

	t.Run("Firewall_DNAT", func(t *testing.T) {
		svc := service.NewFirewallService(client)
		items, err := svc.GetDNAT(ctx)
		if err != nil {
			t.Fatalf("GetDNAT failed: %v", err)
		}
		t.Logf("Found %d DNAT rules", len(items))
	})

	t.Run("Firewall_ACL", func(t *testing.T) {
		svc := service.NewFirewallService(client)
		items, err := svc.GetACL(ctx)
		if err != nil {
			t.Fatalf("GetACL failed: %v", err)
		}
		t.Logf("Found %d ACL rules", len(items))
	})

	t.Run("Firewall_CustomISP", func(t *testing.T) {
		svc := service.NewFirewallService(client)
		items, err := svc.GetCustomISP(ctx)
		if err != nil {
			t.Fatalf("GetCustomISP failed: %v", err)
		}
		t.Logf("Found %d CustomISP items", len(items))
	})

	t.Run("Firewall_StreamDomain", func(t *testing.T) {
		svc := service.NewFirewallService(client)
		items, err := svc.GetStreamDomain(ctx)
		if err != nil {
			t.Fatalf("GetStreamDomain failed: %v", err)
		}
		t.Logf("Found %d StreamDomain items", len(items))
	})

	t.Run("System_Upgrade", func(t *testing.T) {
		svc := service.NewSystemService(client)
		info, err := svc.GetUpgradeInfo(ctx)
		if err != nil {
			t.Fatalf("GetUpgradeInfo failed: %v", err)
		}
		t.Logf("System: %s, New: %s", info.SystemVer, info.NewSystemVer)
	})

	t.Run("VPN_PPTP", func(t *testing.T) {
		svc := service.NewVPNService(client)
		items, err := svc.GetPPTPClients(ctx)
		if err != nil {
			t.Fatalf("GetPPTPClients failed: %v", err)
		}
		t.Logf("Found %d PPTP clients", len(items))
	})

	t.Run("VPN_L2TP", func(t *testing.T) {
		svc := service.NewVPNService(client)
		items, err := svc.GetL2TPClients(ctx)
		if err != nil {
			t.Fatalf("GetL2TPClients failed: %v", err)
		}
		t.Logf("Found %d L2TP clients", len(items))
	})
}
