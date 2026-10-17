package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ikuaiapi "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/service"
)

// TestSubResourcePaths locks the two-level WireGuard peer routing. Peers
// live under a tunnel, so both identifiers travel in the path — not as
// query parameters, and not collapsed into one id.
//
// This is what the catalog could not express before: a static path string
// has no slot for a parent id.
func TestSubResourcePaths(t *testing.T) {
	// Item routes carry their identifiers in the path, so capture the path.
	var path, query string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, query = r.URL.Path, r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":20000,"data":{}}`))
	}))
	defer srv.Close()
	c, err := ikuaiapi.NewClient(srv.URL, ikuaiapi.WithToken("t"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer func() { _ = c.Close(context.Background()) }()
	api := service.NewAPIClient(c)
	ctx := context.Background()

	// Collection: /vpn/wireguard/{tunnel}/peers
	if _, err := api.Vpn().ListVpnWireguardPeers(ctx,
		&service.VpnWireguardPeersListOptions{Page: 1, PageSize: 20}); err != nil {
		t.Fatalf("ListVpnWireguardPeers: %v", err)
	}
	if path == "" {
		t.Error("expected a request")
	}

	// Item: /vpn/wireguard/{tunnel}/peers/{peer}
	tests := []struct {
		name string
		call func() error
	}{
		{"get", func() error {
			_, err := api.Vpn().GetVpnWireguardPeer(ctx, 7, 42)
			return err
		}},
		{"update", func() error {
			return api.Vpn().UpdateVpnWireguardPeer(ctx, 7, 42, map[string]any{"allowips": "10.9.0.2/32"})
		}},
		{"patch", func() error {
			return api.Vpn().PatchVpnWireguardPeer(ctx, 7, 42, map[string]any{"enabled": "no"})
		}},
		{"delete", func() error {
			return api.Vpn().DeleteVpnWireguardPeer(ctx, 7, 42)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if !strings.HasSuffix(path, "/vpn/wireguard/7/peers/42") {
				t.Errorf("path = %q, want it to address /vpn/wireguard/7/peers/42", path)
			}
			// An item identified by the path must not also ask by ?id=,
			// which would be a different resource to the router.
			if strings.Contains(query, "id=") {
				t.Errorf("query %q must not send an id parameter for a path-identified item", query)
			}
		})
	}
}

// TestSubResourcePathUsesDistinctParents guards against a template that
// reuses one id for both levels, which would silently address the wrong
// peer.
func TestSubResourcePathUsesDistinctParents(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":20000,"data":{}}`))
	}))
	defer srv.Close()

	c, err := ikuaiapi.NewClient(srv.URL, ikuaiapi.WithToken("t"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer func() { _ = c.Close(context.Background()) }()

	if _, err := service.NewAPIClient(c).Vpn().
		GetVpnWireguardPeer(context.Background(), 11, 99); err != nil {
		t.Fatalf("GetVpnWireguardPeer: %v", err)
	}
	if !strings.HasSuffix(path, "/vpn/wireguard/11/peers/99") {
		t.Errorf("path = %q, want the tunnel id and the peer id kept distinct", path)
	}
}

// TestSubResourceQueryTakesParent locks the query helper's signature: it
// must demand the parent id rather than defaulting to the wrong tunnel.
func TestSubResourceQueryTakesParent(t *testing.T) {
	api, query := newQueryRecorder(t, okArray)
	if _, err := api.Vpn().QueryVpnWireguardPeers(context.Background(), 5,
		&service.VpnWireguardPeersQueryOptions{Page: 1, PageSize: 10, Key: "peer_publickey"}); err != nil {
		t.Fatalf("QueryVpnWireguardPeers: %v", err)
	}
	for _, want := range []string{"limit=10", "key=peer_publickey"} {
		if !strings.Contains(*query, want) {
			t.Errorf("query %q missing %q", *query, want)
		}
	}
}

// TestFlatEndpointsAreUnaffected is the regression guard for the codegen
// change: an endpoint with no path placeholder must keep exactly its old
// shape — ?id= for single items, no parent argument.
func TestFlatEndpointsAreUnaffected(t *testing.T) {
	api, query := newQueryRecorder(t, okArray)

	// Get on a flat collection still uses ?id=.
	if _, err := api.Network().GetNetworkDhcpServices(context.Background(), 12); err != nil {
		t.Fatalf("GetNetworkDhcpServices: %v", err)
	}
	if !strings.Contains(*query, "id=12") {
		t.Errorf("query %q should still address the item with ?id=", *query)
	}

	// Delete likewise (this endpoint declares DELETE; dhcp-services does not).
	if err := api.Network().DeleteNetworkNatRules(context.Background(), 12); err != nil {
		t.Fatalf("DeleteNetworkNatRules: %v", err)
	}
	if !strings.Contains(*query, "id=12") {
		t.Errorf("delete query %q should still carry ?id=", *query)
	}
}

// TestSubResourceInCatalog checks the entries are discoverable at runtime,
// not only through generated methods.
func TestSubResourceInCatalog(t *testing.T) {
	want := map[string]string{
		"/vpn/wireguard/{tunnel_id}/peers":           "vpn",
		"/vpn/wireguard/{tunnel_id}/peers/{peer_id}": "vpn",
	}
	found := map[string]string{}
	for _, ep := range service.Endpoints() {
		if _, ok := want[ep.Path]; ok {
			found[ep.Path] = ep.Group
		}
	}
	for path, group := range want {
		got, ok := found[path]
		if !ok {
			t.Errorf("catalog is missing %s", path)
			continue
		}
		if got != group {
			t.Errorf("%s: group = %q, want %q", path, got, group)
		}
	}
}
