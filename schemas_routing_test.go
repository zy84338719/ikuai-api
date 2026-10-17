package ikuaiapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// TestRoutingStripsDerivedIntFields checks the rule that matters most on a
// route update: the router returns a "_int" companion for several fields and
// rejects a body carrying one back.
func TestRoutingStripsDerivedIntFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":20000,"data":[{
			"id":4,"tagname":"to_lan2","dst_addr":"10.10.10.0",
			"gateway":"10.66.0.1","netmask":"255.255.255.0","interface":"wan1",
			"prio":1,"prio_int":0,"dst_addr_int":0,"enabled":"yes",
			"comment":""}]}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, WithToken("t"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer func() { _ = c.Close(context.Background()) }()

	body, err := UpdatePayload(context.Background(), c, "/routing/static-routes", 4,
		map[string]any{"comment": "updated"}, nil)
	if err != nil {
		t.Fatalf("UpdatePayload: %v", err)
	}
	for _, key := range []string{"prio_int", "dst_addr_int"} {
		if _, ok := body[key]; ok {
			t.Errorf("derived field %s must not be sent back", key)
		}
	}
	// prio is on the integer allowlist, so the read-back number is coerced.
	if body["prio"] != int64(1) {
		t.Errorf("prio = %#v, want the current value as int64(1)", body["prio"])
	}
	if body["comment"] != "updated" {
		t.Errorf("comment = %#v, want the caller's value", body["comment"])
	}
	if _, ok := body["id"]; ok {
		t.Error("id must be dropped")
	}
}

// TestSelectorObjectBecomesArray covers the normalization that stops a
// read-modify-write from failing: the firmware sometimes reads back
// {"custom":{},"object":{}} where it wrote arrays, and rejects the object
// form on the next write.
func TestSelectorObjectBecomesArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":20000,"data":[{
			"id":9,"tagname":"web",
			"src_addr":{"custom":{},"object":{}},
			"dst_port":{"custom":[{"port":"80"}],"object":{}}}]}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, WithToken("t"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer func() { _ = c.Close(context.Background()) }()

	body, err := UpdatePayload(context.Background(), c, "/routing/static-routes", 9,
		map[string]any{"comment": "x"}, nil)
	if err != nil {
		t.Fatalf("UpdatePayload: %v", err)
	}

	src, ok := body["src_addr"].(map[string]any)
	if !ok {
		t.Fatalf("src_addr = %#v, want a selector pair", body["src_addr"])
	}
	for _, half := range []string{"custom", "object"} {
		if _, isObject := src[half].(map[string]any); isObject {
			t.Errorf("src_addr.%s = %#v, want an array", half, src[half])
		}
		if _, isArray := src[half].([]any); !isArray {
			t.Errorf("src_addr.%s = %#v, want an empty array", half, src[half])
		}
	}

	// A selector that actually holds members keeps them; only the empty half
	// is rewritten.
	dst, ok := body["dst_port"].(map[string]any)
	if !ok {
		t.Fatalf("dst_port = %#v, want a selector pair", body["dst_port"])
	}
	custom, ok := dst["custom"].([]any)
	if !ok || len(custom) != 1 {
		t.Errorf("dst_port.custom = %#v, want the member preserved", dst["custom"])
	}
	if _, isObject := dst["object"].(map[string]any); isObject {
		t.Errorf("dst_port.object = %#v, want an array", dst["object"])
	}
}

// TestSingletonConfigMerge covers a PUT endpoint that addresses one
// document: there is no id, and a partial body would let the firmware reset
// settings the caller never mentioned.
func TestSingletonConfigMerge(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":20000,"data":{
			"dns1":"192.168.1.1","dns2":"8.8.8.8","cachemode":1,
			"cache_ttl":3600,"defense":1,"network":"udp"}}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, WithToken("t"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer func() { _ = c.Close(context.Background()) }()

	body, err := UpdatePayload(context.Background(), c, "/network/dns/config", 0,
		map[string]any{"dns1": "223.5.5.5"}, nil)
	if err != nil {
		t.Fatalf("UpdatePayload: %v", err)
	}
	if body["dns1"] != "223.5.5.5" {
		t.Errorf("dns1 = %#v, want the caller's value", body["dns1"])
	}
	// Everything else the router returned is carried forward, so the write
	// cannot reset it.
	for _, key := range []string{"dns2", "cachemode", "cache_ttl", "defense", "network"} {
		if _, ok := body[key]; !ok {
			t.Errorf("%s was dropped; a DNS change must carry the whole document", key)
		}
	}
}

// TestDHCPv4AccessRulesMatchIPv6 guards the two rule sets against drifting
// apart — they take the same body and differ only in path.
func TestDHCPv4AccessRulesMatchIPv6(t *testing.T) {
	v4, ok4 := Schema("/network/dhcp/access-control/rules")
	v6, ok6 := Schema("/network/dhcp6/access-control/rules")
	if !ok4 || !ok6 {
		t.Fatal("both access-control rule sets should carry a contract")
	}
	if !reflect.DeepEqual(v4.InputFields, v6.InputFields) {
		t.Errorf("v4 %v and v6 %v rule bodies must match", v4.InputFields, v6.InputFields)
	}
	body, err := CreateBody("/network/dhcp/access-control/rules", map[string]any{
		"tagname": "block-bad-mac", "mac": "00:11:22:33:44:55",
	})
	if err != nil {
		t.Fatalf("CreateBody: %v", err)
	}
	if body["enabled"] != "yes" {
		t.Errorf("enabled = %#v, want the default", body["enabled"])
	}
}

// TestReadModifyWriteKeepsLargeIntegers is the reason the read path decodes
// with json.Number. A row id or a 64-bit counter that round-trips through
// float64 comes back altered, and an update sends whatever came back — so
// the corruption lands on the router.
func TestReadModifyWriteKeepsLargeIntegers(t *testing.T) {
	const bigID = int64(9007199254740993) // 2^53 + 1, not representable as float64

	var sent map[string]any
	var rawBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":20000,"data":[{"id":1,"tagname":"keep","counter":9007199254740993}]}`))
			return
		}
		b := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(b)
		rawBody = string(b)
		var out map[string]any
		if err := decodeUseNumber(b, &out); err != nil {
			t.Errorf("request body is not valid JSON: %v", err)
		}
		sent = out
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":20000,"message":"ok"}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, WithToken("t"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer func() { _ = c.Close(context.Background()) }()

	body, err := UpdatePayload(context.Background(), c, "/routing/static-routes", 1,
		map[string]any{"comment": "x"}, nil)
	if err != nil {
		t.Fatalf("UpdatePayload: %v", err)
	}
	if got, ok := body["counter"].(json.Number); !ok || got.String() != "9007199254740993" {
		t.Errorf("counter = %#v, want the exact literal 9007199254740993", body["counter"])
	}

	if _, err := c.Put(context.Background(), "/routing/static-routes/1", body); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if _, ok := sent["counter"]; !ok {
		t.Error("counter did not reach the wire")
	}
	// Check the serialized bytes, not a re-decoded float64: the point is
	// that the literal survives to the wire intact.
	if !strings.Contains(rawBody, "9007199254740993") {
		t.Errorf("the wire body is %s, want the exact literal 9007199254740993", rawBody)
	}
	if strings.Contains(rawBody, "e+15") {
		t.Errorf("the wire body lost precision: %s", rawBody)
	}
	_ = bigID
	_ = fmt.Sprint
}
