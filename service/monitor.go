package service

import (
	"context"

	ikuaiapi "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type monitorService struct {
	client *ikuaiapi.Client
}

func NewMonitorService(client *ikuaiapi.Client) MonitorService {
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
	var checkResp types.MonitorIFaceShowResponse
	if err := s.client.Call(ctx, "monitor_iface", "show", map[string]string{"TYPE": "iface_check"}, &checkResp); err != nil {
		return nil, err
	}
	if !checkResp.IsSuccess() {
		return nil, ikuaiapi.NewSDKError(ikuaiapi.ErrCodeRequestFailed, checkResp.GetErrorMessage(), nil)
	}

	var streamResp types.MonitorIFaceShowResponse
	if err := s.client.Call(ctx, "monitor_iface", "show", map[string]string{"TYPE": "iface_stream"}, &streamResp); err != nil {
		return nil, err
	}
	if !streamResp.IsSuccess() {
		return nil, ikuaiapi.NewSDKError(ikuaiapi.ErrCodeRequestFailed, streamResp.GetErrorMessage(), nil)
	}

	var resp types.MonitorIFaceShowResponse
	resp.Result = checkResp.Result
	resp.ErrMsg = checkResp.ErrMsg
	resp.Data.IFaceCheck = checkResp.GetIFaceCheck()
	resp.Data.IFaceStream = streamResp.GetIFaceStream()
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
