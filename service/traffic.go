package service

import (
	"context"
	"fmt"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type TrafficService interface {
	GetRealtime(ctx context.Context) ([]types.TrafficRealtimeItem, error)
	GetHistory(ctx context.Context, hours int64) ([]types.TrafficHistoryItem, error)
}
type trafficService struct {
	client *ikuaisdk.Client
}

func NewTrafficService(client *ikuaisdk.Client) TrafficService {
	return &trafficService{client: client}
}
func (s *trafficService) GetRealtime(ctx context.Context) ([]types.TrafficRealtimeItem, error) {
	var resp types.TrafficRealtimeShowResponse
	if err := s.client.Call(ctx, "traffic_realtime", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
func (s *trafficService) GetHistory(ctx context.Context, hours int64) ([]types.TrafficHistoryItem, error) {
	if hours <= 0 || hours > 168 {
		return nil, fmt.Errorf("hours must be between 1 and 168")
	}
	req := &types.TrafficHistoryRequest{Hours: hours}
	var resp types.TrafficHistoryShowResponse
	if err := s.client.Call(ctx, "traffic_history", "show", req, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
