package ikuaiapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// iKuai has several body-encoding conventions that a generic JSON client
// gets wrong by default. They are collected here because every write path
// in the firmware depends on them:
//
//   - Booleans are the strings "yes" / "no", never JSON true/false.
//   - A handful of numeric-looking fields must be JSON numbers, while most
//     must stay strings so values like an SNMP community "161" or a VLAN
//     name "0010" are not mangled.
//   - Address, port and protocol selectors are not a plain array; they are
//     a {"custom": [...], "object": [...]} pair.
//   - Object groups carry their members in a group_value array of
//     single-key objects, keyed by the object type.
//
// The helpers below encode those conventions so callers do not have to
// remember them per endpoint.

// integerAPIFields lists the body fields iKuai expects as JSON numbers.
//
// Everything outside this set is sent as a string on purpose: many iKuai
// fields accept numeric-looking strings, and coercing them would corrupt
// values such as the SNMP community "161" or a VLAN name "0010".
var integerAPIFields = map[string]bool{
	"sshd_port": true, "http_port": true, "https_port": true,
	"id": true, "language": true, "time_zone": true, "switch_nat": true,
	"switch_ntp": true, "switch_ntpd": true, "switch_ntpserver": true,
	"ntp_sync_cycle": true, "link_mode": true, "lan_nat": true, "fast_nat": true,
	"open_telnetd": true, "open_wanweb": true, "open_sshd": true,
	"open_ftp": true, "ftp_port": true, "ftp_access": true, "force_https": true,
	"ssl_on": true, "autoindex": true, "access": true, "wsdd2": true,
	"listen_port": true, "version": true,
	"interval": true, "auto_sync": true, "single_line": true,
	"ignore_wanstatus": true, "support_ftp": true, "support_tftp": true,
	"support_sip": true, "support_h323": true, "valid_days": true,
	"restore_type": true, "sync_bind_cloud": true, "group_id": true,
	"force": true, "sesstimeout": true, "turbo": true,
	"server_port": true, "open_mppe": true, "force_ipsec": true,
	"local_listenport": true, "endpoint_port": true, "method": true,
	"ikelifetime": true, "lifetime": true, "dpddelay": true, "dpdtimeout": true,
	"keepalive": true,
	"prio":      true, "priority": true,
	"syn_recv_timeout": true, "established_timeout": true,
	"upload": true, "download": true, "share": true, "packages": true,
	"expires": true, "start_time": true, "auto_mac": true, "auto_vlanid": true,
	"ip_type": true, "price": true, "up_speed": true, "down_speed": true,
	"ftp_ports": true, "sip_ports": true, "tftp_ports": true,
	"limits": true, "hit_rate": true,
	"authmode": true, "authport": true, "accountport": true,
	"force_verify_name": true, "rate_limit_lan": true, "drop_client": true,
	"force_pppoe": true, "enhance_check": true, "share_deny": true,
	"bind_vlan": true, "verify_vlan": true, "bind_iface": true,
	"mtu": true, "mru": true, "lcp_echo_interval": true,
	"lcp_echo_failure": true, "maxconnect": true, "restart_timer": true,
	"acl_mac": true, "mode": true,
	"vlanid": true, "channel": true,
	"type": true, "iface_band": true, "dst_type": true,
	"src_addr_inv": true, "dst_addr_inv": true,
	"noping_lan": true, "noping_wan": true, "notracert": true,
	"hijack_ping": true, "invalid": true, "dos_lan": true,
	"dos_lan_num": true, "tcp_mss": true, "tcp_mss_num": true,
	"nol2rt": true, "ttl_num": true,
}

// IsIntegerAPIField reports whether the named body field must be sent as a
// JSON number rather than a string.
func IsIntegerAPIField(name string) bool { return integerAPIFields[name] }

// IntegerAPIFields returns a copy of the integer-field allowlist. It exists
// so callers can audit or extend the classification without reaching into
// the SDK's internals.
func IntegerAPIFields() map[string]bool {
	out := make(map[string]bool, len(integerAPIFields))
	for k, v := range integerAPIFields {
		out[k] = v
	}
	return out
}

