package service

import (
	"context"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type upnpService struct {
	client *ikuaisdk.Client
}

func NewUPnPService(client *ikuaisdk.Client) UPnPService {
	return &upnpService{client: client}
}

func (s *upnpService) List(ctx context.Context) ([]types.UPnPItem, error) {
	var resp types.UPnPShowResponse
	if err := s.client.Call(ctx, "upnp", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *upnpService) Add(ctx context.Context, req *types.UPnPAddRequest) (int, error) {
	var result struct {
		types.BaseResponse
		ID int `json:"id"`
	}
	if err := s.client.Call(ctx, "upnp", "add", req, &result); err != nil {
		return 0, err
	}
	return result.ID, nil
}

func (s *upnpService) Edit(ctx context.Context, req *types.UPnPEditRequest) error {
	var result types.BaseResponse
	return s.client.Call(ctx, "upnp", "edit", req, &result)
}

func (s *upnpService) Del(ctx context.Context, id int) error {
	req := &types.UPnPDelRequest{ID: id}
	var result types.BaseResponse
	return s.client.Call(ctx, "upnp", "del", req, &result)
}
