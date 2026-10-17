package ikuaiapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestEnabledAndParseEnabled(t *testing.T) {
	if got := Enabled(true); got != "yes" {
		t.Errorf("Enabled(true) = %q, want \"yes\"", got)
	}
	if got := Enabled(false); got != "no" {
		t.Errorf("Enabled(false) = %q, want \"no\"", got)
	}
	for _, in := range []string{"yes", "YES", " yes ", "1", "true", "on", "enabled"} {
		v, err := ParseEnabled(in)
		if err != nil || !v {
			t.Errorf("ParseEnabled(%q) = %v, %v; want true, nil", in, v, err)
		}
	}
	for _, in := range []string{"no", "NO", "0", "false", "off", "disabled"} {
		v, err := ParseEnabled(in)
		if err != nil || v {
			t.Errorf("ParseEnabled(%q) = %v, %v; want false, nil", in, v, err)
		}
	}
	if _, err := ParseEnabled("maybe"); err == nil {
		t.Error("expected an error for an unrecognized enable value")
	}
}

func TestAddressValues(t *testing.T) {
	got := AddressValues("80, 443 ,", "8080")
	want := map[string]any{"custom": []string{"80", "443", "8080"}, "object": []any{}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("AddressValues = %#v, want %#v", got, want)
	}
	// An empty selector must still produce the pair, never a bare array,
	// because the firmware rejects a missing "object" key.
	empty := AddressValues()
	if _, ok := empty["object"]; !ok {
		t.Error("AddressValues() must always include the object key")
	}
	if c, ok := empty["custom"].([]string); !ok || len(c) != 0 {
		t.Errorf("AddressValues() custom = %#v, want empty", empty["custom"])
	}
}

