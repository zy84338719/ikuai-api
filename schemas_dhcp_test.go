package ikuaiapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestDHCPCreateWritesNoOptionTypes guards the split between the two verbs.
// A create must not carry opt_type* fields at all; sending them would
// reconfigure DHCP options the operator never asked about.
func TestDHCPCreateWritesNoOptionTypes(t *testing.T) {
	body, err := CreateBody("/network/dhcp/services", map[string]any{
		"tagname":     "office",
		"interface":   "lan1",
		"phy_ifnames": "all",
		"addr_pool":   "192.168.1.100-200",
		"netmask":     "255.255.255.0",
		"gateway":     "192.168.1.1",
		"lease":       "120",
	})
	if err != nil {
		t.Fatalf("CreateBody: %v", err)
	}
	if body["enabled"] != "yes" {
		t.Errorf("enabled = %#v, want \"yes\"", body["enabled"])
	}
	for key := range body {
		if len(key) > 8 && key[:8] == "opt_type" {
			t.Errorf("a create must not carry %s", key)
		}
	}
	// lease looks numeric but stays a string: it is not on the SDK's
	// integer-field allowlist, and the firmware is verified to accept it
	// that way. Coercing it would be a change of contract, not a fix.
	if body["lease"] != "120" {
		t.Errorf("lease = %#v, want the string \"120\"", body["lease"])
	}
}

// TestDHCPUpdateFillsOptionTypes is the other half: an update must send all
// twelve option-type fields, because the firmware revalidates them and
// treats an absent one as unset rather than unchanged.
func TestDHCPUpdateFillsOptionTypes(t *testing.T) {
	var sent map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":20000,"data":[{
				"id":7,"tagname":"office","interface":"lan1",
				"opt_type121":"","opt15":"custom-value"}]}`))
			return
		}
		b := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(b)
		sent = decode(t, b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":20000,"message":"ok"}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, WithToken("t"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer func() { _ = c.Close(context.Background()) }()

	body, err := UpdatePayload(context.Background(), c, "/network/dhcp/services", 7,
		map[string]any{"lease": "240"}, nil)
	if err != nil {
		t.Fatalf("UpdatePayload: %v", err)
	}

	for _, key := range []string{
		"opt_type15", "opt_type28", "opt_type43", "opt_type60", "opt_type66",
		"opt_type67", "opt_type80", "opt_type119", "opt_type125",
		"opt_type128", "opt_type138", "opt_type121",
	} {
		if _, ok := body[key]; !ok {
			t.Errorf("update is missing %s", key)
		}
	}
	// opt_type121 is the one option that defaults to something other than 0.
	if got := body["opt_type121"]; got != 2 {
		t.Errorf("opt_type121 = %#v, want the default 2", got)
	}
	for _, key := range []string{"opt_type15", "opt_type28", "opt_type43"} {
		if got := body[key]; got != 0 {
			t.Errorf("%s = %#v, want the default 0", key, got)
		}
	}
	// An empty option value in the response must not survive as ""; the
	// default replaces it, or the router reads it as a blank option.
	if got, ok := body["opt15"]; ok && got == "" {
		t.Error("an empty opt value should have been replaced, not forwarded")
	}
	if body["lease"] != "240" {
		t.Errorf("lease = %#v, want the update to win", body["lease"])
	}
	// A value that came back from the router survives.
	if body["tagname"] != "office" {
		t.Errorf("tagname = %#v, want the current value preserved", body["tagname"])
	}

	// The prepared payload is what actually reaches the wire.
	if _, err := c.Put(context.Background(), "/network/dhcp/services/7", body); err != nil {
		t.Fatalf("Put: %v", err)
	}
	for _, key := range []string{"opt_type15", "opt_type121"} {
		if _, ok := sent[key]; !ok {
			t.Errorf("the wire payload is missing %s", key)
		}
	}
}

// TestVLANMirrorsNameOntoTagname covers the one mirror rule, and the fact
// that it overwrites rather than only filling.
func TestVLANMirrorsNameOntoTagname(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":20000,"data":[{
			"id":3,"vlan_name":"iot","tagname":"stale","interface":"lan2"}]}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, WithToken("t"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer func() { _ = c.Close(context.Background()) }()

	body, err := UpdatePayload(context.Background(), c, "/network/vlan", 3,
		map[string]any{"netmask": "255.255.255.0"}, nil)
	if err != nil {
		t.Fatalf("UpdatePayload: %v", err)
	}
	// The firmware copies vlan_name onto tagname unconditionally, so a
	// stale tagname read back from the router is replaced.
	if body["tagname"] != "iot" {
		t.Errorf("tagname = %#v, want it mirrored from vlan_name", body["tagname"])
	}
}

// TestObjectGroupsRequireGroupName checks the one field every object
// endpoint spells differently from the obvious guess.
func TestObjectGroupsRequireGroupName(t *testing.T) {
	for _, path := range []string{
		"/ip-objects", "/ip6-objects", "/mac-objects", "/port-objects",
		"/proto-objects", "/domain-objects", "/time-objects",
	} {
		if _, err := CreateBody(path, map[string]any{}); err == nil {
			t.Errorf("%s: expected a validation error with no group name", path)
		}
		body, err := CreateBody(path, map[string]any{"group_name": "servers"})
		if err != nil {
			t.Errorf("%s: %v", path, err)
			continue
		}
		if body["group_name"] != "servers" {
			t.Errorf("%s: group_name = %#v", path, body["group_name"])
		}
	}
}

// TestObjectGroupMembers checks the group_value encoding, including the
// IPv6 group whose member key is "ipv6" rather than "ip6".
func TestObjectGroupMembers(t *testing.T) {
	v6, err := ObjectGroup("ipv6", "2001:db8::1,2001:db8::2")
	if err != nil {
		t.Fatalf("ObjectGroup: %v", err)
	}
	if got := v6[0].(map[string]any)["ipv6"]; got != "2001:db8::1" {
		t.Errorf("IPv6 member key = %#v, want \"ipv6\"", v6[0])
	}
	// A comma list with gaps must not produce empty members.
	got, err := ObjectGroup("ip", "10.0.0.1,,10.0.0.2, ")
	if err != nil {
		t.Fatalf("ObjectGroup: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("members = %#v, want the empty entries dropped", got)
	}
}

func decode(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	return out
}
