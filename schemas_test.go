package ikuaiapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestSchemaCoverageMatchesCatalog(t *testing.T) {
	// Every schema path must be a real endpoint. A schema keyed on a path
	// the router does not serve would silently never be consulted.
	catalog := map[string]bool{}
	for _, ep := range V4EndpointCatalog {
		catalog[ep.Path] = true
	}
	for _, path := range SchemaPaths() {
		if !catalog[path] {
			t.Errorf("schema %q has no matching endpoint in the catalog", path)
		}
	}
	if len(V4Schemas) < 40 {
		t.Errorf("only %d schemas recorded, expected the full rule-bearing set", len(V4Schemas))
	}
}

func TestCreateBodyAppliesDefaults(t *testing.T) {
	body, err := CreateBody("/security/acl-rules", map[string]any{
		"tagname":  "block_ssh",
		"action":   "drop",
		"protocol": "tcp",
		"dst_port": AddressValues("22"),
	})
	if err != nil {
		t.Fatalf("CreateBody: %v", err)
	}

	if body["enabled"] != "yes" {
		t.Errorf("enabled = %#v, want the \"yes\" default", body["enabled"])
	}
	if body["prio"] != 50 {
		t.Errorf("prio = %#v, want the ACL default 50", body["prio"])
	}
	if body["dir"] != "forward" {
		t.Errorf("dir = %#v, want \"forward\"", body["dir"])
	}
	// The caller's override must win over the default.
	sel, ok := body["dst_port"].(map[string]any)
	if !ok {
		t.Fatalf("dst_port = %#v, want a selector pair", body["dst_port"])
	}
	if custom, _ := sel["custom"].([]string); len(custom) != 1 || custom[0] != "22" {
		t.Errorf("dst_port custom = %#v, want [22]", sel["custom"])
	}
	// An untouched selector stays the empty pair the firmware expects.
	src, _ := body["src_addr"].(map[string]any)
	if _, ok := src["object"]; !ok {
		t.Error("src_addr must keep both selector keys")
	}
	// Every rule carries the all-day weekly window unless overridden.
	win, _ := body["time"].(map[string]any)
	if win == nil || win["custom"] == nil {
		t.Errorf("time = %#v, want the weekly default window", body["time"])
	}
}

func TestCreateBodyIsDeepCopied(t *testing.T) {
	// Two callers must never share structure with the contract table, or
	// the first call's overrides would leak into the second.
	first, err := CreateBody("/security/acl-rules", map[string]any{
		"tagname": "a", "action": "drop", "protocol": "tcp",
		"dst_port": AddressValues("22"),
	})
	if err != nil {
		t.Fatalf("CreateBody: %v", err)
	}
	first["tagname"] = "mutated"
	firstSel, _ := first["dst_port"].(map[string]any)
	if custom, _ := firstSel["custom"].([]string); len(custom) > 0 {
		custom[0] = "9999"
	}

	second, err := CreateBody("/security/acl-rules", map[string]any{
		"tagname": "b", "action": "drop", "protocol": "tcp",
	})
	if err != nil {
		t.Fatalf("CreateBody: %v", err)
	}
	if second["tagname"] != "b" {
		t.Errorf("second create saw a mutated tagname: %#v", second["tagname"])
	}
	secondSel, _ := second["dst_port"].(map[string]any)
	if custom, _ := secondSel["custom"].([]string); len(custom) != 0 {
		t.Errorf("selector leaked between calls: %#v", secondSel["custom"])
	}
}

func TestCreateBodyRequiredFields(t *testing.T) {
	// Missing required fields must fail here with a readable message
	// rather than becoming a confusing 3001 from the router.
	if _, err := CreateBody("/security/acl-rules", map[string]any{"tagname": "x"}); err == nil {
		t.Error("expected an error when required fields are missing")
	} else if !strings.Contains(err.Error(), "action") || !strings.Contains(err.Error(), "protocol") {
		t.Errorf("error should name the missing fields, got %v", err)
	}

	// An empty string does not count as supplied.
	_, err := CreateBody("/security/acl-rules", map[string]any{
		"tagname": "x", "action": "  ", "protocol": "tcp",
	})
	if err == nil {
		t.Error("a blank required field should be reported as missing")
	}
}

