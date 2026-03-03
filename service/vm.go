package service

import (
	"context"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type vmService struct {
	client *ikuaisdk.Client
}

func NewVMService(client *ikuaisdk.Client) VMService {
	return &vmService{client: client}
}

func (s *vmService) List(ctx context.Context) ([]types.QemuVM, error) {
	var resp types.QemuShowResponse
	if err := s.client.Call(ctx, "qemu", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *vmService) Add(ctx context.Context, req *types.QemuAddRequest) (int, error) {
	var result struct {
		types.BaseResponse
		ID int `json:"id"`
	}
	if err := s.client.Call(ctx, "qemu", "add", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

func (s *vmService) Edit(ctx context.Context, req *types.QemuEditRequest) error {
	var result types.BaseResponse
	return s.client.Call(ctx, "qemu", "edit", req, &result)
}

func (s *vmService) Del(ctx context.Context, id int) error {
	req := &types.QemuDelRequest{ID: id}
	var result types.BaseResponse
	return s.client.Call(ctx, "qemu", "del", req, &result)
}

func (s *vmService) Start(ctx context.Context, id int) error {
	req := &types.QemuStartRequest{ID: id}
	var result types.BaseResponse
	return s.client.Call(ctx, "qemu", "start", req, &result)
}

func (s *vmService) Stop(ctx context.Context, id int) error {
	req := &types.QemuStopRequest{ID: id}
	var result types.BaseResponse
	return s.client.Call(ctx, "qemu", "stop", req, &result)
}

func (s *vmService) Restart(ctx context.Context, id int) error {
	req := &types.QemuRestartRequest{ID: id}
	var result types.BaseResponse
	return s.client.Call(ctx, "qemu", "restart", req, &result)
}
