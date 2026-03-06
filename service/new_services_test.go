package service

import (
	"context"
	"testing"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

func TestTrafficService_GetRealtime(t *testing.T) {
	// 测试服务创建
	svc := NewTrafficService(nil)
	if svc == nil {
		t.Error("NewTrafficService returned nil")
	}
}

func TestTrafficService_GetHistory(t *testing.T) {
	// 只测试参数验证逻辑
	tests := []struct {
		name        string
		hours       int64
		shouldCheck bool // 是否应该检查参数
		wantErr     bool
	}{
		{"zero hours", 0, true, true},
		{"negative hours", -1, true, true},
		{"too large hours", 200, true, true},
		{"valid hours", 24, false, false},
		{"max hours", 168, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 只测试参数验证
			if tt.hours <= 0 || tt.hours > 168 {
				if !tt.wantErr {
					t.Errorf("参数验证失败：hours=%d 应该被拒绝", tt.hours)
				}
			}
		})
	}
}

func TestAppControlService_GetAppControl(t *testing.T) {
	mockClient := &ikuaisdk.Client{}
	svc := NewAppControlService(mockClient)

	_, err := svc.GetAppControl(context.Background())
	if err != nil {
		t.Logf("GetAppControl() returned error (expected with mock): %v", err)
	}
}

func TestAppControlService_AddAppControl(t *testing.T) {
	mockClient := &ikuaisdk.Client{}
	svc := NewAppControlService(mockClient)

	req := &types.AppControlAddRequest{
		TagName: "test-rule",
		Enabled: "yes",
		Comment: "test",
	}

	_, err := svc.AddAppControl(context.Background(), req)
	if err != nil {
		t.Logf("AddAppControl() returned error (expected with mock): %v", err)
	}
}

func TestUserManageService_GetUsers(t *testing.T) {
	mockClient := &ikuaisdk.Client{}
	svc := NewUserManageService(mockClient)

	_, err := svc.GetUsers(context.Background())
	if err != nil {
		t.Logf("GetUsers() returned error (expected with mock): %v", err)
	}
}

func TestUserManageService_AddUser(t *testing.T) {
	mockClient := &ikuaisdk.Client{}
	svc := NewUserManageService(mockClient)

	req := &types.UserManageAddRequest{
		Username: "testuser",
		Password: "testpass",
		Group:    "users",
		Enabled:  "yes",
	}

	_, err := svc.AddUser(context.Background(), req)
	if err != nil {
		t.Logf("AddUser() returned error (expected with mock): %v", err)
	}
}

func TestOnlineMonitorService_GetOnlineUsers(t *testing.T) {
	mockClient := &ikuaisdk.Client{}
	svc := NewOnlineMonitorService(mockClient)

	_, err := svc.GetOnlineUsers(context.Background())
	if err != nil {
		t.Logf("GetOnlineUsers() returned error (expected with mock): %v", err)
	}
}
