package ikuaiapi

import (
	"fmt"
	"strings"
)

// QueryDialect names the query-parameter contract an endpoint speaks. The
// iKuai v4 surface is not uniform: several dialects show up across the
// monitoring and auth groups, and picking the wrong one makes the router
// silently ignore the whole query string.
//
// The dialects, as observed on device and cross-checked against the
// upstream CLI implementation:
//
//	DialectPaged — page/limit/filter/order/order_by (the default for most
//	  collections), optionally narrowed further by key/pattern.
//	DialectLegacy — the uppercase dialect used by /monitoring/downstream,
//	  /monitoring/cameras and /monitoring/switch. It wants TYPE=data,
//	  LIMIT="<offset>,<count>", OFFSET, ORDER, ORDER_BY and KEYWORDS.
//	  There is no "page"; the caller computes the offset.
//	DialectMixed — /auth/online-users takes lowercase page/limit but
//	  uppercase KEYWORDS/FINDS/ORDER/ORDER_BY.
//	DialectNone — no paging at all; the endpoint returns everything.
//	DialectLoad — the datetype/start_time/end_time/math window used by the
//	  monitoring load endpoints.
//	DialectWindow — starttime/stoptime (no underscore) plus optional
//	  per-endpoint selectors such as ip/mac/appids.
type QueryDialect string

const (
	// DialectPaged is the default lowercase page/limit dialect.
	DialectPaged QueryDialect = ""
	// DialectLegacy is the uppercase TYPE/LIMIT/OFFSET dialect.
	DialectLegacy QueryDialect = "legacy"
	// DialectMixed is lowercase paging with uppercase sort/search keys.
	DialectMixed QueryDialect = "mixed"
	// DialectNone sends no paging parameters.
	DialectNone QueryDialect = "none"
	// DialectLoad is the datetype/start_time/end_time/math window.
	DialectLoad QueryDialect = "load"
	// DialectWindow is the starttime/stoptime window.
	DialectWindow QueryDialect = "window"
	// DialectRef is a reverse lookup: one required name parameter, no
	// paging. It answers "which rules reference this object group".
	DialectRef QueryDialect = "ref"
)

// ParseQuerySpec decodes the compact Query spec carried by a catalog entry.
//
// The grammar is:
//
//	spec    := clause ("+" clause)*
//	clause  := dialect | "req:" field ("," field)* | "opt:" field ("," field)*
//	dialect := "paged" | "legacy" | "mixed" | "none" | "load" |
//	            "window" | "ref"
//
// field is the literal query-key name as the router expects it, e.g. "ip",
// "mac", "KEYWORDS". A numeric field is suffixed with "#", which makes the
// generated Options struct use an int and render it without a decimal
// point. An empty string decodes to the zero value: the default paged
// dialect with no extra parameters.
//
// It returns an error for an unknown dialect or a malformed clause, so a
// typo in v4_catalog.go fails the build instead of silently producing a
// query the router ignores.
func ParseQuerySpec(spec string) (QueryDialect, []QueryField, []QueryField, error) {
	dialect := DialectPaged
	var required, optional []QueryField

	for _, clause := range strings.Split(spec, "+") {
		switch {
		case clause == "":
			continue
		case strings.HasPrefix(clause, "req:"):
			fields, err := parseQueryFields(strings.TrimPrefix(clause, "req:"), true)
			if err != nil {
				return "", nil, nil, err
			}
			required = append(required, fields...)
		case strings.HasPrefix(clause, "opt:"):
			fields, err := parseQueryFields(strings.TrimPrefix(clause, "opt:"), false)
			if err != nil {
				return "", nil, nil, err
			}
			optional = append(optional, fields...)
		default:
			d := QueryDialect(clause)
			if clause == "paged" {
				// "paged" is the explicit spelling of the default dialect.
				d = DialectPaged
			}
			switch d {
			case DialectPaged, DialectLegacy, DialectMixed, DialectNone,
				DialectLoad, DialectWindow, DialectRef:
				dialect = d
			default:
				return "", nil, nil, fmt.Errorf("ikuaiapi: unknown query dialect %q", clause)
			}
		}
	}
	return dialect, required, optional, nil
}

