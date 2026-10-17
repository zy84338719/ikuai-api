package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ikuaiapi "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/service"
)

// newQueryRecorder returns an API client wired to a test server plus the
// captured query string of the last request.
func newQueryRecorder(t *testing.T, handler http.HandlerFunc) (*service.APIClient, *string) {
	t.Helper()
	got := new(string)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*got = r.URL.RawQuery
		handler(w, r)
	}))
	t.Cleanup(srv.Close)

	c, err := ikuaiapi.NewClient(srv.URL, ikuaiapi.WithToken("t"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = c.Close(context.Background()) })
	return service.NewAPIClient(c), got
}

func okArray(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"code":20000,"data":[]}`))
}

// TestLegacyDialectQuery locks the uppercase TYPE/LIMIT/OFFSET contract that
// /monitoring/downstream, /monitoring/cameras and /monitoring/switch speak.
// Sending the default page/limit paging here makes the router ignore the
// whole query, so this asserts the exact legacy key set.
func TestLegacyDialectQuery(t *testing.T) {
	api, query := newQueryRecorder(t, okArray)

	_, err := api.Monitoring().QueryMonitoringDownstream(context.Background(),
		&service.MonitoringDownstreamQueryOptions{
			Page:     3,
			PageSize: 25,
			Order:    "asc",
			OrderBy:  "id",
			Status:   "1",
			Device:   "camera",
		})
	if err != nil {
		t.Fatalf("QueryMonitoringDownstream: %v", err)
	}

	for _, want := range []string{
		"TYPE=data",
		"LIMIT=50%2C25", // (3-1)*25 = 50 offset, 25 per page
		"OFFSET=50",
		"ORDER=asc",
		"ORDER_BY=id",
		"status=1",
		"device=camera",
	} {
		if !strings.Contains(*query, want) {
			t.Errorf("query %q missing %q", *query, want)
		}
	}
	// The legacy dialect has no lowercase paging at all.
	for _, unwanted := range []string{"page=", "limit=", "order=", "order_by="} {
		if strings.Contains(*query, unwanted) {
			t.Errorf("query %q must not contain lowercase key %q", *query, unwanted)
		}
	}
}

// TestLegacyDialectValidation checks the guards that stop a nonsense request
// before it reaches the router.
func TestLegacyDialectValidation(t *testing.T) {
	api, query := newQueryRecorder(t, okArray)
	ctx := context.Background()

	if _, err := api.Monitoring().QueryMonitoringDownstream(ctx, nil); err == nil {
		t.Error("expected an error for nil options")
	}
	if _, err := api.Monitoring().QueryMonitoringDownstream(ctx,
		&service.MonitoringDownstreamQueryOptions{Page: 0, PageSize: 10}); err == nil {
		t.Error("expected an error for Page < 1")
	}
	if _, err := api.Monitoring().QueryMonitoringDownstream(ctx,
		&service.MonitoringDownstreamQueryOptions{Page: 1, PageSize: 0}); err == nil {
		t.Error("expected an error for PageSize < 1")
	}
	if *query != "" {
		t.Errorf("no request should have been sent, got query %q", *query)
	}
}

// TestMixedDialectQuery covers /auth/online-users: lowercase paging with
// uppercase search and sort keys.
func TestMixedDialectQuery(t *testing.T) {
	api, query := newQueryRecorder(t, okArray)

	_, err := api.Auth().QueryAuthOnlineUsers(context.Background(),
		&service.AuthOnlineUsersQueryOptions{
			Page:     2,
			PageSize: 10,
			Keywords: "office",
			Finds:    "username,name",
			Order:    "desc",
			OrderBy:  "start_time",
		})
	if err != nil {
		t.Fatalf("QueryAuthOnlineUsers: %v", err)
	}
	for _, want := range []string{
		"page=2", "limit=10",
		"KEYWORDS=office", "FINDS=username%2Cname",
		"ORDER=desc", "ORDER_BY=start_time",
	} {
		if !strings.Contains(*query, want) {
			t.Errorf("query %q missing %q", *query, want)
		}
	}
}

// TestWindowDialectQuery covers the starttime/stoptime endpoints and their
// required selectors.
func TestWindowDialectQuery(t *testing.T) {
	api, query := newQueryRecorder(t, okArray)

	_, err := api.Monitoring().QueryMonitoringClientsProtocols(context.Background(),
		&service.MonitoringClientsProtocolsQueryOptions{
			IP:        "192.168.1.100",
			MAC:       "08:9b:4b:01:7e:7c",
			StartTime: 1773300000,
			StopTime:  1773303600,
		})
	if err != nil {
		t.Fatalf("QueryMonitoringClientsProtocols: %v", err)
	}
	for _, want := range []string{
		"ip=192.168.1.100",
		"mac=08%3A9b%3A4b%3A01%3A7e%3A7c",
		"starttime=1773300000",
		"stoptime=1773303600",
	} {
		if !strings.Contains(*query, want) {
			t.Errorf("query %q missing %q", *query, want)
		}
	}

	// ip and mac are required: the router returns nothing useful without them.
	api2, query2 := newQueryRecorder(t, okArray)
	if _, err := api2.Monitoring().QueryMonitoringClientsProtocols(context.Background(),
		&service.MonitoringClientsProtocolsQueryOptions{MAC: "aa:bb:cc:dd:ee:ff"}); err == nil {
		t.Error("expected an error when the required IP is missing")
	}
	if *query2 != "" {
		t.Errorf("no request should have been sent, got query %q", *query2)
	}
}

// TestNoDialectQuery covers an unpaginated endpoint with an optional
// selector, and confirms no paging keys leak in.
func TestNoDialectQuery(t *testing.T) {
	api, query := newQueryRecorder(t, okArray)

	_, err := api.Monitoring().QueryMonitoringChannelClients(context.Background(),
		&service.MonitoringChannelClientsQueryOptions{Channel: 36})
	if err != nil {
		t.Fatalf("QueryMonitoringChannelClients: %v", err)
	}
	if !strings.Contains(*query, "channel=36") {
		t.Errorf("query %q missing channel=36", *query)
	}
	for _, unwanted := range []string{"page=", "limit=", "TYPE=", "OFFSET="} {
		if strings.Contains(*query, unwanted) {
			t.Errorf("query %q must not contain %q", *query, unwanted)
		}
	}
}

// TestPagedDialectWithRequiredFields covers /monitoring/clients/app-protocols/load,
// which pages like the standard dialect but also needs ip and mac.
func TestPagedDialectWithRequiredFields(t *testing.T) {
	api, query := newQueryRecorder(t, okArray)

	_, err := api.Monitoring().QueryMonitoringClientsAppProtocolsLoad(context.Background(),
		&service.MonitoringClientsAppProtocolsLoadQueryOptions{
			Page:     1,
			PageSize: 20,
			IP:       "192.168.1.50",
			MAC:      "aa:bb:cc:dd:ee:ff",
		})
	if err != nil {
		t.Fatalf("QueryMonitoringClientsAppProtocolsLoad: %v", err)
	}
	for _, want := range []string{"page=1", "limit=20", "ip=192.168.1.50"} {
		if !strings.Contains(*query, want) {
			t.Errorf("query %q missing %q", *query, want)
		}
	}

	// The optional numeric "limit" field is covered by PageSize, so it must
	// not be emitted twice.
	if strings.Count(*query, "limit=") != 1 {
		t.Errorf("query %q should carry limit exactly once", *query)
	}
}

// TestURLVisitsEndpoint covers the endpoint that was missing from the
// catalog entirely before this change.
func TestURLVisitsEndpoint(t *testing.T) {
	api, query := newQueryRecorder(t, okArray)

	_, err := api.Log().QueryLogUrlVisits(context.Background(),
		&service.LogUrlVisitsQueryOptions{
			Page:      1,
			PageSize:  20,
			Pattern:   "example.com",
			StartTime: 1761842000,
			StopTime:  1761843000,
		})
	if err != nil {
		t.Fatalf("QueryLogUrlVisits: %v", err)
	}
	for _, want := range []string{
		"page=1", "limit=20", "pattern=example.com",
		"starttime=1761842000", "stoptime=1761843000",
	} {
		if !strings.Contains(*query, want) {
			t.Errorf("query %q missing %q", *query, want)
		}
	}
}

// TestLoadEndpointQueryParams checks the load-style window still speaks
// datetype/start_time/end_time/math (with underscores, unlike the
// starttime/stoptime endpoints above).
func TestLoadEndpointQueryParams(t *testing.T) {
	api, query := newQueryRecorder(t, okArray)

	_, err := api.Monitoring().LoadCpu(context.Background(),
		&service.MonitoringCpuLoadOptions{
			DataType:  "hour",
			StartTime: 1773300000,
			EndTime:   1773303600,
			Math:      "avg",
		})
	if err != nil {
		t.Fatalf("LoadMonitoringCpu: %v", err)
	}
	for _, want := range []string{"datetype=hour", "start_time=1773300000", "end_time=1773303600", "math=avg"} {
		if !strings.Contains(*query, want) {
			t.Errorf("query %q missing %q", *query, want)
		}
	}
}

// TestQueryDialectRoundTrip verifies every catalog entry parses and that the
// dialect is one the renderer knows about.
func TestQueryDialectRoundTrip(t *testing.T) {
	seen := 0
	for _, ep := range ikuaiapi.V4EndpointCatalog {
		dialect, required, optional, err := ep.QuerySpec()
		if err != nil {
			t.Errorf("%s: %v", ep.Path, err)
			continue
		}
		seen++
		switch dialect {
		case ikuaiapi.DialectPaged, ikuaiapi.DialectLegacy, ikuaiapi.DialectMixed,
			ikuaiapi.DialectNone, ikuaiapi.DialectLoad, ikuaiapi.DialectWindow,
			ikuaiapi.DialectRef:
		default:
			t.Errorf("%s: unknown dialect %q", ep.Path, dialect)
		}
		for _, f := range append(required, optional...) {
			if f.Name == "" {
				t.Errorf("%s: field with empty name", ep.Path)
			}
		}
	}
	if seen == 0 {
		t.Fatal("catalog is empty")
	}
}

// TestDryRunPreviewsRequest keeps the new query helpers compatible with the
// SDK's dry-run mode: nothing hits the network, the preview is returned.
func TestDryRunPreviewsRequest(t *testing.T) {
	c, err := ikuaiapi.NewClient("https://192.0.2.1", ikuaiapi.WithToken("t"), ikuaiapi.WithDryRun(true))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	raw, err := service.NewAPIClient(c).Monitoring().QueryMonitoringDownstream(context.Background(),
		&service.MonitoringDownstreamQueryOptions{Page: 2, PageSize: 10})
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}
	var preview map[string]any
	if err := json.Unmarshal(raw, &preview); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if preview["dry_run"] != true {
		t.Errorf("expected a dry_run preview, got %s", raw)
	}
	if got := preview["path"]; got != "/monitoring/downstream" {
		t.Errorf("preview path = %v, want /monitoring/downstream", got)
	}
	q, _ := preview["query"].(string)
	for _, want := range []string{"TYPE=data", "OFFSET=10", "LIMIT=10%2C10"} {
		if !strings.Contains(q, want) {
			t.Errorf("preview query %q missing %q", q, want)
		}
	}
}