// CoerceBody normalizes a request body in place to match the firmware's
// expectations: every field listed in IntegerAPIFields is converted to an
// int64 when it can be parsed, and every other field is left as the caller
// supplied it. Fields that are already numbers are untouched.
//
// It is the programmatic form of what a CLI does when it merges flags into
// a --data payload, and is safe to call on any map destined for Create,
// Update or Patch.
func CoerceBody(body map[string]any) map[string]any {
	if body == nil {
		return nil
	}
	for name, value := range body {
		if !integerAPIFields[name] {
			continue
		}
		switch v := value.(type) {
		case string:
			if n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
				body[name] = n
			}
		case json.Number:
			// A number that came back from a read-modify-write arrives as a
			// json.Number. Leaving it would serialize fine but skip the
			// integer guarantee, so normalize it here.
			if n, err := strconv.ParseInt(v.String(), 10, 64); err == nil {
				body[name] = n
			}
		}
	}
	return body
}

// Enabled renders a Go bool the way the firmware expects it: the strings
// "yes" and "no", never a JSON boolean. Every rule-like resource in the
// catalog uses this convention for its "enabled" field.
func Enabled(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

// ParseEnabled reads an iKuai enable flag. It accepts the firmware's own
// "yes"/"no" plus the boolean spellings that show up in payloads read back
// from the router, so a round-tripped value can be fed straight back in.
func ParseEnabled(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "yes", "y", "true", "1", "on", "enable", "enabled":
		return true, nil
	case "no", "n", "false", "0", "off", "disable", "disabled":
		return false, nil
	}
	return false, fmt.Errorf("ikuaiapi: %q is not a valid enable value (want yes or no)", s)
}

// AddressValues builds the {"custom": [...], "object": [...]} payload iKuai
// uses for every address, port, protocol and domain selector. The input may
// be a single comma-separated string or discrete values; each entry is
// trimmed and empty entries are dropped.
//
// A rule with no selectors must send an empty pair rather than a bare
// array, which is why this returns a struct and not a []string.
func AddressValues(values ...string) map[string]any {
	return map[string]any{
		"custom": splitCSV(values),
		"object": []any{},
	}
}