func TestObjectGroup(t *testing.T) {
	got, err := ObjectGroup("ip", "192.168.1.10,192.168.1.11")
	if err != nil {
		t.Fatalf("ObjectGroup: %v", err)
	}
	want := []any{
		map[string]any{"ip": "192.168.1.10"},
		map[string]any{"ip": "192.168.1.11"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ObjectGroup = %#v, want %#v", got, want)
	}
	if _, err := ObjectGroup("nope", "x"); err == nil {
		t.Error("expected an error for an unknown object group")
	}
}

func TestCoerceBody(t *testing.T) {
	body := map[string]any{
		// allowlisted: coerced to a number
		"mtu":      "1500",
		"prio":     "31",
		"priority": "not-a-number",
		// not allowlisted: left as a string even though it looks numeric
		"community": "161",
		"vlan_name": "0010",
		"comment":   "42",
		// already a number: untouched
		"id": 7,
	}
	out := CoerceBody(body)
	if out["mtu"] != int64(1500) {
		t.Errorf("mtu = %#v, want int64(1500)", out["mtu"])
	}
	if out["prio"] != int64(31) {
		t.Errorf("prio = %#v, want int64(31)", out["prio"])
	}
	if out["priority"] != "not-a-number" {
		t.Errorf("priority = %#v, want the original string", out["priority"])
	}
	for _, k := range []string{"community", "vlan_name", "comment"} {
		if _, ok := out[k].(string); !ok {
			t.Errorf("%s = %#v, want it left as a string", k, out[k])
		}
	}
	if out["id"] != 7 {
		t.Errorf("id = %#v, want 7", out["id"])
	}
	if CoerceBody(nil) != nil {
		t.Error("CoerceBody(nil) should stay nil")
	}
}

func TestIntegerAPIFieldClassification(t *testing.T) {
	if !IsIntegerAPIField("mtu") {
		t.Error("mtu should be an integer field")
	}
	if IsIntegerAPIField("community") {
		t.Error("community must stay a string")
	}
	fields := IntegerAPIFields()
	delete(fields, "mtu")
	if !IsIntegerAPIField("mtu") {
		t.Error("IntegerAPIFields must return a copy, not the live map")
	}
}

func TestParseErrorCodes(t *testing.T) {
	err := &APIError{Code: 3007, Message: "Invalid token"}
	if !err.IsAuthFailure() {
		t.Error("3007 should be an auth failure")
	}
	if err.IsPermissionFailure() {
		t.Error("3007 is not a permission failure")
	}
	if err.Remedy() == "" {
		t.Error("3007 should carry a remedy")
	}

	perm := &APIError{Code: 10014, Message: "Permission denied"}
	if perm.IsAuthFailure() {
		t.Error("10014 is not an auth failure")
	}
	if !perm.IsPermissionFailure() {
		t.Error("10014 should be a permission failure")
	}
	if perm.Remedy() == "" {
		t.Error("10014 should carry a remedy")
	}

	// The client returns the data field unwrapped, so the same helper must
	// cope with a bare array and a bare object too.
	bareArray, aerr := FirstObject(json.RawMessage(`[{"id":4,"tagname":"d"}]`))
	if aerr != nil {
		t.Fatalf("FirstObject on a bare array: %v", aerr)
	}
	if bareArray["tagname"] != "d" {
		t.Errorf("bare array: tagname = %#v, want \"d\"", bareArray["tagname"])
	}
	bareObject, berr := FirstObject(json.RawMessage(`{"id":5,"tagname":"e"}`))
	if berr != nil {
		t.Fatalf("FirstObject on a bare object: %v", berr)
	}
	if bareObject["tagname"] != "e" {
		t.Errorf("bare object: tagname = %#v, want \"e\"", bareObject["tagname"])
	}

	unknown := &APIError{Code: 99999, Message: "?"}
	if unknown.Remedy() != "" {
		t.Errorf("an unknown code should have no remedy, got %q", unknown.Remedy())
	}
}

func TestFirstObject(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{"data array", `{"code":20000,"data":[{"id":1,"tagname":"a"}]}`},
		{"results array", `{"code":20000,"results":[{"id":2,"tagname":"b"}]}`},
		{"single object", `{"code":20000,"data":{"id":3,"tagname":"c"}}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FirstObject(json.RawMessage(tc.raw))
			if err != nil {
				t.Fatalf("FirstObject: %v", err)
			}
			if got["tagname"] == nil {
				t.Errorf("FirstObject = %#v, want the tagname field", got)
			}
		})
	}
	if _, err := FirstObject(json.RawMessage(`{"code":20000,"data":[]}`)); err == nil {
		t.Error("expected an error for an empty collection")
	}
	if _, err := FirstObject(nil); err == nil {
		t.Error("expected an error for an empty body")
	}
}

func TestMergeBody(t *testing.T) {
	existing := map[string]any{"id": int64(5), "tagname": "old", "comment": "keep"}
	updates := map[string]any{"tagname": "new"}
	got := MergeBody(existing, updates)
	if got["tagname"] != "new" {
		t.Errorf("tagname = %#v, want the update to win", got["tagname"])
	}
	if got["comment"] != "keep" {
		t.Errorf("comment = %#v, want it preserved", got["comment"])
	}
	if existing["tagname"] != "old" {
		t.Error("MergeBody must not mutate the existing map")
	}
}

func TestUpdatePayloadReadModifyWrite(t *testing.T) {
	var seenPath, seenBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":20000,"data":[{"id":7,"tagname":"old","comment":"keep","junk":"drop","addr_pool":"192.168.1.100"}]}`))
			return
		}
		b := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(b)
		seenBody = string(b)
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
		map[string]any{"tagname": "new"},
		[]string{"tagname", "comment", "addr_pool"})
	if err != nil {
		t.Fatalf("UpdatePayload: %v", err)
	}
	if body["tagname"] != "new" {
		t.Errorf("tagname = %#v, want the update to win", body["tagname"])
	}
	if body["comment"] != "keep" {
		t.Errorf("comment = %#v, want the current value preserved", body["comment"])
	}
	if _, ok := body["junk"]; ok {
		t.Error("fields outside the allowlist must be dropped")
	}
	if _, ok := body["id"]; ok {
		t.Error("id must be dropped: the firmware rejects it on update")
	}
	if seenPath != "/api/v4.0/network/dhcp/services/7" {
		t.Errorf("read path = %q, want the path-segment form", seenPath)
	}

	// The prepared payload must be sendable as-is.
	if _, err := c.Put(context.Background(), "/network/dhcp/services/7", body); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if seenBody == "" {
		t.Error("expected the write request to carry a body")
	}
}
