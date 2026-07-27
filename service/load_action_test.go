package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	ikuaiapi "github.com/zy84338719/ikuai-api"
)

// TestLoadOptionsValidation covers the typed Load<Name> method
// signature and the four validation paths on <Name>LoadOptions.
func TestLoadOptionsValidation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{}}`))
	}))
	defer srv.Close()

	c, _ := ikuaiapi.NewClient(srv.URL, ikuaiapi.WithToken("x"))
	api := NewAPIClient(c)
	ctx := context.Background()

	cases := []struct {
		name string
		opts *MonitoringCpuLoadOptions
		want string
	}{
		{"nil options", nil, "options are required"},
		{"bad datetype", &MonitoringCpuLoadOptions{DataType: "year", StartTime: 1, EndTime: 2, Math: "avg"}, "DataType must be one of"},
		{"bad math", &MonitoringCpuLoadOptions{DataType: "hour", StartTime: 1, EndTime: 2, Math: "sum"}, "Math must be one of"},
		{"zero start", &MonitoringCpuLoadOptions{DataType: "hour", StartTime: 0, EndTime: 100, Math: "avg"}, "are required"},
		{"start >= end", &MonitoringCpuLoadOptions{DataType: "hour", StartTime: 100, EndTime: 100, Math: "avg"}, "must be less than"},
		{"valid hour/avg", &MonitoringCpuLoadOptions{DataType: "hour", StartTime: 100, EndTime: 200, Math: "avg"}, ""},
		{"valid day/max", &MonitoringCpuLoadOptions{DataType: "day", StartTime: 1, EndTime: 2, Math: "max"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := api.Monitoring().LoadCpu(ctx, tc.opts)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.want)
			}
			var apiErr *ikuaiapi.APIError
			if !errorsAsAPIError(err, &apiErr) {
				t.Errorf("expected *APIError, got %T", err)
			}
		})
	}
}

// TestLoadEndpointSendsCorrectQuery verifies the query string that
// the Load<Name> method sends to the router.
func TestLoadEndpointSendsCorrectQuery(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"code":0,"data":{}}`))
	}))
	defer srv.Close()

	c, _ := ikuaiapi.NewClient(srv.URL, ikuaiapi.WithToken("x"))
	api := NewAPIClient(c)
	_, err := api.Monitoring().LoadMemory(context.Background(),
		&MonitoringMemoryLoadOptions{DataType: "week", StartTime: 1773300000, EndTime: 1773303600, Math: "max"})
	if err != nil {
		t.Fatalf("LoadMemory: %v", err)
	}
	q, _ := url.ParseQuery(gotQuery)
	for k, v := range map[string]string{
		"datetype":   "week",
		"start_time": "1773300000",
		"end_time":   "1773303600",
		"math":       "max",
	} {
		if q.Get(k) != v {
			t.Errorf("query %s = %q, want %q", k, q.Get(k), v)
		}
	}
}

// TestActionHelpersRoute verifies the Start/Stop/Restart/Sync/Restore/Check
// helpers post to the right paths and (for body-taking actions) forward
// the body JSON.
func TestActionHelpersRoute(t *testing.T) {
	type callSpec struct {
		name        string
		call        func(api *APIClient) error
		wantPath    string
		wantBodySet bool
	}
	cases := []callSpec{
		{
			name: "StartNetworkAcServices",
			call: func(api *APIClient) error {
				_, err := api.Network().StartNetworkAcServices(context.Background(), nil)
				return err
			},
			wantPath: "/network/ac/services:start",
		},
		{
			name: "StopNetworkAcServices",
			call: func(api *APIClient) error {
				_, err := api.Network().StopNetworkAcServices(context.Background())
				return err
			},
			wantPath: "/network/ac/services:stop",
		},
		{
			name: "RestartNetworkDhcpServices",
			call: func(api *APIClient) error {
				_, err := api.Network().RestartNetworkDhcpServices(context.Background())
				return err
			},
			wantPath: "/network/dhcp/services:restart",
		},
		{
			name: "SyncSystemBasicNtp",
			call: func(api *APIClient) error {
				_, err := api.System().SyncSystemBasicNtp(context.Background(), nil)
				return err
			},
			wantPath: "/system/basic/ntp:sync",
		},
		{
			name: "RestoreSystemBackup",
			call: func(api *APIClient) error {
				_, err := api.System().RestoreSystemBackup(context.Background(),
					map[string]any{"srcfile": "backup-2026-07-27.tar"})
				return err
			},
			wantPath:    "/system/backup:restore",
			wantBodySet: true,
		},
		{
			name: "CheckSystemUpgrade",
			call: func(api *APIClient) error {
				_, err := api.System().CheckSystemUpgrade(context.Background(), nil)
				return err
			},
			wantPath: "/system/upgrade:check",
		},
		{
			name: "StartSystemUpgrade",
			call: func(api *APIClient) error {
				_, err := api.System().StartSystemUpgrade(context.Background(),
					map[string]any{"srcfile": "firmware.tar"})
				return err
			},
			wantPath:    "/system/upgrade:start",
			wantBodySet: true,
		},
		{
			name: "StartSystemVrrp",
			call: func(api *APIClient) error {
				_, err := api.System().StartSystemVrrp(context.Background(), nil)
				return err
			},
			wantPath: "/system/vrrp:start",
		},
		{
			name: "StopSystemVrrp",
			call: func(api *APIClient) error {
				_, err := api.System().StopSystemVrrp(context.Background())
				return err
			},
			wantPath: "/system/vrrp:stop",
		},
		{
			name: "DeleteBackupViaCall",
			call: func(api *APIClient) error {
				_, err := api.Call(context.Background(), "system", "backup", "DELETE", nil,
					map[string]string{"srcfile": "backup-2026-07-27.tar"})
				return err
			},
			wantPath: "/system/backup",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotPath, gotQuery, gotBody string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotQuery = r.URL.RawQuery
				buf := make([]byte, 0, 1024)
				tmp := make([]byte, 512)
				for {
					n, err := r.Body.Read(tmp)
					buf = append(buf, tmp[:n]...)
					if err != nil {
						break
					}
				}
				gotBody = string(buf)
				_, _ = w.Write([]byte(`{"code":0,"data":{}}`))
			}))
			defer srv.Close()

			c, err := ikuaiapi.NewClient(srv.URL, ikuaiapi.WithToken("x"))
			if err != nil {
				t.Fatal(err)
			}
			api := NewAPIClient(c)

			if err := tc.call(api); err != nil {
				t.Fatalf("call: %v", err)
			}
			if gotPath != "/api/v4.0"+tc.wantPath {
				t.Errorf("path = %q, want %q", gotPath, "/api/v4.0"+tc.wantPath)
			}
			if tc.name == "DeleteBackupViaCall" {
				if gotQuery != "srcfile=backup-2026-07-27.tar" {
					t.Errorf("query = %q, want srcfile=backup-2026-07-27.tar", gotQuery)
				}
			}
			if tc.wantBodySet && gotBody == "" {
				t.Errorf("body was empty; want non-empty JSON")
			}
		})
	}
}

func errorsAsAPIError(err error, target **ikuaiapi.APIError) bool {
	for {
		if err == nil {
			return false
		}
		if e, ok := err.(*ikuaiapi.APIError); ok {
			*target = e
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
}
