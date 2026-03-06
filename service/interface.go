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
	AddStaticBinding(ctx context.Context, req *types.DHCPStaticAddRequest) (int, error)
	EditStaticBinding(ctx context.Context, req *types.DHCPStaticEditRequest) error
	DelStaticBinding(ctx context.Context, id int) error
	GetLeases(ctx context.Context) ([]types.DHCPLeaseItem, error)

	GetDNSForward(ctx context.Context) ([]types.DNSForwardItem, error)
	GetDNSStatic(ctx context.Context) ([]types.DNSStaticItem, error)
	AddDNSStatic(ctx context.Context, req *types.DNSStaticAddRequest) (int, error)
	EditDNSStatic(ctx context.Context, req *types.DNSStaticEditRequest) error
	DelDNSStatic(ctx context.Context, id int) error

	GetRouteStatic(ctx context.Context) ([]types.RouteStaticItem, error)
	AddRouteStatic(ctx context.Context, req *types.RouteStaticAddRequest) (int, error)
	EditRouteStatic(ctx context.Context, req *types.RouteStaticEditRequest) error
	DelRouteStatic(ctx context.Context, id int) error

	GetRoutePolicy(ctx context.Context) ([]types.RoutePolicyItem, error)
	GetFlowControl(ctx context.Context) ([]types.FlowControlItem, error)
	GetBandwidth(ctx context.Context) ([]types.BandwidthItem, error)
	GetQoS(ctx context.Context) ([]types.QoSItem, error)
}

type FirewallService interface {
	GetACL(ctx context.Context) ([]types.ACLItem, error)
	AddACL(ctx context.Context, req *types.ACLAddRequest) (int, error)
	EditACL(ctx context.Context, req *types.ACLEditRequest) error
	DelACL(ctx context.Context, id int) error

	GetDNAT(ctx context.Context) ([]types.DNATItem, error)
	AddDNAT(ctx context.Context, req *types.DNATAddRequest) (int, error)
	EditDNAT(ctx context.Context, req *types.DNATEditRequest) error
	DelDNAT(ctx context.Context, id int) error

	GetConnLimit(ctx context.Context) ([]types.ConnLimitItem, error)
	AddConnLimit(ctx context.Context, req *types.ConnLimitAddRequest) (int, error)
	EditConnLimit(ctx context.Context, req *types.ConnLimitEditRequest) error
	DelConnLimit(ctx context.Context, id int) error

	GetDomainGroups(ctx context.Context) ([]types.DomainGroupItem, error)
	GetCustomISP(ctx context.Context) ([]types.CustomISPItem, error)
	GetStreamDomain(ctx context.Context) ([]types.StreamDomainItem, error)
}

type VPNService interface {
	GetPPTPClients(ctx context.Context) ([]types.PPTPClientItem, error)
	AddPPTPClient(ctx context.Context, req *types.PPTPClientAddRequest) (int, error)
	EditPPTPClient(ctx context.Context, req *types.PPTPClientEditRequest) error
	DelPPTPClient(ctx context.Context, id int) error

	GetL2TPClients(ctx context.Context) ([]types.L2TPClientItem, error)
	AddL2TPClient(ctx context.Context, req *types.L2TPClientAddRequest) (int, error)
	EditL2TPClient(ctx context.Context, req *types.L2TPClientEditRequest) error
	DelL2TPClient(ctx context.Context, id int) error
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
