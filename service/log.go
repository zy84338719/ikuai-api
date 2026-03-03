package service

import (
	"context"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type logService struct {
	client *ikuaisdk.Client
}

func NewLogService(client *ikuaisdk.Client) LogService {
	return &logService{client: client}
}

func (s *logService) GetNotice(ctx context.Context) ([]types.SyslogItem, error) {
	var resp types.SyslogNoticeShowResponse
	if err := s.client.Call(ctx, "syslog-notice", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *logService) GetWanPPPoE(ctx context.Context) ([]types.SyslogWanPPPoEItem, error) {
	var resp types.SyslogWanPPPoEShowResponse
	if err := s.client.Call(ctx, "syslog-wanpppoe", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *logService) GetDHCPD(ctx context.Context) ([]types.SyslogItem, error) {
	var resp types.SyslogDHCPDShowResponse
	if err := s.client.Call(ctx, "syslog-dhcpd", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *logService) GetARP(ctx context.Context) ([]types.SyslogItem, error) {
	var resp types.SyslogARPShowResponse
	if err := s.client.Call(ctx, "syslog-arp", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *logService) GetDDNS(ctx context.Context) ([]types.SyslogItem, error) {
	var resp types.SyslogDDNSShowResponse
	if err := s.client.Call(ctx, "syslog-ddns", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *logService) GetWebAdmin(ctx context.Context) ([]types.SyslogItem, error) {
	var resp types.SyslogWebAdminShowResponse
	if err := s.client.Call(ctx, "syslog-webadmin", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *logService) GetSysEvent(ctx context.Context) ([]types.SyslogItem, error) {
	var resp types.SyslogSysEventShowResponse
	if err := s.client.Call(ctx, "syslog-sysevent", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
