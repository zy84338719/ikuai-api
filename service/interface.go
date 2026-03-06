package service

import (
	"context"

	"github.com/zy84338719/ikuai-api/types"
)

type MonitorService interface {
	GetLanIP(ctx context.Context) ([]types.MonitorLanIPItem, error)
	GetLanIPv6(ctx context.Context) ([]types.MonitorLanIPv6Item, error)
	GetInterfaces(ctx context.Context) (*types.MonitorIFaceShowResponse, error)
	GetSystem(ctx context.Context) ([]types.MonitorSystemItem, error)
	GetARP(ctx context.Context) ([]types.ARPItem, error)
}

type SystemService interface {
	GetHomepage(ctx context.Context) (*types.HomepageSysStat, error)
	GetUpgradeInfo(ctx context.Context) (*types.UpgradeInfo, error)
	GetBackupList(ctx context.Context) ([]types.BackupItem, error)
	GetWebUsers(ctx context.Context) ([]types.WebUserItem, error)
}

type NetworkService interface {
	GetWan(ctx context.Context) ([]types.WanItem, error)
	GetLan(ctx context.Context) ([]types.LanItem, error)
	GetVLAN(ctx context.Context) ([]types.VLANItem, error)
	GetIPv6(ctx context.Context) ([]types.IPv6Item, error)
	GetIPTV(ctx context.Context) ([]types.IPTVItem, error)
	GetDDNS(ctx context.Context) ([]types.DDNSItem, error)
	GetDHCPD(ctx context.Context) ([]types.DHCPDItem, error)
	GetStaticBindings(ctx context.Context) ([]types.DHCPStaticItem, error)
	GetLeases(ctx context.Context) ([]types.DHCPLeaseItem, error)
	// DNS related
	GetDNSForward(ctx context.Context) ([]types.DNSForwardItem, error)
	GetDNSStatic(ctx context.Context) ([]types.DNSStaticItem, error)
	// Routing related
	GetRouteStatic(ctx context.Context) ([]types.RouteStaticItem, error)
	GetRoutePolicy(ctx context.Context) ([]types.RoutePolicyItem, error)
	// Flow control related
	GetFlowControl(ctx context.Context) ([]types.FlowControlItem, error)
	// Bandwidth control
	GetBandwidth(ctx context.Context) ([]types.BandwidthItem, error)
	// QoS
	GetQoS(ctx context.Context) ([]types.QoSItem, error)
}

type FirewallService interface {
	GetACL(ctx context.Context) ([]types.ACLItem, error)
	GetDNAT(ctx context.Context) ([]types.DNATItem, error)
	GetConnLimit(ctx context.Context) ([]types.ConnLimitItem, error)
	GetDomainGroups(ctx context.Context) ([]types.DomainGroupItem, error)
	GetCustomISP(ctx context.Context) ([]types.CustomISPItem, error)
	GetStreamDomain(ctx context.Context) ([]types.StreamDomainItem, error)
}

type VPNService interface {
	GetPPTPClients(ctx context.Context) ([]types.PPTPClientItem, error)
	GetL2TPClients(ctx context.Context) ([]types.L2TPClientItem, error)
}

type LogService interface {
	GetNotice(ctx context.Context) ([]types.SyslogItem, error)
	GetWanPPPoE(ctx context.Context) ([]types.SyslogWanPPPoEItem, error)
	GetDHCPD(ctx context.Context) ([]types.SyslogItem, error)
	GetARP(ctx context.Context) ([]types.SyslogItem, error)
	GetDDNS(ctx context.Context) ([]types.SyslogItem, error)
	GetWebAdmin(ctx context.Context) ([]types.SyslogItem, error)
	GetSysEvent(ctx context.Context) ([]types.SyslogItem, error)
}

type DockerService interface {
	GetImages(ctx context.Context) ([]types.DockerImageItem, error)
	GetContainers(ctx context.Context) ([]types.DockerContainerItem, error)
	GetNetworks(ctx context.Context) ([]types.DockerNetworkItem, error)
	GetComposes(ctx context.Context) ([]types.DockerComposeItem, error)
}

type VMService interface {
	List(ctx context.Context) ([]types.QemuVM, error)
	Add(ctx context.Context, req *types.QemuAddRequest) (int, error)
	Edit(ctx context.Context, req *types.QemuEditRequest) error
	Del(ctx context.Context, id int) error
	Start(ctx context.Context, id int) error
	Stop(ctx context.Context, id int) error
	Restart(ctx context.Context, id int) error
}

type UPnPService interface {
	List(ctx context.Context) ([]types.UPnPItem, error)
	Add(ctx context.Context, req *types.UPnPAddRequest) (int, error)
	Edit(ctx context.Context, req *types.UPnPEditRequest) error
	Del(ctx context.Context, id int) error
}

type APIClient interface {
	Monitor() MonitorService
	System() SystemService
	Network() NetworkService
	Firewall() FirewallService
	VPN() VPNService
	Log() LogService
	Docker() DockerService
	VM() VMService
	UPnP() UPnPService
}
