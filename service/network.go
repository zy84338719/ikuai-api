package service

import (
	"context"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type networkService struct {
	client *ikuaisdk.Client
}

func NewNetworkService(client *ikuaisdk.Client) NetworkService {
	return &networkService{client: client}
}

func (s *networkService) GetWan(ctx context.Context) ([]types.WanItem, error) {
	var resp types.WanShowResponse
	if err := s.client.Call(ctx, "wan", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *networkService) GetLan(ctx context.Context) ([]types.LanItem, error) {
	var resp types.LanShowResponse
	if err := s.client.Call(ctx, "lan", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *networkService) GetVLAN(ctx context.Context) ([]types.VLANItem, error) {
	var resp types.VLANShowResponse
	if err := s.client.Call(ctx, "vlan", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *networkService) GetIPv6(ctx context.Context) ([]types.IPv6Item, error) {
	var resp types.IPv6ShowResponse
	if err := s.client.Call(ctx, "ipv6", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *networkService) GetIPTV(ctx context.Context) ([]types.IPTVItem, error) {
	var resp types.IPTVShowResponse
	if err := s.client.Call(ctx, "iptv", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *networkService) GetDDNS(ctx context.Context) ([]types.DDNSItem, error) {
	var resp types.DDNSShowResponse
	if err := s.client.Call(ctx, "ddns", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *networkService) GetDHCPD(ctx context.Context) ([]types.DHCPDItem, error) {
	var resp types.DHCPDShowResponse
	if err := s.client.Call(ctx, "dhcpd", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *networkService) GetStaticBindings(ctx context.Context) ([]types.DHCPStaticItem, error) {
	var resp types.DHCPStaticShowResponse
	if err := s.client.Call(ctx, "dhcp_static", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *networkService) GetLeases(ctx context.Context) ([]types.DHCPLeaseItem, error) {
	var resp types.DHCPLeaseShowResponse
	if err := s.client.Call(ctx, "dhcp_lease", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
// 新增方法实现
func (s *networkService) GetDNSForward(ctx context.Context) ([]types.DNSForwardItem, error) {
	var resp types.DNSForwardShowResponse
	if err := s.client.Call(ctx, "dns_forward", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *networkService) GetDNSStatic(ctx context.Context) ([]types.DNSStaticItem, error) {
	var resp types.DNSStaticShowResponse
	if err := s.client.Call(ctx, "dns_static", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
func (s *networkService) GetRouteStatic(ctx context.Context) ([]types.RouteStaticItem, error) {
	var resp types.RouteStaticShowResponse
	if err := s.client.Call(ctx, "route_static", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
func (s *networkService) GetRoutePolicy(ctx context.Context) ([]types.RoutePolicyItem, error) {
	var resp types.RoutePolicyShowResponse
	if err := s.client.Call(ctx, "route_policy", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
func (s *networkService) GetFlowControl(ctx context.Context) ([]types.FlowControlItem, error) {
	var resp types.FlowControlShowResponse
	if err := s.client.Call(ctx, "flow_control", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
func (s *networkService) GetBandwidth(ctx context.Context) ([]types.BandwidthItem, error) {
	var resp types.BandwidthShowResponse
	if err := s.client.Call(ctx, "bandwidth", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
func (s *networkService) GetQoS(ctx context.Context) ([]types.QoSItem, error) {
	var resp types.QoSShowResponse
	if err := s.client.Call(ctx, "qos", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
