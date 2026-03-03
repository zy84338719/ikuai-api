package service

import (
	"context"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type vpnService struct {
	client *ikuaisdk.Client
}

func NewVPNService(client *ikuaisdk.Client) VPNService {
	return &vpnService{client: client}
}

func (s *vpnService) GetPPTPClients(ctx context.Context) ([]types.PPTPClientItem, error) {
	var resp types.PPTPClientShowResponse
	if err := s.client.Call(ctx, "pptp_client", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *vpnService) GetL2TPClients(ctx context.Context) ([]types.L2TPClientItem, error) {
	var resp types.L2TPClientShowResponse
	if err := s.client.Call(ctx, "l2tp_client", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