func parseQueryFields(raw string, required bool) ([]QueryField, error) {
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	fields := make([]QueryField, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return nil, fmt.Errorf("ikuaiapi: empty field name in query clause")
		}
		numeric := strings.HasSuffix(p, "#")
		p = strings.TrimSuffix(p, "#")
		if p == "" {
			return nil, fmt.Errorf("ikuaiapi: numeric marker without a field name")
		}
		fields = append(fields, QueryField{Name: p, Required: required, Numeric: numeric})
	}
	return fields, nil
}

// QueryField describes one query parameter an endpoint accepts on top of
// its dialect's defaults.
type QueryField struct {
	// Name is the key exactly as the router expects it, including case
	// ("ip", "mac", "KEYWORDS", "FINDS", "TYPE").
	Name string
	// Required marks parameters the router rejects the call without.
	Required bool
	// Numeric marks parameters that must be rendered as an integer.
	Numeric bool
}

type V4Endpoint struct {
	Group   string
	Name    string
	Path    string
	Methods []string
	// Load marks monitoring load-style endpoints. Such endpoints accept
	// datetype/start_time/end_time/math query params rather than the
	// usual page/limit/filter/order/order_by. The codegen emits a
	// typed <Name>LoadOptions struct plus enum validation for these.
	//
	// Deprecated: prefer Query: "load". Load is still honoured and is
	// treated as DialectLoad when Query is empty.
	Load bool
	// Action is the verb suffix for action-style endpoints (those whose
	// path ends with ":start", ":stop", ":restart", ":sync", ":restore",
	// ":check"). The codegen uses this to emit semantically named
	// helpers (Start<Name>, Stop<Name>, Restore<Name>, etc.) instead of
	// a generic Do<Name>. Empty means the endpoint follows the standard
	// CRUD shape.
	Action string
	// Query is the compact per-endpoint query contract; see
	// ParseQuerySpec. Empty means the standard paged dialect, which is
	// what the generic List helpers already emit. Endpoints with a
	// non-default contract additionally get a typed
	// Query<Name>(ctx, *<Name>QueryOptions) helper.
	//
	//	Query: "legacy+opt:status,device"
	//	Query: "paged+req:ip,mac+opt:limit#"
	//	Query: "window+req:appids+opt:starttime#,stoptime#"
	Query string
}

// QuerySpec decodes the endpoint's Query contract into its dialect and
// its extra required / optional parameters. An endpoint that still uses
// the deprecated Load flag decodes as DialectLoad.
func (e V4Endpoint) QuerySpec() (QueryDialect, []QueryField, []QueryField, error) {
	if e.Query == "" && e.Load {
		return DialectLoad, nil, nil, nil
	}
	return ParseQuerySpec(e.Query)
}

