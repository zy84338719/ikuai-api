//go:build integration
// +build integration

package ikuaisdk_test

import (
	"context"
	"os"
	"testing"
	"time"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/service"
)

var (
	testAddr     = os.Getenv("IKUAI_TEST_ADDR")
	testUsername = os.Getenv("IKUAI_TEST_USERNAME")
	testPassword = os.Getenv("IKUAI_TEST_PASSWORD")
)

func getTestConfig() (string, string, string) {
	addr := testAddr
	if addr == "" {
		addr = "10.10.40.254"
	}
	username := testUsername
	if username == "" {
		username = "zhangyi"
	}
	password := testPassword
	if password == "" {
		password = "zx19950124"
	}
	return addr, username, password
}

func TestIntegration_Login(t *testing.T) {
	addr, username, password := getTestConfig()

	client := ikuaisdk.NewClient(addr, username, password,
		ikuaisdk.WithTimeout(30*time.Second),
		ikuaisdk.WithInsecureSkipVerify(true),
	)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Login(ctx); err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if !client.IsLoggedIn() {
		t.Error("Client should be logged in after successful login")
	}

	if client.GetVersion() == ikuaisdk.VersionUnknown {
		t.Error("Version should be detected after login")
	}

	t.Logf("Successfully logged in, version: %s", client.GetVersion())
}

func TestIntegration_NewClientWithLogin(t *testing.T) {
	addr, username, password := getTestConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := ikuaisdk.NewClientWithLoginContext(ctx, addr, username, password,
		ikuaisdk.WithTimeout(30*time.Second),
		ikuaisdk.WithInsecureSkipVerify(true),
	)
	if err != nil {
		t.Fatalf("NewClientWithLogin failed: %v", err)
	}
	defer client.Close()

	if !client.IsLoggedIn() {
		t.Error("Client should be logged in")
	}
}

func TestIntegration_SystemService(t *testing.T) {
	addr, username, password := getTestConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := ikuaisdk.NewClientWithLoginContext(ctx, addr, username, password,
		ikuaisdk.WithTimeout(30*time.Second),
		ikuaisdk.WithInsecureSkipVerify(true),
	)
	if err != nil {
		t.Fatalf("NewClientWithLogin failed: %v", err)
	}
	defer client.Close()

	api := service.NewAPIClient(client)

	// Test GetHomepage
	homepage, err := api.System().GetHomepage(ctx)
	if err != nil {
		t.Fatalf("GetHomepage failed: %v", err)
	}

	t.Logf("Hostname: %s", homepage.Hostname)
	t.Logf("Version: %s", homepage.VerInfo.Version)
	t.Logf("Uptime: %d seconds", homepage.Uptime)

	if homepage.Hostname == "" {
		t.Error("Hostname should not be empty")
	}

	// Test GetWebUsers
	users, err := api.System().GetWebUsers(ctx)
	if err != nil {
		t.Logf("GetWebUsers failed (may not have permission): %v", err)
	} else {
		t.Logf("Found %d web users", len(users))
	}
}

func TestIntegration_MonitorService(t *testing.T) {
	addr, username, password := getTestConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := ikuaisdk.NewClientWithLoginContext(ctx, addr, username, password,
		ikuaisdk.WithTimeout(30*time.Second),
		ikuaisdk.WithInsecureSkipVerify(true),
	)
	if err != nil {
		t.Fatalf("NewClientWithLogin failed: %v", err)
	}
	defer client.Close()

	api := service.NewAPIClient(client)

	// Test GetLanIP
	devices, err := api.Monitor().GetLanIP(ctx)
	if err != nil {
		t.Fatalf("GetLanIP failed: %v", err)
	}
	t.Logf("Found %d LAN devices", len(devices))

	// Test GetInterfaces
	ifaces, err := api.Monitor().GetInterfaces(ctx)
	if err != nil {
		t.Fatalf("GetInterfaces failed: %v", err)
	}
	t.Logf("Found %d interface check items", len(ifaces.GetIFaceCheck()))
	t.Logf("Found %d interface stream items", len(ifaces.GetIFaceStream()))

	// Test GetSystem
	sys, err := api.Monitor().GetSystem(ctx)
	if err != nil {
		t.Fatalf("GetSystem failed: %v", err)
	}
	if len(sys) > 0 {
		t.Logf("CPU usage: %.2f%%", sys[0].CPU)
		t.Logf("Memory use: %d%%", sys[0].MemoryUse)
	}

	// Test GetARP
	arp, err := api.Monitor().GetARP(ctx)
	if err != nil {
		t.Logf("GetARP failed: %v", err)
	} else {
		t.Logf("Found %d ARP entries", len(arp))
	}
}

