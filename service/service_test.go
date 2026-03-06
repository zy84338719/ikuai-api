package service

import (
	"context"
	"testing"

	ikuaisdk "github.com/zy84338719/ikuai-api"
)

// mockClient is a mock implementation for testing service creation
type mockClient struct {
	loggedIn bool
}

func TestNewAPIClient(t *testing.T) {
	// Create a real client (but don't connect)
	client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password")

	api := NewAPIClient(client)

	// Verify all services are created
	if api.Monitor() == nil {
		t.Error("Monitor service should not be nil")
	}
	if api.System() == nil {
		t.Error("System service should not be nil")
	}
	if api.Network() == nil {
		t.Error("Network service should not be nil")
	}
	if api.Firewall() == nil {
		t.Error("Firewall service should not be nil")
	}
	if api.VPN() == nil {
		t.Error("VPN service should not be nil")
	}
	if api.Log() == nil {
		t.Error("Log service should not be nil")
	}
	if api.Docker() == nil {
		t.Error("Docker service should not be nil")
	}
	if api.VM() == nil {
		t.Error("VM service should not be nil")
	}
	if api.UPnP() == nil {
		t.Error("UPnP service should not be nil")
	}
	if api.Traffic() == nil {
		t.Error("Traffic service should not be nil")
	}
	if api.AppControl() == nil {
		t.Error("AppControl service should not be nil")
	}
	if api.UserManage() == nil {
		t.Error("UserManage service should not be nil")
	}
	if api.OnlineMonitor() == nil {
		t.Error("OnlineMonitor service should not be nil")
	}
}

func TestServiceInterfaces(t *testing.T) {
	// Verify that service implementations satisfy their interfaces
	var _ MonitorService = (*monitorService)(nil)
	var _ SystemService = (*systemService)(nil)
	var _ NetworkService = (*networkService)(nil)
	var _ FirewallService = (*firewallService)(nil)
	var _ VPNService = (*vpnService)(nil)
	var _ LogService = (*logService)(nil)
	var _ DockerService = (*dockerService)(nil)
	var _ VMService = (*vmService)(nil)
	var _ UPnPService = (*upnpService)(nil)
	var _ TrafficService = (*trafficService)(nil)
	var _ AppControlService = (*appControlService)(nil)
	var _ UserManageService = (*userManageService)(nil)
	var _ OnlineMonitorService = (*onlineMonitorService)(nil)
}

func TestNewMonitorService(t *testing.T) {
	client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password")
	svc := NewMonitorService(client)

	if svc == nil {
		t.Fatal("NewMonitorService returned nil")
	}

	// Test that methods exist and return error when not logged in
	ctx := context.Background()

	_, err := svc.GetLanIP(ctx)
	if err == nil {
		t.Error("GetLanIP should return error when not logged in")
	}
}

func TestNewSystemService(t *testing.T) {
	client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password")
	svc := NewSystemService(client)

	if svc == nil {
		t.Fatal("NewSystemService returned nil")
	}
}

func TestNewNetworkService(t *testing.T) {
	client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password")
	svc := NewNetworkService(client)

	if svc == nil {
		t.Fatal("NewNetworkService returned nil")
	}
}

func TestNewFirewallService(t *testing.T) {
	client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password")
	svc := NewFirewallService(client)

	if svc == nil {
		t.Fatal("NewFirewallService returned nil")
	}
}

func TestNewVPNService(t *testing.T) {
	client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password")
	svc := NewVPNService(client)

	if svc == nil {
		t.Fatal("NewVPNService returned nil")
	}
}

func TestNewLogService(t *testing.T) {
	client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password")
	svc := NewLogService(client)

	if svc == nil {
		t.Fatal("NewLogService returned nil")
	}
}

func TestNewDockerService(t *testing.T) {
	client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password")
	svc := NewDockerService(client)

	if svc == nil {
		t.Fatal("NewDockerService returned nil")
	}
}

func TestNewVMService(t *testing.T) {
	client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password")
	svc := NewVMService(client)

	if svc == nil {
		t.Fatal("NewVMService returned nil")
	}
}

func TestNewUPnPService(t *testing.T) {
	client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password")
	svc := NewUPnPService(client)

	if svc == nil {
		t.Fatal("NewUPnPService returned nil")
	}
}

func TestNewTrafficService(t *testing.T) {
	client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password")
	svc := NewTrafficService(client)

	if svc == nil {
		t.Fatal("NewTrafficService returned nil")
	}
}

func TestNewAppControlService(t *testing.T) {
	client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password")
	svc := NewAppControlService(client)

	if svc == nil {
		t.Fatal("NewAppControlService returned nil")
	}
}

func TestNewUserManageService(t *testing.T) {
	client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password")
	svc := NewUserManageService(client)

	if svc == nil {
		t.Fatal("NewUserManageService returned nil")
	}
}

func TestNewOnlineMonitorService(t *testing.T) {
	client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password")
	svc := NewOnlineMonitorService(client)

	if svc == nil {
		t.Fatal("NewOnlineMonitorService returned nil")
	}
}