// V4EndpointCatalog is the verified inventory of the v4 API.
//
// Paths here were checked against iKuai OS 4.x hardware, and several
// entries exist because a probe proved the obvious spelling wrong:
//
//   - VPN server configuration lives at /vpn/<proto>/services. The bare
//     /vpn/<proto> answers 404.
//   - The IPv6 and protocol object groups are /ip6-objects and
//     /proto-objects. /ipv6-objects and /protocol-objects answer 404.
//   - The page-size query parameter is "limit". "page_size" is accepted
//     and ignored: the router returns every row, so paging silently does
//     nothing.
//   - Every collection uses order=<direction> and order_by=<field>.
//     Swapping the two returns no rows at all, on security collections
//     included.
//   - Four AC entries used to sit at the API root (/services, /clients,
//     /peers, /ref). All four 404. The controller's real endpoints are
//     /network/ac/services and /network/ac/ap-config, which the network
//     group already carried.
//
// A 404 here means the path is wrong, not that the feature is absent: the
// router answers an unknown path with an HTML 404 body rather than the
// usual JSON envelope.
var V4EndpointCatalog = []V4Endpoint{
	{Group: "advanced", Name: "ftp-config", Path: "/advanced-service/ftp-config", Methods: []string{"GET", "PUT"}},
	{Group: "advanced", Name: "ftp-users", Path: "/advanced-service/ftp-users", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "advanced", Name: "http-users", Path: "/advanced-service/http-users", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "advanced", Name: "samba-config", Path: "/advanced-service/samba-config", Methods: []string{"GET", "PUT"}},
	{Group: "advanced", Name: "samba-users", Path: "/advanced-service/samba-users", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "advanced", Name: "snmpd-config", Path: "/advanced-service/snmpd-config", Methods: []string{"GET", "PUT"}},
	{Group: "auth", Name: "online-users", Path: "/auth/online-users", Methods: []string{"GET", "DELETE"}, Query: "mixed+opt:KEYWORDS,FINDS"},
	{Group: "auth", Name: "packages", Path: "/auth/packages", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "auth", Name: "users", Path: "/auth/users", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "auth", Name: "web-services", Path: "/auth/web/services", Methods: []string{"GET", "PUT"}},
	{Group: "interfaces", Name: "lan-config", Path: "/interfaces/lan-config", Methods: []string{"GET"}},
	{Group: "interfaces", Name: "physical", Path: "/interfaces/physical", Methods: []string{"GET"}},
	{Group: "interfaces", Name: "wan-config", Path: "/interfaces/wan-config", Methods: []string{"GET"}},
	{Group: "interfaces", Name: "wan-vlan-config", Path: "/interfaces/wan-vlan-config", Methods: []string{"GET"}},
	{Group: "log", Name: "arp", Path: "/log/arp", Methods: []string{"GET", "DELETE"}},
	{Group: "log", Name: "auth", Path: "/log/auth", Methods: []string{"GET", "DELETE"}},
	{Group: "log", Name: "ddns", Path: "/log/ddns", Methods: []string{"GET", "DELETE"}},
	{Group: "log", Name: "dhcp", Path: "/log/dhcp", Methods: []string{"GET", "DELETE"}},
	{Group: "log", Name: "notice", Path: "/log/notice", Methods: []string{"GET", "DELETE"}},
	{Group: "log", Name: "pppoe", Path: "/log/pppoe", Methods: []string{"GET", "DELETE"}},
	{Group: "log", Name: "system", Path: "/log/system", Methods: []string{"GET", "DELETE"}},
	{Group: "log", Name: "web-activity", Path: "/log/web_activity", Methods: []string{"GET", "DELETE"}},
	{Group: "log", Name: "url-visits", Path: "/log/url-visits", Methods: []string{"GET", "DELETE"}, Query: "paged+opt:pattern,starttime#,stoptime#"},
	{Group: "log", Name: "wireless", Path: "/log/wireless", Methods: []string{"GET", "DELETE"}},
	{Group: "monitoring", Name: "app-protocols-history-load", Path: "/monitoring/app-protocols/history-load", Methods: []string{"GET"}, Query: "window+req:appids+opt:starttime#,stoptime#"},
	{Group: "monitoring", Name: "app-protocols-load", Path: "/monitoring/app-protocols/load", Methods: []string{"GET"}},
	{Group: "monitoring", Name: "app-protocols-terminal-load", Path: "/monitoring/app-protocols/terminal-load", Methods: []string{"GET"}, Query: "none+req:appid#"},
	{Group: "monitoring", Name: "app-traffic-summary", Path: "/monitoring/app-traffic-summary", Methods: []string{"GET"}},
	{Group: "monitoring", Name: "cameras", Path: "/monitoring/cameras", Methods: []string{"GET"}, Query: "legacy+opt:status,KEYWORDS"},
	{Group: "monitoring", Name: "channel-clients", Path: "/monitoring/channel-clients", Methods: []string{"GET"}, Query: "none+opt:channel#"},
	{Group: "monitoring", Name: "clients-app-protocols-load", Path: "/monitoring/clients/app-protocols/load", Methods: []string{"GET"}, Query: "paged+req:ip,mac+opt:limit#"},
	{Group: "monitoring", Name: "clients-ip6-offline", Path: "/monitoring/clients-ip6-offline", Methods: []string{"GET"}, Query: "paged+opt:key,pattern"},
	{Group: "monitoring", Name: "clients-ip6-online", Path: "/monitoring/clients-ip6-online", Methods: []string{"GET"}, Query: "paged+opt:key,pattern"},
	{Group: "monitoring", Name: "clients-offline", Path: "/monitoring/clients-offline", Methods: []string{"GET"}, Query: "paged+opt:key,pattern"},
	{Group: "monitoring", Name: "clients-online", Path: "/monitoring/clients-online", Methods: []string{"GET"}, Query: "paged+opt:key,pattern"},
	{Group: "monitoring", Name: "clients-protocols", Path: "/monitoring/clients/protocols", Methods: []string{"GET"}, Query: "window+req:ip,mac+opt:starttime#,stoptime#"},
	{Group: "monitoring", Name: "clients-protocols-history-load", Path: "/monitoring/clients/protocols/history-load", Methods: []string{"GET"}, Query: "window+req:ip,mac+opt:starttime#,stoptime#"},
	{Group: "monitoring", Name: "clients-traffic-load", Path: "/monitoring/clients-traffic-load", Methods: []string{"GET"}, Query: "none+req:ip,mac"},
	{Group: "monitoring", Name: "clients-traffic-summary", Path: "/monitoring/clients-traffic-summary", Methods: []string{"GET"}},
	{Group: "monitoring", Name: "connections", Path: "/monitoring/connections", Methods: []string{"GET"}, Load: true},
	{Group: "monitoring", Name: "cpu", Path: "/monitoring/cpu", Methods: []string{"GET"}, Load: true},
	{Group: "monitoring", Name: "cputemp", Path: "/monitoring/cputemp", Methods: []string{"GET"}, Load: true},
	{Group: "monitoring", Name: "disk", Path: "/monitoring/disk", Methods: []string{"GET"}, Load: true},
	{Group: "monitoring", Name: "downstream", Path: "/monitoring/downstream", Methods: []string{"GET"}, Query: "legacy+opt:status,device"},
	{Group: "monitoring", Name: "flow-shunting", Path: "/monitoring/flow-shunting", Methods: []string{"GET"}, Query: "none+req:TYPE"},
	{Group: "monitoring", Name: "interfaces-config", Path: "/monitoring/interfaces-config", Methods: []string{"GET"}},
	{Group: "monitoring", Name: "interfaces-physical", Path: "/monitoring/interfaces-physical", Methods: []string{"GET"}},
	{Group: "monitoring", Name: "interfaces-status", Path: "/monitoring/interfaces-status", Methods: []string{"GET"}},
	{Group: "monitoring", Name: "interfaces-traffic", Path: "/monitoring/interfaces-traffic", Methods: []string{"GET"}},
	{Group: "monitoring", Name: "interfaces-traffic-v6", Path: "/monitoring/interfaces-traffic-v6", Methods: []string{"GET"}},
	{Group: "monitoring", Name: "memory", Path: "/monitoring/memory", Methods: []string{"GET"}, Load: true},
	{Group: "monitoring", Name: "network", Path: "/monitoring/network", Methods: []string{"GET"}, Load: true},
	{Group: "monitoring", Name: "protocols", Path: "/monitoring/protocols", Methods: []string{"GET"}, Query: "window+opt:starttime#,stoptime#"},
	{Group: "monitoring", Name: "protocols-history-load", Path: "/monitoring/protocols/history-load", Methods: []string{"GET"}, Query: "window+opt:starttime#,stoptime#"},
	{Group: "monitoring", Name: "ssid-clients", Path: "/monitoring/ssid-clients", Methods: []string{"GET"}, Query: "none+opt:ssid"},
	{Group: "monitoring", Name: "switch", Path: "/monitoring/switch", Methods: []string{"GET"}, Query: "legacy+opt:status,device"},
	{Group: "monitoring", Name: "system", Path: "/monitoring/system", Methods: []string{"GET"}},
	{Group: "monitoring", Name: "terminals", Path: "/monitoring/terminals", Methods: []string{"GET"}, Load: true},
	{Group: "monitoring", Name: "wireless-score", Path: "/monitoring/wireless-score", Methods: []string{"GET"}},
	{Group: "monitoring", Name: "wireless-statistics", Path: "/monitoring/wireless-statistics", Methods: []string{"GET"}},
	{Group: "monitoring", Name: "wireless-traffic", Path: "/monitoring/wireless-traffic", Methods: []string{"GET"}, Query: "none+opt:apmac"},
	{Group: "network", Name: "ac-ap-config", Path: "/network/ac/ap-config", Methods: []string{"GET", "PUT", "DELETE"}},
	{Group: "network", Name: "ac-services", Path: "/network/ac/services", Methods: []string{"GET", "PUT"}},
	{Group: "network", Name: "ac-services-start", Path: "/network/ac/services:start", Methods: []string{"POST"}, Action: "start"},
	{Group: "network", Name: "ac-services-stop", Path: "/network/ac/services:stop", Methods: []string{"POST"}, Action: "stop"},
	{Group: "network", Name: "dhcp-access-control-mode", Path: "/network/dhcp/access-control/mode", Methods: []string{"GET", "PUT"}},
	{Group: "network", Name: "dhcp-access-control-rules", Path: "/network/dhcp/access-control/rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "network", Name: "dhcp-clients", Path: "/network/dhcp/clients", Methods: []string{"GET"}},
	{Group: "network", Name: "dhcp-services", Path: "/network/dhcp/services", Methods: []string{"GET", "POST", "PUT", "PATCH"}},
	{Group: "network", Name: "dhcp-services-restart", Path: "/network/dhcp/services:restart", Methods: []string{"POST"}, Action: "restart"},
	{Group: "network", Name: "dhcp-services-start", Path: "/network/dhcp/services:start", Methods: []string{"POST"}, Action: "start"},
	{Group: "network", Name: "dhcp-services-stop", Path: "/network/dhcp/services:stop", Methods: []string{"POST"}, Action: "stop"},
	{Group: "network", Name: "dhcp-static", Path: "/network/dhcp/static", Methods: []string{"GET", "POST", "PUT", "PATCH"}},
	{Group: "network", Name: "dhcp6-access-control-mode", Path: "/network/dhcp6/access-control/mode", Methods: []string{"GET", "PUT"}},
	{Group: "network", Name: "dhcp6-access-control-rules", Path: "/network/dhcp6/access-control/rules", Methods: []string{"GET", "POST", "PUT", "PATCH"}},
	{Group: "network", Name: "dhcp6-clients", Path: "/network/dhcp6/clients", Methods: []string{"GET"}},
	{Group: "network", Name: "dmz-rules", Path: "/network/dmz/rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "network", Name: "dnat-rules", Path: "/network/dnat/rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "network", Name: "dns-config", Path: "/network/dns/config", Methods: []string{"GET", "PUT"}},
	{Group: "network", Name: "dns-proxy-rules", Path: "/network/dns/proxy/rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "network", Name: "dns-stats", Path: "/network/dns/stats", Methods: []string{"GET"}},
	{Group: "network", Name: "nat-rules", Path: "/network/nat/rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "network", Name: "pppoe-services", Path: "/network/pppoe/services", Methods: []string{"GET", "PUT"}},
	{Group: "network", Name: "qos-ip", Path: "/network/qos/ip", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "network", Name: "qos-mac", Path: "/network/qos/mac", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "network", Name: "vlan", Path: "/network/vlan", Methods: []string{"GET", "POST", "PUT", "PATCH"}},
	{Group: "objects", Name: "domain-objects", Path: "/domain-objects", Methods: []string{"GET", "POST", "PUT", "DELETE"}},
	{Group: "objects", Name: "ip-objects", Path: "/ip-objects", Methods: []string{"GET", "POST", "PUT", "DELETE"}},
	{Group: "objects", Name: "ip6-objects", Path: "/ip6-objects", Methods: []string{"GET", "POST", "PUT", "DELETE"}},
	{Group: "objects", Name: "mac-objects", Path: "/mac-objects", Methods: []string{"GET", "POST", "PUT", "DELETE"}},
	{Group: "objects", Name: "port-objects", Path: "/port-objects", Methods: []string{"GET", "POST", "PUT", "DELETE"}},
	{Group: "objects", Name: "proto-objects", Path: "/proto-objects", Methods: []string{"GET", "POST", "PUT", "DELETE"}},
	{Group: "objects", Name: "time-objects", Path: "/time-objects", Methods: []string{"GET", "POST", "PUT", "DELETE"}},

	// Each object group answers "which rules reference this group" on a
	// "ref" sub-path with a group_name query parameter, wrapping the
	// result in an "object_ref" object.
	{Group: "objects", Name: "ip-objects-ref", Path: "/ip-objects/ref", Methods: []string{"GET"}, Query: "ref+req:group_name"},
	{Group: "objects", Name: "ip6-objects-ref", Path: "/ip6-objects/ref", Methods: []string{"GET"}, Query: "ref+req:group_name"},
	{Group: "objects", Name: "mac-objects-ref", Path: "/mac-objects/ref", Methods: []string{"GET"}, Query: "ref+req:group_name"},
	{Group: "objects", Name: "port-objects-ref", Path: "/port-objects/ref", Methods: []string{"GET"}, Query: "ref+req:group_name"},
	{Group: "objects", Name: "proto-objects-ref", Path: "/proto-objects/ref", Methods: []string{"GET"}, Query: "ref+req:group_name"},
	{Group: "objects", Name: "domain-objects-ref", Path: "/domain-objects/ref", Methods: []string{"GET"}, Query: "ref+req:group_name"},
	{Group: "objects", Name: "time-objects-ref", Path: "/time-objects/ref", Methods: []string{"GET"}, Query: "ref+req:group_name"},
	{Group: "system", Name: "backup", Path: "/system/backup", Methods: []string{"GET", "POST", "DELETE"}},
	{Group: "system", Name: "backup-auto", Path: "/system/backup-auto", Methods: []string{"GET", "POST", "PUT", "DELETE"}},
	{Group: "system", Name: "backup-restore", Path: "/system/backup:restore", Methods: []string{"POST"}, Action: "restore"},
	{Group: "system", Name: "disks", Path: "/system/disks", Methods: []string{"GET"}},
	{Group: "system", Name: "files", Path: "/system/files", Methods: []string{"GET"}},
	{Group: "system", Name: "upgrade", Path: "/system/upgrade", Methods: []string{"GET", "POST"}},
	{Group: "system", Name: "upgrade-check", Path: "/system/upgrade:check", Methods: []string{"POST"}, Action: "check"},
	{Group: "system", Name: "upgrade-start", Path: "/system/upgrade:start", Methods: []string{"POST"}, Action: "start"},
	{Group: "system", Name: "upgrade-status", Path: "/system/upgrade:status", Methods: []string{"GET"}},
	{Group: "system", Name: "web-admin-accounts", Path: "/system/web-admin/accounts", Methods: []string{"GET", "POST", "PUT", "DELETE"}},
	{Group: "system", Name: "web-admin-groups", Path: "/system/web-admin/groups", Methods: []string{"GET", "POST", "PUT", "DELETE"}},
	{Group: "system", Name: "web-admin-password", Path: "/system/web-admin/password", Methods: []string{"PUT"}},
	{Group: "system", Name: "web-admin-password-status", Path: "/system/web-admin/password-status", Methods: []string{"GET"}},
	{Group: "routing", Name: "app-protocols", Path: "/routing/app-protocols", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "routing", Name: "domain-rules", Path: "/routing/domain-rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "routing", Name: "five-tuple-rules", Path: "/routing/five-tuple-rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "routing", Name: "load-balance-rules", Path: "/routing/load-balance-rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "routing", Name: "static-routes", Path: "/routing/static-routes", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "routing", Name: "updown", Path: "/routing/updown", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "security", Name: "acl-rules", Path: "/security/acl-rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "security", Name: "advanced-config", Path: "/security/advanced/config", Methods: []string{"GET", "PUT"}},
	{Group: "security", Name: "app-protocols-professional-rules", Path: "/security/app-protocols/professional/rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "security", Name: "domain-blacklist-rules", Path: "/security/domain-blacklist/rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "security", Name: "mac-mode", Path: "/security/mac-mode", Methods: []string{"GET", "PUT"}},
	{Group: "security", Name: "mac-rules", Path: "/security/mac-rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "security", Name: "peerconn-rules", Path: "/security/peerconn/rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "security", Name: "secondary-route-config", Path: "/security/secondary-route/config", Methods: []string{"GET", "PUT"}},
	{Group: "security", Name: "terminals", Path: "/security/terminals", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "security", Name: "url-black-rules", Path: "/security/url-black/rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "security", Name: "url-keywords-rules", Path: "/security/url-keywords/rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "security", Name: "url-redirect-rules", Path: "/security/url-redirect/rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "security", Name: "url-replace-rules", Path: "/security/url-replace/rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "system", Name: "alg", Path: "/system/alg", Methods: []string{"GET", "PUT"}},
	{Group: "system", Name: "basic-config", Path: "/system/basic/config", Methods: []string{"GET", "PUT"}},
	{Group: "system", Name: "basic-ntp-sync", Path: "/system/basic/ntp:sync", Methods: []string{"POST"}, Action: "sync"},
	{Group: "system", Name: "cpufreq", Path: "/system/cpufreq", Methods: []string{"GET", "PUT"}},
	{Group: "system", Name: "cpufreq-mode", Path: "/system/cpufreq/mode", Methods: []string{"PUT"}},
	{Group: "system", Name: "kernel-params", Path: "/system/kernel-params", Methods: []string{"GET", "PUT"}},
	{Group: "system", Name: "reboot-schedules", Path: "/system/reboot-schedules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "system", Name: "remote-access", Path: "/system/remote-access", Methods: []string{"GET", "PUT"}},
	{Group: "system", Name: "vrrp-config", Path: "/system/vrrp/config", Methods: []string{"GET", "PUT"}},
	{Group: "system", Name: "vrrp-start", Path: "/system/vrrp:start", Methods: []string{"POST"}, Action: "start"},
	{Group: "system", Name: "vrrp-stop", Path: "/system/vrrp:stop", Methods: []string{"POST"}, Action: "stop"},
	{Group: "vpn", Name: "ikev2", Path: "/vpn/ikev2/services", Methods: []string{"GET", "PUT"}},
	{Group: "vpn", Name: "ikev2-clients", Path: "/vpn/ikev2/clients", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "vpn", Name: "ipsec-clients", Path: "/vpn/ipsec/clients", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "vpn", Name: "l2tp", Path: "/vpn/l2tp/services", Methods: []string{"GET", "PUT"}},
	{Group: "vpn", Name: "l2tp-clients", Path: "/vpn/l2tp/clients", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "vpn", Name: "openvpn", Path: "/vpn/openvpn/services", Methods: []string{"GET", "PUT"}},
	{Group: "vpn", Name: "openvpn-clients", Path: "/vpn/openvpn/clients", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "vpn", Name: "pptp", Path: "/vpn/pptp/services", Methods: []string{"GET", "PUT"}},
	{Group: "vpn", Name: "pptp-clients", Path: "/vpn/pptp/clients", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "vpn", Name: "wireguard", Path: "/vpn/wireguard", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "vpn", Name: "wireguard-peers", Path: "/vpn/wireguard/{tunnel_id}/peers", Methods: []string{"GET", "POST"}, Query: "paged+opt:key,pattern"},
	{Group: "vpn", Name: "wireguard-peer", Path: "/vpn/wireguard/{tunnel_id}/peers/{peer_id}", Methods: []string{"GET", "PUT", "PATCH", "DELETE"}},
	{Group: "wireless", Name: "access-control-rules", Path: "/wireless/access-control/rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
	{Group: "wireless", Name: "vlan-rules", Path: "/wireless/vlan/rules", Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"}},
}

func V4EndpointByName(name string) (V4Endpoint, bool) {
	for _, endpoint := range V4EndpointCatalog {
		if endpoint.Name == name {
			return endpoint, true
		}
	}
	return V4Endpoint{}, false
}

// V4EndpointByGroupName resolves an endpoint by its (group, name) pair.
// Use this when the same Name appears under multiple Groups (e.g.
// "system" exists in both "log" and "monitoring").
func V4EndpointByGroupName(group, name string) (V4Endpoint, bool) {
	for _, endpoint := range V4EndpointCatalog {
		if endpoint.Group == group && endpoint.Name == name {
			return endpoint, true
		}
	}
	return V4Endpoint{}, false
}