func TestIntegration_NetworkService(t *testing.T) {
	addr, username, password := getTestConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := ikuaisdk.NewClientWithLoginContext(ctx, addr, username, password,
		ikuaisdk.WithTimeout(30*time.Second),
		ikuaisdk.WithInsecureSkipVerify(true),
	)
	if err != nil {
		t.Fatalf("NewClientWithLogin failed: %v", err)
	}
	defer client.Close()

	api := service.NewAPIClient(client)

	// Test GetWan
	wans, err := api.Network().GetWan(ctx)
	if err != nil {
		t.Logf("GetWan failed: %v", err)
	} else {
		t.Logf("Found %d WAN interfaces", len(wans))
		for _, wan := range wans {
			t.Logf("  WAN %s: internet=%d", wan.Name, wan.Internet)
		}
	}

	// Test GetLan
	lans, err := api.Network().GetLan(ctx)
	if err != nil {
		t.Logf("GetLan failed: %v", err)
	} else {
		t.Logf("Found %d LAN interfaces", len(lans))
	}

	// Test GetDHCPD
	dhcpd, err := api.Network().GetDHCPD(ctx)
	if err != nil {
		t.Logf("GetDHCPD failed: %v", err)
	} else {
		t.Logf("Found %d DHCP servers", len(dhcpd))
	}

	// Test GetLeases
	leases, err := api.Network().GetLeases(ctx)
	if err != nil {
		t.Logf("GetLeases failed: %v", err)
	} else {
		t.Logf("Found %d DHCP leases", len(leases))
	}
}

func TestIntegration_FirewallService(t *testing.T) {
	addr, username, password := getTestConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := ikuaisdk.NewClientWithLoginContext(ctx, addr, username, password,
		ikuaisdk.WithTimeout(30*time.Second),
		ikuaisdk.WithInsecureSkipVerify(true),
	)
	if err != nil {
		t.Fatalf("NewClientWithLogin failed: %v", err)
	}
	defer client.Close()

	api := service.NewAPIClient(client)

	// Test GetDNAT (port forwarding)
	dnats, err := api.Firewall().GetDNAT(ctx)
	if err != nil {
		t.Logf("GetDNAT failed: %v", err)
	} else {
		t.Logf("Found %d DNAT rules", len(dnats))
	}

	// Test GetACL
	acls, err := api.Firewall().GetACL(ctx)
	if err != nil {
		t.Logf("GetACL failed: %v", err)
	} else {
		t.Logf("Found %d ACL rules", len(acls))
	}
}

func TestIntegration_LogService(t *testing.T) {
	addr, username, password := getTestConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := ikuaisdk.NewClientWithLoginContext(ctx, addr, username, password,
		ikuaisdk.WithTimeout(30*time.Second),
		ikuaisdk.WithInsecureSkipVerify(true),
	)
	if err != nil {
		t.Fatalf("NewClientWithLogin failed: %v", err)
	}
	defer client.Close()

	api := service.NewAPIClient(client)

	// Test GetNotice
	notice, err := api.Log().GetNotice(ctx)
	if err != nil {
		t.Logf("GetNotice failed: %v", err)
	} else {
		t.Logf("Found %d notice log entries", len(notice))
	}

	// Test GetSysEvent
	events, err := api.Log().GetSysEvent(ctx)
	if err != nil {
		t.Logf("GetSysEvent failed: %v", err)
	} else {
		t.Logf("Found %d system event log entries", len(events))
	}
}

func TestIntegration_TrafficService(t *testing.T) {
	addr, username, password := getTestConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := ikuaisdk.NewClientWithLoginContext(ctx, addr, username, password,
		ikuaisdk.WithTimeout(30*time.Second),
		ikuaisdk.WithInsecureSkipVerify(true),
	)
	if err != nil {
		t.Fatalf("NewClientWithLogin failed: %v", err)
	}
	defer client.Close()

	api := service.NewAPIClient(client)

	// Test GetRealtime
	realtime, err := api.Traffic().GetRealtime(ctx)
	if err != nil {
		t.Logf("GetRealtime failed: %v", err)
	} else {
		t.Logf("Found %d realtime traffic items", len(realtime))
	}

	// Test GetHistory
	history, err := api.Traffic().GetHistory(ctx, 24)
	if err != nil {
		t.Logf("GetHistory failed: %v", err)
	} else {
		t.Logf("Found %d traffic history items", len(history))
	}
}
