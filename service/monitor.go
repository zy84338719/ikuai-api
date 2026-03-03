package service

import (
	"context"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type monitorService struct {
	client *ikuaisdk.Client
}

func NewMonitorService(client *ikuaisdk.Client) MonitorService {
	return &monitorService{client: client}
}

func (s *monitorService) GetLanIP(ctx context.Context) ([]types.MonitorLanIPItem, error) {
	var resp types.MonitorLanIPShowResponse
	if err := s.client.Call(ctx, "monitor_lanip", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *monitorService) GetLanIPv6(ctx context.Context) ([]types.MonitorLanIPv6Item, error) {
	var resp types.MonitorLanIPv6ShowResponse
	if err := s.client.Call(ctx, "monitor_lanipv6", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *monitorService) GetInterfaces(ctx context.Context) (*types.MonitorIFaceShowResponse, error) {
	var resp types.MonitorIFaceShowResponse
	if err := s.client.Call(ctx, "monitor_iface", "show", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *monitorService) GetSystem(ctx context.Context) ([]types.MonitorSystemItem, error) {
	var resp types.MonitorSystemShowResponse
	if err := s.client.Call(ctx, "monitor_system", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *monitorService) GetARP(ctx context.Context) ([]types.ARPItem, error) {
	var resp types.ARPShowResponse
	if err := s.client.Call(ctx, "arp", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