// splitCSV flattens comma-separated values across all arguments, trimming
// each entry and dropping empties.
func splitCSV(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			if part = strings.TrimSpace(part); part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

// objectGroupKeys maps an iKuai object group to the key each member object
// is wrapped in inside group_value.
var objectGroupKeys = map[string]string{
	"ip":     "ip",
	"ip6":    "ipv6",
	"ipv6":   "ipv6",
	"mac":    "mac",
	"port":   "port",
	"proto":  "proto",
	"domain": "domain",
}

// ObjectGroup builds the group_value array for an object group. iKuai wraps
// every member in a single-key object named after the group type, so
// ObjectGroup("ip", "192.168.1.10", "192.168.1.11") produces
// [{"ip":"192.168.1.10"},{"ip":"192.168.1.11"}].
//
// The time object group is the exception: its members carry named fields
// rather than a single value, so use TimeObjectGroup for it.
func ObjectGroup(group string, values ...string) ([]any, error) {
	key, ok := objectGroupKeys[strings.ToLower(group)]
	if !ok {
		return nil, fmt.Errorf("ikuaiapi: unknown object group %q", group)
	}
	members := splitCSV(values)
	out := make([]any, 0, len(members))
	for _, m := range members {
		out = append(out, map[string]any{key: m})
	}
	return out, nil
}

// TimeObjectGroup builds a group_value member for the time object group,
// which is a weekly schedule rather than a single value.
func TimeObjectGroup(scheduleType, weekdays, startTime, endTime, comment string) map[string]any {
	return map[string]any{
		"type":       scheduleType,
		"weekdays":   weekdays,
		"start_time": startTime,
		"end_time":   endTime,
		"comment":    comment,
	}
}

// WeekSchedule is the time-object schedule iKuai applies to most rules
// when no explicit window is given: every day, all day.
func WeekSchedule() map[string]any {
	return TimeObjectGroup("weekly", "1234567", "00:00", "23:59", "")
}

// FirstObject unwraps the first element of an iKuai collection response.
// Endpoints disagree about where the array lives — some answer with
// "data", some with "results", some with "data" as a single object — so
// callers that need one element use this instead of re-implementing the
// probing logic for every read-modify-write update.
func FirstObject(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 {
		return nil, NewError(ErrEmptyResponse.Error())
	}
	var probe any
	if err := decodeUseNumber(raw, &probe); err != nil {
		return nil, NewError("decode response: "+err.Error(), WithCause(err))
	}
	return firstValue(probe)
}

// firstValue walks a decoded response down to its first object. The
// client's Get unwraps the envelope, so most callers hand this a bare
// array or a bare object; a full envelope is still accepted for callers
// that ran in RawMode or parsed the body themselves.
func firstValue(v any) (map[string]any, error) {
	switch t := v.(type) {
	case []any:
		if len(t) == 0 {
			return nil, NewError(ErrNoObject.Error())
		}
		return firstValue(t[0])
	case map[string]any:
		// An envelope still wrapping the collection?
		for _, key := range []string{"data", "results"} {
			if inner, ok := t[key]; ok && inner != nil {
				return firstValue(inner)
			}
		}
		// A third wrapping shape: several endpoints answer
		// {"<group>_data": [...], "<group>_total": n} instead of "data".
		// Observed on the object groups (/ip-objects → ip_data) and on the
		// DHCP static list (static_data). Without this the whole wrapper
		// comes back as the "object" and a read-modify-write would send
		// the counters back as if they were the row.
		for key, inner := range t {
			if strings.HasSuffix(key, "_data") && inner != nil {
				return firstValue(inner)
			}
		}
		if len(t) == 0 {
			return nil, NewError(ErrNoObject.Error())
		}
		return t, nil
	case nil:
		return nil, NewError(ErrNoObject.Error())
	default:
		return nil, NewError(ErrNoObject.Error())
	}
}

// MergeBody merges updates over an existing object for a read-modify-write
// update. Keys present in updates win; every other key from existing is
// preserved. Passing a nil existing behaves like a plain assignment, which
// makes the same code path usable for create and update.
func MergeBody(existing, updates map[string]any) map[string]any {
	out := make(map[string]any, len(existing)+len(updates))
	for k, v := range existing {
		out[k] = v
	}
	for k, v := range updates {
		out[k] = v
	}
	return out
}

// decodeUseNumber decodes JSON with numbers preserved as json.Number rather
// than float64.
//
// A router row id or a large subnet mask does not survive a round trip
// through float64, and a read-modify-write sends whatever came back. Using
// json.Number keeps the original literal all the way to the wire.
func decodeUseNumber(raw []byte, out any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	return dec.Decode(out)
}

// UpdatePayload prepares a body for a read-modify-write update: it fetches
// the current object, merges the updates over it, and drops the fields the
// firmware rejects on an update. This is what the iKuai web UI does, and
// several endpoints (DHCP services, VLAN, NAT rules, PPPoE) refuse a
// partial body rather than defaulting the missing fields.
//
// The assembly order matches the firmware: caller updates win, then the
// create defaults fill what is still missing, then the update-only defaults
// do, then the endpoint's mirror rules run. Nothing overwrites a value the
// caller supplied, except a mirror rule explicitly marked Overwrite — the
// VLAN endpoint requires that, because its rule name and display name are
// the same field to the firmware.
//
// keep lists the body fields the endpoint accepts. Pass nil to use the
// endpoint's recorded allowlist from V4Schemas; if the endpoint has no
// recorded allowlist, every field is kept.
func UpdatePayload(ctx context.Context, c *Client, path string, id int64,
	updates map[string]any, keep []string) (map[string]any, error) {

	if keep == nil {
		keep = UpdateFields(path)
	}

	current, err := fetchOne(ctx, c, path, id)
	if err != nil {
		return nil, err
	}
	// Values read back from the router are normalized first, so the
	// endpoint default fills them rather than a blank or a derived field
	// surviving.
	cleanReadBack(current, path)
	merged := MergeBody(current, updates)

	// Fill the defaults before the allowlist runs, so a default that the
	// endpoint does not accept on an update is dropped rather than sent.
	if s, ok := V4Schemas[path]; ok {
		merged = fillDefaults(merged, s.CreateDefaults)
		merged = fillDefaults(merged, s.UpdateDefaults)
	}

	// "id" identifies the row and the firmware rejects it on an update, so
	// it is dropped whether or not the endpoint declares an allowlist.
	delete(merged, "id")

	if keep == nil {
		return CoerceBody(merged), nil
	}
	allowed := make(map[string]bool, len(keep))
	for _, k := range keep {
		allowed[k] = true
	}
	out := make(map[string]any, len(merged))
	for k, v := range merged {
		if allowed[k] {
			out[k] = v
		}
	}
	// Mirrors run after the allowlist so a rule can reintroduce a field the
	// allowlist dropped, which is what the firmware itself does.
	applyMirrorRules(out, path)
	return CoerceBody(out), nil
}

// cleanReadBack prepares the object the router returned before the
// caller's updates are merged over it. It runs only on that object, so a
// value the caller supplied on purpose is never rewritten.
//
// Three normalizations happen here, each because the firmware mishandles
// the alternative on write:
//   - a blank DropBlanks field is dropped, so the default fills it rather
//     than an empty option value going out;
//   - a DropSuffix field is dropped, because it is derived for display and
//     is rejected on write;
//   - an empty selector pair becomes an empty array, because the firmware
//     sometimes reads back an object where it wrote an array.
func cleanReadBack(body map[string]any, path string) {
	s, ok := V4Schemas[path]
	if !ok {
		return
	}
	for _, key := range s.DropBlanks {
		if v, present := body[key]; present {
			if str, isString := v.(string); isString && str == "" {
				delete(body, key)
			}
		}
	}
	for _, suffix := range s.DropSuffix {
		for key := range body {
			if key != suffix && strings.HasSuffix(key, suffix) {
				delete(body, key)
			}
		}
	}
	if s.NormalizeSelectors {
		for _, v := range body {
			normalizeSelector(v)
		}
	}
}

// normalizeSelector rewrites {"custom":{}, "object":{}} into the array
// form the firmware expects. It recurses, since selectors nest inside a
// rule body.
func normalizeSelector(v any) {
	switch t := v.(type) {
	case map[string]any:
		for _, half := range []string{"custom", "object"} {
			if inner, ok := t[half]; ok {
				if m, isMap := inner.(map[string]any); isMap && len(m) == 0 {
					t[half] = []any{}
				}
			}
		}
		for _, child := range t {
			normalizeSelector(child)
		}
	case []any:
		for _, child := range t {
			normalizeSelector(child)
		}
	}
}

// fillDefaults adds every default whose key the body does not already
// carry. An existing value always wins, whatever its type.
func fillDefaults(body, defaults map[string]any) map[string]any {
	for k, v := range defaults {
		if _, ok := body[k]; !ok {
			body[k] = deepCopyValue(v)
		}
	}
	return body
}

// applyMirrorRules copies fields onto each other as the endpoint's contract
// requires. It is exported behaviour through UpdatePayload; callers that
// build an update by hand can apply it to their own map.
func applyMirrorRules(body map[string]any, path string) {
	s, ok := V4Schemas[path]
	if !ok {
		return
	}
	for _, rule := range s.Mirror {
		v, ok := body[rule.From]
		if !ok || v == nil {
			continue
		}
		if _, taken := body[rule.To]; taken && !rule.Overwrite {
			continue
		}
		body[rule.To] = v
	}
}

// fetchOne reads a single item, preferring the path-segment form the
// firmware uses and falling back to an ?id= query.
func fetchOne(ctx context.Context, c *Client, path string, id int64) (map[string]any, error) {
	idStr := strconv.FormatInt(id, 10)
	raw, err := c.Get(ctx, path+"/"+idStr, nil)
	if err != nil {
		raw, err = c.Get(ctx, path, map[string]string{"id": idStr})
		if err != nil {
			return nil, err
		}
	}
	return FirstObject(raw)
}
