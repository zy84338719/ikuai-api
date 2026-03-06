package service

import (
	ikuaisdk "github.com/zy84338719/ikuai-api"
)

type apiClient struct {
	monitor       MonitorService
	system        SystemService
	network       NetworkService
	firewall      FirewallService
	vpn           VPNService
	log           LogService
	docker        DockerService
	vm            VMService
	upnp          UPnPService
	traffic       TrafficService
	appControl    AppControlService
	userManage    UserManageService
	onlineMonitor OnlineMonitorService
}

func NewAPIClient(client *ikuaisdk.Client) APIClient {
	return &apiClient{
		monitor:       NewMonitorService(client),
		system:        NewSystemService(client),
		network:       NewNetworkService(client),
		firewall:      NewFirewallService(client),
		vpn:           NewVPNService(client),
		log:           NewLogService(client),
		docker:        NewDockerService(client),
		vm:            NewVMService(client),
		upnp:          NewUPnPService(client),
		traffic:       NewTrafficService(client),
		appControl:    NewAppControlService(client),
		userManage:    NewUserManageService(client),
		onlineMonitor: NewOnlineMonitorService(client),
	}
}

func (c *apiClient) Monitor() MonitorService {
	return c.monitor
}

func (c *apiClient) System() SystemService {
	return c.system
}

func (c *apiClient) Network() NetworkService {
	return c.network
}

func (c *apiClient) Firewall() FirewallService {
	return c.firewall
}

func (c *apiClient) VPN() VPNService {
	return c.vpn
}

func (c *apiClient) Log() LogService {
	return c.log
}

func (c *apiClient) Docker() DockerService {
	return c.docker
}

func (c *apiClient) VM() VMService {
	return c.vm
}

func (c *apiClient) UPnP() UPnPService {
	return c.upnp
}

func (c *apiClient) Traffic() TrafficService {
	return c.traffic
}
func (c *apiClient) AppControl() AppControlService {
	return c.appControl
}
func (c *apiClient) UserManage() UserManageService {
	return c.userManage
}
func (c *apiClient) OnlineMonitor() OnlineMonitorService {
	return c.onlineMonitor
}
