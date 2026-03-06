package service

import (
	"context"
	"fmt"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type appControlService struct {
	client *ikuaisdk.Client
}

func NewAppControlService(client *ikuaisdk.Client) AppControlService {
	return &appControlService{client: client}
}

func (s *appControlService) GetAppControl(ctx context.Context) ([]types.AppControlItem, error) {
	var resp types.AppControlShowResponse
	if err := s.client.Call(ctx, "app_control", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *appControlService) AddAppControl(ctx context.Context, req *types.AppControlAddRequest) (int, error) {
	var result struct {
		types.BaseResponse
		ID int `json:"id"`
	}
	if err := s.client.Call(ctx, "app_control", "add", req, &result); err != nil {
		return 0, err
	}
	if !result.IsSuccess() {
		return 0, fmt.Errorf("failed to add app control: %s", result.GetErrorMessage())
	}
	return result.ID, nil
}

func (s *appControlService) EditAppControl(ctx context.Context, req *types.AppControlEditRequest) error {
	var result types.BaseResponse
	if err := s.client.Call(ctx, "app_control", "edit", req, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to edit app control: %s", result.GetErrorMessage())
	}
	return nil
}

func (s *appControlService) DelAppControl(ctx context.Context, id int) error {
	var result types.BaseResponse
	if err := s.client.Call(ctx, "app_control", "del", map[string]int{"id": id}, &result); err != nil {
		return err
	}
	if !result.IsSuccess() {
		return fmt.Errorf("failed to delete app control: %s", result.GetErrorMessage())
	}
	return nil
}
