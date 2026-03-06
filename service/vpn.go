package service

import (
	"context"
	"fmt"

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

func (s *vpnService) AddPPTPClient(ctx context.Context, req *types.PPTPClientAddRequest) (int, error) {
	var result struct {
		types.BaseResponse
		ID int `json:"id"`
	}
	if err := s.client.Call(ctx, "pptp_client", "add", req, &result); err != nil {
		return 0, err
	}
	if !result.IsSuccess() {
		return 0, fmt.Errorf("failed to add PPTP client: %s", result.GetErrorMessage())
	}
	return result.ID, nil
}

func (s *vpnService) EditPPTPClient(ctx context.Context, req *types.PPTPClientEditRequest) error {
	var result types.BaseResponse
	if err := s.client.Call(ctx, "pptp_client", "edit", req, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to edit PPTP client: %s", result.GetErrorMessage())
	}
	return nil
}

func (s *vpnService) DelPPTPClient(ctx context.Context, id int) error {
	var result types.BaseResponse
	if err := s.client.Call(ctx, "pptp_client", "del", map[string]int{"id": id}, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to delete PPTP client: %s", result.GetErrorMessage())
	}
	return nil
}

func (s *vpnService) AddL2TPClient(ctx context.Context, req *types.L2TPClientAddRequest) (int, error) {
	var result struct {
		types.BaseResponse
		ID int `json:"id"`
	}
	if err := s.client.Call(ctx, "l2tp_client", "add", req, &result); err != nil {
		return 0, err
	}
	if !result.IsSuccess() {
		return 0, fmt.Errorf("failed to add L2TP client: %s", result.GetErrorMessage())
	}
	return result.ID, nil
}

func (s *vpnService) EditL2TPClient(ctx context.Context, req *types.L2TPClientEditRequest) error {
	var result types.BaseResponse
	if err := s.client.Call(ctx, "l2tp_client", "edit", req, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to edit L2TP client: %s", result.GetErrorMessage())
	}
	return nil
}

func (s *vpnService) DelL2TPClient(ctx context.Context, id int) error {
	var result types.BaseResponse
	if err := s.client.Call(ctx, "l2tp_client", "del", map[string]int{"id": id}, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to delete L2TP client: %s", result.GetErrorMessage())
	}
	return nil
}
