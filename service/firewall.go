package service

import (
	"context"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type firewallService struct {
	client *ikuaisdk.Client
}

func NewFirewallService(client *ikuaisdk.Client) FirewallService {
	return &firewallService{client: client}
}

func (s *firewallService) GetACL(ctx context.Context) ([]types.ACLItem, error) {
	var resp types.ACLShowResponse
	if err := s.client.Call(ctx, "acl", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *firewallService) GetDNAT(ctx context.Context) ([]types.DNATItem, error) {
	var resp types.DNATShowResponse
	if err := s.client.Call(ctx, "dnat", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *firewallService) GetConnLimit(ctx context.Context) ([]types.ConnLimitItem, error) {
	var resp types.ConnLimitShowResponse
	if err := s.client.Call(ctx, "conn_limit", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *firewallService) GetDomainGroups(ctx context.Context) ([]types.DomainGroupItem, error) {
	var resp types.DomainGroupShowResponse
	if err := s.client.Call(ctx, "domain_group", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *firewallService) GetCustomISP(ctx context.Context) ([]types.CustomISPItem, error) {
	var resp types.CustomISPShowResponse
	if err := s.client.Call(ctx, "custom_isp", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *firewallService) GetStreamDomain(ctx context.Context) ([]types.StreamDomainItem, error) {
	var resp types.StreamDomainShowResponse
	if err := s.client.Call(ctx, "stream_domain", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