func TestCreateBodyCoercesIntegers(t *testing.T) {
	body, err := CreateBody("/routing/domain-rules", map[string]any{
		"tagname":   "baidu",
		"interface": "wan2",
		"prio":      "40", // allowlisted → number
		"domain":    AddressValues("www.baidu.com"),
	})
	if err != nil {
		t.Fatalf("CreateBody: %v", err)
	}
	if body["prio"] != int64(40) {
		t.Errorf("prio = %#v, want int64(40)", body["prio"])
	}
}

func TestCreateBodyUnknownEndpointFallsThrough(t *testing.T) {
	body, err := CreateBody("/not/in/catalog", map[string]any{"tagname": "x"})
	if err != nil {
		t.Fatalf("an unknown endpoint should still merge overrides: %v", err)
	}
	if len(body) != 1 || body["tagname"] != "x" {
		t.Errorf("body = %#v, want just the override", body)
	}
}

func TestSchemaPreservesFirmwareInconsistencies(t *testing.T) {
	// The auth-accounts src_addr is a bare map with no "object" key.
	// Reproducing it is deliberate: normalizing it would break the write.
	body, err := CreateBody("/auth/users", map[string]any{
		"username": "guest1", "passwd": "secret",
	})
	if err != nil {
		t.Fatalf("CreateBody: %v", err)
	}
	src, ok := body["src_addr"].(map[string]any)
	if !ok {
		t.Fatalf("src_addr = %#v, want a map", body["src_addr"])
	}
	if _, hasObject := src["object"]; hasObject {
		t.Error("auth accounts src_addr must not gain an object key; the firmware does not use one")
	}
	if src["custom"] == nil {
		t.Error("auth accounts src_addr.custom must be present")
	}

	// The load-balance weight is a string, unlike most numeric fields.
	lb, err := CreateBody("/routing/load-balance-rules", map[string]any{
		"tagname": "dual-wan", "interface": "wan1",
	})
	if err != nil {
		t.Fatalf("CreateBody: %v", err)
	}
	if lb["weight"] != "1" {
		t.Errorf("weight = %#v, want the string \"1\"", lb["weight"])
	}
}

func TestWirelessUsesFlatWindow(t *testing.T) {
	// The wireless blacklist keeps a flat "00:00-23:59" string plus a
	// weekday, not the nested weekly structure used elsewhere.
	body, err := CreateBody("/wireless/access-control/rules", map[string]any{
		"tagname": "block1", "lmac": AddressValues("00:11:22:33:44:55"),
	})
	if err != nil {
		t.Fatalf("CreateBody: %v", err)
	}
	if body["time"] != "00:00-23:59" {
		t.Errorf("time = %#v, want the flat string", body["time"])
	}
	if body["week"] != "1234567" {
		t.Errorf("week = %#v, want \"1234567\"", body["week"])
	}
	if _, isMap := body["time"].(map[string]any); isMap {
		t.Error("wireless time must stay a flat string")
	}
}

func TestURLBlackHasEmptyTimeSelector(t *testing.T) {
	// The one rule that carries an empty selector pair instead of the
	// all-day weekly window.
	body, err := CreateBody("/security/url-black/rules", map[string]any{
		"tagname": "kw1", "mode": "exact", "domain": AddressValues("bad.example"),
	})
	if err != nil {
		t.Fatalf("CreateBody: %v", err)
	}
	sel, ok := body["time"].(map[string]any)
	if !ok {
		t.Fatalf("time = %#v, want a selector pair", body["time"])
	}
	if custom, _ := sel["custom"].([]any); len(custom) != 0 {
		t.Errorf("url-black time.custom = %#v, want empty", sel["custom"])
	}
}

func TestUpdateFieldsAllowlist(t *testing.T) {
	fields := UpdateFields("/network/nat/rules")
	if len(fields) == 0 {
		t.Fatal("NAT rules should carry an update allowlist")
	}
	s, _ := Schema("/network/nat/rules")
	if !s.AllowsField("action") {
		t.Error("action should be updatable on NAT rules")
	}
	if s.AllowsField("some_injected_key") {
		t.Error("an unknown key should not be updatable on NAT rules")
	}
	// An endpoint with no recorded contract at all accepts anything.
	open, _ := Schema("/network/ac/services")
	if open.HasInputFields() {
		t.Error("/network/ac/services should have no recorded allowlist")
	}
	if !open.AllowsField("anything") {
		t.Error("an endpoint without an allowlist must accept every field")
	}
}

