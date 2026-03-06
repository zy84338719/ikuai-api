package service

import (
	"context"
	"fmt"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type userManageService struct {
	client *ikuaisdk.Client
}

func NewUserManageService(client *ikuaisdk.Client) UserManageService {
	return &userManageService{client: client}
}

func (s *userManageService) GetUsers(ctx context.Context) ([]types.UserManageItem, error) {
	var resp types.UserManageShowResponse
	if err := s.client.Call(ctx, "user_manage", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *userManageService) AddUser(ctx context.Context, req *types.UserManageAddRequest) (int, error) {
	var result struct {
		types.BaseResponse
		ID int `json:"id"`
	}
	if err := s.client.Call(ctx, "user_manage", "add", req, &result); err != nil {
		return 0, err
	}
	if !result.IsSuccess() {
		return 0, fmt.Errorf("failed to add user: %s", result.GetErrorMessage())
	}
	return result.ID, nil
}

func (s *userManageService) EditUser(ctx context.Context, req *types.UserManageEditRequest) error {
	var result types.BaseResponse
	if err := s.client.Call(ctx, "user_manage", "edit", req, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to edit user: %s", result.GetErrorMessage())
	}
	return nil
}

func (s *userManageService) DelUser(ctx context.Context, id int) error {
	var result types.BaseResponse
	if err := s.client.Call(ctx, "user_manage", "del", map[string]int{"id": id}, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to delete user: %s", result.GetErrorMessage())
	}
	return nil
}

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
