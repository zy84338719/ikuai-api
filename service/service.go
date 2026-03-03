package service

import (
	"context"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type MonitorService struct {
	client *ikuaisdk.Client
}

func NewMonitorService(client *ikuaisdk.Client) *MonitorService {
	return &MonitorService{client: client}
}

func (s *MonitorService) GetLanIP(ctx context.Context) ([]types.MonitorLanIPItem, error) {
	var resp types.MonitorLanIPShowResponse
	if err := s.client.Call(ctx, "monitor_lanip", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *MonitorService) GetLanIPv6(ctx context.Context) ([]types.MonitorLanIPv6Item, error) {
	var resp types.MonitorLanIPv6ShowResponse
	if err := s.client.Call(ctx, "monitor_lanipv6", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *MonitorService) GetInterfaces(ctx context.Context) (*types.MonitorIFaceShowResponse, error) {
	var resp types.MonitorIFaceShowResponse
	if err := s.client.Call(ctx, "monitor_iface", "show", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *MonitorService) GetSystem(ctx context.Context) ([]types.MonitorSystemItem, error) {
	var resp types.MonitorSystemShowResponse
	if err := s.client.Call(ctx, "monitor_system", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *MonitorService) GetARP(ctx context.Context) ([]types.ARPItem, error) {
	var resp types.ARPShowResponse
	if err := s.client.Call(ctx, "arp", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

type SystemService struct {
	client *ikuaisdk.Client
}

func NewSystemService(client *ikuaisdk.Client) *SystemService {
	return &SystemService{client: client}
}

func (s *SystemService) GetHomepage(ctx context.Context) (*types.HomepageSysStat, error) {
	var resp types.HomepageShowResponse
	if err := s.client.Call(ctx, "homepage", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *SystemService) GetUpgradeInfo(ctx context.Context) (*types.UpgradeInfo, error) {
	var resp types.UpgradeShowResponse
	if err := s.client.Call(ctx, "upgrade", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *SystemService) GetBackupList(ctx context.Context) ([]types.BackupItem, error) {
	var resp types.BackupShowResponse
	if err := s.client.Call(ctx, "backup", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *SystemService) GetWebUsers(ctx context.Context) ([]types.WebUserItem, error) {
	var resp types.WebUserShowResponse
	if err := s.client.Call(ctx, "webuser", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

type NetworkService struct {
	client *ikuaisdk.Client
}

func NewNetworkService(client *ikuaisdk.Client) *NetworkService {
	return &NetworkService{client: client}
}

func (s *NetworkService) GetWan(ctx context.Context) ([]types.WanItem, error) {
	var resp types.WanShowResponse
	if err := s.client.Call(ctx, "wan", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *NetworkService) GetLan(ctx context.Context) ([]types.LanItem, error) {
	var resp types.LanShowResponse
	if err := s.client.Call(ctx, "lan", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *NetworkService) GetVLAN(ctx context.Context) ([]types.VLANItem, error) {
	var resp types.VLANShowResponse
	if err := s.client.Call(ctx, "vlan", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *NetworkService) GetIPv6(ctx context.Context) ([]types.IPv6Item, error) {
	var resp types.IPv6ShowResponse
	if err := s.client.Call(ctx, "ipv6", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *NetworkService) GetIPTV(ctx context.Context) ([]types.IPTVItem, error) {
	var resp types.IPTVShowResponse
	if err := s.client.Call(ctx, "iptv", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *NetworkService) GetDDNS(ctx context.Context) ([]types.DDNSItem, error) {
	var resp types.DDNSShowResponse
	if err := s.client.Call(ctx, "ddns", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *NetworkService) GetDHCPD(ctx context.Context) ([]types.DHCPDItem, error) {
	var resp types.DHCPDShowResponse
	if err := s.client.Call(ctx, "dhcpd", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *NetworkService) GetStaticBindings(ctx context.Context) ([]types.DHCPStaticItem, error) {
	var resp types.DHCPStaticShowResponse
	if err := s.client.Call(ctx, "dhcp_static", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *NetworkService) GetLeases(ctx context.Context) ([]types.DHCPLeaseItem, error) {
	var resp types.DHCPLeaseShowResponse
	if err := s.client.Call(ctx, "dhcp_lease", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

type FirewallService struct {
	client *ikuaisdk.Client
}

func NewFirewallService(client *ikuaisdk.Client) *FirewallService {
	return &FirewallService{client: client}
}

func (s *FirewallService) GetACL(ctx context.Context) ([]types.ACLItem, error) {
	var resp types.ACLShowResponse
	if err := s.client.Call(ctx, "acl", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *FirewallService) GetDNAT(ctx context.Context) ([]types.DNATItem, error) {
	var resp types.DNATShowResponse
	if err := s.client.Call(ctx, "dnat", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *FirewallService) GetConnLimit(ctx context.Context) ([]types.ConnLimitItem, error) {
	var resp types.ConnLimitShowResponse
	if err := s.client.Call(ctx, "conn_limit", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *FirewallService) GetDomainGroups(ctx context.Context) ([]types.DomainGroupItem, error) {
	var resp types.DomainGroupShowResponse
	if err := s.client.Call(ctx, "domain_group", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *FirewallService) GetCustomISP(ctx context.Context) ([]types.CustomISPItem, error) {
	var resp types.CustomISPShowResponse
	if err := s.client.Call(ctx, "custom_isp", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *FirewallService) GetStreamDomain(ctx context.Context) ([]types.StreamDomainItem, error) {
	var resp types.StreamDomainShowResponse
	if err := s.client.Call(ctx, "stream_domain", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

type VPNService struct {
	client *ikuaisdk.Client
}

func NewVPNService(client *ikuaisdk.Client) *VPNService {
	return &VPNService{client: client}
}

func (s *VPNService) GetPPTPClients(ctx context.Context) ([]types.PPTPClientItem, error) {
	var resp types.PPTPClientShowResponse
	if err := s.client.Call(ctx, "pptp_client", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *VPNService) GetL2TPClients(ctx context.Context) ([]types.L2TPClientItem, error) {
	var resp types.L2TPClientShowResponse
	if err := s.client.Call(ctx, "l2tp_client", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

type LogService struct {
	client *ikuaisdk.Client
}

func NewLogService(client *ikuaisdk.Client) *LogService {
	return &LogService{client: client}
}

func (s *LogService) GetNotice(ctx context.Context) ([]types.SyslogItem, error) {
	var resp types.SyslogNoticeShowResponse
	if err := s.client.Call(ctx, "syslog-notice", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *LogService) GetWanPPPoE(ctx context.Context) ([]types.SyslogWanPPPoEItem, error) {
	var resp types.SyslogWanPPPoEShowResponse
	if err := s.client.Call(ctx, "syslog-wanpppoe", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *LogService) GetDHCPD(ctx context.Context) ([]types.SyslogItem, error) {
	var resp types.SyslogDHCPDShowResponse
	if err := s.client.Call(ctx, "syslog-dhcpd", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *LogService) GetARP(ctx context.Context) ([]types.SyslogItem, error) {
	var resp types.SyslogARPShowResponse
	if err := s.client.Call(ctx, "syslog-arp", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *LogService) GetDDNS(ctx context.Context) ([]types.SyslogItem, error) {
	var resp types.SyslogDDNSShowResponse
	if err := s.client.Call(ctx, "syslog-ddns", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *LogService) GetWebAdmin(ctx context.Context) ([]types.SyslogItem, error) {
	var resp types.SyslogWebAdminShowResponse
	if err := s.client.Call(ctx, "syslog-webadmin", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *LogService) GetSysEvent(ctx context.Context) ([]types.SyslogItem, error) {
	var resp types.SyslogSysEventShowResponse
	if err := s.client.Call(ctx, "syslog-sysevent", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

type DockerService struct {
	client *ikuaisdk.Client
}

func NewDockerService(client *ikuaisdk.Client) *DockerService {
	return &DockerService{client: client}
}

func (s *DockerService) GetImages(ctx context.Context) ([]types.DockerImageItem, error) {
	var resp types.DockerImageShowResponse
	if err := s.client.Call(ctx, "docker_image", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *DockerService) GetContainers(ctx context.Context) ([]types.DockerContainerItem, error) {
	var resp types.DockerContainerShowResponse
	if err := s.client.Call(ctx, "docker_container", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *DockerService) GetNetworks(ctx context.Context) ([]types.DockerNetworkItem, error) {
	var resp types.DockerNetworkShowResponse
	if err := s.client.Call(ctx, "docker_network", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *DockerService) GetComposes(ctx context.Context) ([]types.DockerComposeItem, error) {
	var resp types.DockerComposeShowResponse
	if err := s.client.Call(ctx, "docker_compose", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

type VMService struct {
	client *ikuaisdk.Client
}

func NewVMService(client *ikuaisdk.Client) *VMService {
	return &VMService{client: client}
}

func (s *VMService) List(ctx context.Context) ([]types.QemuVM, error) {
	var resp types.QemuShowResponse
	if err := s.client.Call(ctx, "qemu", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *VMService) Add(ctx context.Context, req *types.QemuAddRequest) (int, error) {
	var result struct {
		types.BaseResponse
		ID int `json:"id"`
	}
	if err := s.client.Call(ctx, "qemu", "add", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

func (s *VMService) Edit(ctx context.Context, req *types.QemuEditRequest) error {
	var result types.BaseResponse
	return s.client.Call(ctx, "qemu", "edit", req, &result)
}

func (s *VMService) Del(ctx context.Context, id int) error {
	req := &types.QemuDelRequest{ID: id}
	var result types.BaseResponse
	return s.client.Call(ctx, "qemu", "del", req, &result)
}

func (s *VMService) Start(ctx context.Context, id int) error {
	req := &types.QemuStartRequest{ID: id}
	var result types.BaseResponse
	return s.client.Call(ctx, "qemu", "start", req, &result)
}

func (s *VMService) Stop(ctx context.Context, id int) error {
	req := &types.QemuStopRequest{ID: id}
	var result types.BaseResponse
	return s.client.Call(ctx, "qemu", "stop", req, &result)
}

func (s *VMService) Restart(ctx context.Context, id int) error {
	req := &types.QemuRestartRequest{ID: id}
	var result types.BaseResponse
	return s.client.Call(ctx, "qemu", "restart", req, &result)
}

type UPnPService struct {
	client *ikuaisdk.Client
}

func NewUPnPService(client *ikuaisdk.Client) *UPnPService {
	return &UPnPService{client: client}
}

func (s *UPnPService) List(ctx context.Context) ([]types.UPnPItem, error) {
	var resp types.UPnPShowResponse
	if err := s.client.Call(ctx, "upnp", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *UPnPService) Add(ctx context.Context, req *types.UPnPAddRequest) (int, error) {
	var result struct {
		types.BaseResponse
		ID int `json:"id"`
	}
	if err := s.client.Call(ctx, "upnp", "add", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

func (s *UPnPService) Edit(ctx context.Context, req *types.UPnPEditRequest) error {
	var result types.BaseResponse
	return s.client.Call(ctx, "upnp", "edit", req, &result)
}

func (s *UPnPService) Del(ctx context.Context, id int) error {
	req := &types.UPnPD[elRequest{ID: id}
	var result types.BaseResponse
	return s.client.Call(ctx, "upnp", "del", req, &result)
}
