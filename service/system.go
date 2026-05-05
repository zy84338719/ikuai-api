package service

import (
	"context"

	ikuaiapi "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type systemService struct {
	client *ikuaiapi.Client
}

func NewSystemService(client *ikuaiapi.Client) SystemService {
	return &systemService{client: client}
}

func (s *systemService) GetHomepage(ctx context.Context) (*types.HomepageSysStat, error) {
	var resp types.HomepageShowResponse
	if err := s.client.Call(ctx, "homepage", "show", map[string]string{"TYPE": "sysstat"}, &resp); err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, ikuaiapi.NewSDKError(ikuaiapi.ErrCodeRequestFailed, resp.GetErrorMessage(), nil)
	}
	return resp.GetData(), nil
}

func (s *systemService) GetUpgradeInfo(ctx context.Context) (*types.UpgradeInfo, error) {
	var resp types.UpgradeShowResponse
	if err := s.client.Call(ctx, "upgrade", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *systemService) GetBackupList(ctx context.Context) ([]types.BackupItem, error) {
	var resp types.BackupShowResponse
	if err := s.client.Call(ctx, "backup", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *systemService) GetWebUsers(ctx context.Context) ([]types.WebUserItem, error) {
	var resp types.WebUserShowResponse
	if err := s.client.Call(ctx, "webuser", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
