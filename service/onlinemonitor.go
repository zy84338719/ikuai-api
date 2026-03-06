package service

import (
	"context"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type onlineMonitorService struct {
	client *ikuaisdk.Client
}

func NewOnlineMonitorService(client *ikuaisdk.Client) OnlineMonitorService {
	return &onlineMonitorService{client: client}
}

func (s *onlineMonitorService) GetOnlineUsers(ctx context.Context) ([]types.OnlineMonitorItem, error) {
	var resp types.OnlineMonitorShowResponse
	if err := s.client.Call(ctx, "online_monitor", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
