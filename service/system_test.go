package service

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSystemService_GetHomepage(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"results": map[string]interface{}{
				"sysstat": map[string]interface{}{
					"hostname": "iKuai-Router",
					"uptime":   86400,
					"verinfo": map[string]interface{}{
						"version":   "3.6.9",
						"modelname": "iKuai-X86",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	client, server := setupTestClient(handler)
	defer server.Close()

	svc := NewSystemService(client)
	ctx := context.Background()

	homepage, err := svc.GetHomepage(ctx)
	if err != nil {
		t.Fatalf("GetHomepage() error = %v", err)
	}

	if homepage.Hostname != "iKuai-Router" {
		t.Errorf("GetHomepage() Hostname = %s, want iKuai-Router", homepage.Hostname)
	}
}

func TestSystemService_GetUpgradeInfo(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"results": map[string]interface{}{
				"data": map[string]interface{}{
					"system_ver":     "3.6.9",
					"new_system_ver": "3.7.0",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	client, server := setupTestClient(handler)
	defer server.Close()

	svc := NewSystemService(client)
	ctx := context.Background()

	upgrade, err := svc.GetUpgradeInfo(ctx)
	if err != nil {
		t.Fatalf("GetUpgradeInfo() error = %v", err)
	}

	if upgrade.SystemVer != "3.6.9" {
		t.Errorf("GetUpgradeInfo() SystemVer = %s, want 3.6.9", upgrade.SystemVer)
	}
}

func TestSystemService_GetBackupList(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"results": map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":         1,
						"enabled":    "yes",
						"strategy":   "auto",
						"cycle_time": "daily",
						"tagname":    "daily-backup",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	client, server := setupTestClient(handler)
	defer server.Close()

	svc := NewSystemService(client)
	ctx := context.Background()

	backups, err := svc.GetBackupList(ctx)
	if err != nil {
		t.Fatalf("GetBackupList() error = %v", err)
	}

	if len(backups) != 1 {
		t.Fatalf("GetBackupList() returned %d items, want 1", len(backups))
	}

	if backups[0].TagName != "daily-backup" {
		t.Errorf("GetBackupList() TagName = %s, want daily-backup", backups[0].TagName)
	}
}

func TestSystemService_GetWebUsers(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"results": map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":         1,
						"username":   "admin",
						"group_name": "administrators",
						"enabled":    "yes",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})

	client, server := setupTestClient(handler)
	defer server.Close()

	svc := NewSystemService(client)
	ctx := context.Background()

	users, err := svc.GetWebUsers(ctx)
	if err != nil {
		t.Fatalf("GetWebUsers() error = %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("GetWebUsers() returned %d items, want 1", len(users))
	}

	if users[0].Username != "admin" {
		t.Errorf("GetWebUsers() Username = %s, want admin", users[0].Username)
	}
}