func TestUpdatePayloadUsesSchemaAllowlistByDefault(t *testing.T) {
	var sentBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"code":20000,"data":[{
				"id":3,"tagname":"old","action":"filter","iinterface":"wan1",
				"ointerface":"wan2","comment":"keep","injected_key":"dropme"}]}`))
			return
		}
		b := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(b)
		sentBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":20000,"message":"ok"}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, WithToken("t"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer func() { _ = c.Close(context.Background()) }()

	// keep == nil must fall back to the schema allowlist.
	body, err := UpdatePayload(context.Background(), c, "/network/nat/rules", 3,
		map[string]any{"nat_addr": "203.0.113.5"}, nil)
	if err != nil {
		t.Fatalf("UpdatePayload: %v", err)
	}
	if body["nat_addr"] != "203.0.113.5" {
		t.Errorf("nat_addr = %#v, want the update to apply", body["nat_addr"])
	}
	if body["comment"] != "keep" {
		t.Errorf("comment = %#v, want the current value preserved", body["comment"])
	}
	if _, ok := body["injected_key"]; ok {
		t.Error("a field outside the schema allowlist must be dropped")
	}
	if _, ok := body["id"]; ok {
		t.Error("id must be dropped on update")
	}

	if _, err := c.Put(context.Background(), "/network/nat/rules/3", body); err != nil {
		t.Fatalf("Put: %v", err)
	}
	var sent map[string]any
	if err := json.Unmarshal([]byte(sentBody), &sent); err != nil {
		t.Fatalf("the prepared payload must serialize: %v", err)
	}
	if _, ok := sent["injected_key"]; ok {
		t.Error("a rejected key reached the wire")
	}
}

func TestUpdatePayloadExplicitAllowlistWins(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":20000,"data":[{"id":4,"tagname":"t","extra":"x"}]}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, WithToken("t"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer func() { _ = c.Close(context.Background()) }()

	// An explicit keep list must not be overridden by the schema.
	body, err := UpdatePayload(context.Background(), c, "/network/nat/rules", 4,
		map[string]any{"nat_addr": "1.1.1.1"}, []string{"tagname", "extra"})
	if err != nil {
		t.Fatalf("UpdatePayload: %v", err)
	}
	if body["extra"] != "x" {
		t.Errorf("extra = %#v, want the explicit allowlist to keep it", body["extra"])
	}
	if _, ok := body["nat_addr"]; ok {
		t.Error("nat_addr is outside the explicit allowlist and should be dropped")
	}
}

func TestSchemaSelectorShapesAreValid(t *testing.T) {
	// Every selector pair in the table must carry both keys, or a default
	// create would be rejected for a missing "object".
	for path, s := range V4Schemas {
		for field, v := range s.CreateDefaults {
			m, ok := v.(map[string]any)
			if !ok {
				continue
			}
			if _, isSelector := m["custom"]; !isSelector {
				continue
			}
			if _, hasObject := m["object"]; !hasObject {
				// The auth-accounts src_addr is the documented exception.
				if path == "/auth/users" && field == "src_addr" {
					continue
				}
				t.Errorf("%s.%s: selector pair is missing the object key", path, field)
			}
		}
	}
}

func TestDeepCopyHandlesNestedStructures(t *testing.T) {
	src := map[string]any{
		"a": map[string]any{"b": []any{map[string]any{"c": 1}}},
		"d": []any{1, 2, 3},
		"e": "plain",
		"f": 42,
	}
	got := deepCopyMap(src)
	if !reflect.DeepEqual(src, got) {
		t.Fatalf("copy differs: %#v vs %#v", src, got)
	}
	got["a"].(map[string]any)["b"].([]any)[0].(map[string]any)["c"] = 99
	if src["a"].(map[string]any)["b"].([]any)[0].(map[string]any)["c"] != 1 {
		t.Error("deep copy shares nested structure with the original")
	}
}
