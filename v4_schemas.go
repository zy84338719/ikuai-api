package ikuaiapi

import (
	"fmt"
	"sort"
	"strings"
)

// This file records the per-endpoint body contract: which fields a create
// must carry by default, and which fields an update is allowed to touch.
//
// It exists because the firmware validates writes strictly and, for most
// rule-like resources, rejects a partial update rather than filling in the
// gaps with its own defaults — which silently reconfigures whatever the
// caller left out. Without this table the caller has to know each
// endpoint's field set by trial and error against a live router.
//
// Every allowlist below was cross-checked against live responses, field by
// field, on real hardware. Two outcomes came out of that sweep:
//
//   - Fields the router returns and accepts on write but the allowlist
//     omitted were added. Without them a read-modify-write silently
//     cleared the setting: open_ftp on remote-access, termname on DHCP
//     static, the limit_* pair on the advanced security config, auth on
//     the OpenVPN server, and so on.
//   - Fields the router computes for display stay out of the allowlist,
//     because sending them back is rejected or meaningless:
//
//     /network/dhcp/services — available, status
//     /system/cpufreq        — coreid, cpuid, freq, hardirq, phyid, softirq, used
//     /system/cpufreq/mode   — cpufreq_list, cpufreq_support,
//                              current_cpufreq, current_turbo, turbo_support
//
// Their absence from an update body is deliberate, not an oversight.
//
// The _int companions the router attaches to several fields are dropped by
// DropSuffix rather than by the allowlist, because they appear on fields
// that are themselves writable.
//
// Every value below was taken from behaviour verified against real
// hardware. Where the firmware is genuinely inconsistent the inconsistency
// is reproduced and called out rather than normalized away — see the
// auth-accounts note.

// V4Schema is the body contract for one endpoint path.
type V4Schema struct {
	// CreateDefaults are the fields a create must carry when the caller
	// supplies nothing. Caller-supplied values always win.
	CreateDefaults map[string]any
	// UpdateDefaults are filled in only on an update, not on a create.
	//
	// Some resources genuinely need a different default set per verb: the
	// DHCP service endpoint writes no opt_type* fields on create, but an
	// update must send twelve of them. A single CreateDefaults map cannot
	// express that, which is why the two sets are separate.
	UpdateDefaults map[string]any
	// InputFields is the allowlist an update may send. Nil means the
	// endpoint imposes no allowlist and every field is accepted.
	InputFields []string
	// RequiredFields are the fields a create must set explicitly; the SDK
	// cannot invent them. Empty when every required field already has a
	// meaningful default.
	RequiredFields []string
	// Mirror describes fields the firmware keeps in sync with one another.
	Mirror []MirrorRule
	// DropBlanks lists fields that must not be forwarded as an empty string
	// when they come back from the router. The firmware reads an empty
	// option value as "no option configured" rather than "unchanged", so
	// the field is removed from the update body and the endpoint's default
	// fills it instead. A value the caller supplies is left alone — the
	// caller means it.
	DropBlanks []string
	// DropSuffix lists field-name suffixes the router derives for its own
	// use and rejects on write. The routing endpoints return a "_int"
	// companion for several fields; sending one back fails the update.
	DropSuffix []string
	// NormalizeSelectors rewrites an empty selector pair from a JSON object
	// into an empty array before the body goes out. Some firmware reads
	// back {"custom":{},"object":{}} where it wrote arrays, and rejects the
	// object form on the next write.
	NormalizeSelectors bool
}

// MirrorRule records that one body field is copied onto another after a
// read-modify-write update.
type MirrorRule struct {
	// From is the field to read.
	From string
	// To is the field to write.
	To string
	// Overwrite distinguishes the two firmware behaviours. When true the
	// copy is unconditional — it replaces a value the caller just set,
	// which is what the VLAN endpoint requires, since its display name and
	// its rule name are the same field in the firmware's eyes. When false
	// the copy only fills an empty To.
	Overwrite bool
}

// HasDefaults reports whether the endpoint carries create defaults.
func (s V4Schema) HasDefaults() bool { return len(s.CreateDefaults) > 0 }

// HasUpdateDefaults reports whether the endpoint fills extra fields on an
// update that a create does not carry.
func (s V4Schema) HasUpdateDefaults() bool { return len(s.UpdateDefaults) > 0 }

// HasInputFields reports whether the endpoint constrains update fields.
func (s V4Schema) HasInputFields() bool { return len(s.InputFields) > 0 }

// AllowsField reports whether an update to this endpoint may carry field.
// An endpoint without an allowlist accepts everything.
func (s V4Schema) AllowsField(field string) bool {
	if len(s.InputFields) == 0 {
		return true
	}
	for _, f := range s.InputFields {
		if f == field {
			return true
		}
	}
	return false
}

// V4Schemas maps an endpoint path to its body contract.
var V4Schemas = map[string]V4Schema{
	// --- security -------------------------------------------------------

	"/security/acl-rules": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "ip_type": "4", "ctdir": 0,
			"dir": "forward", "iinterface": "any", "ointerface": "any",
			"prio": 50, "comment": "",
			"src_addr": selectorPair(), "dst_addr": selectorPair(),
			"src_port": selectorPair(), "dst_port": selectorPair(),
			"src_addr_inv": 0, "dst_addr_inv": 0, "src_type": 0, "dst_type": 0,
			"src_area_code": "", "dst_area_code": "",
			"src6_addr": selectorPair(), "dst6_addr": selectorPair(),
			"src6_mode": 0, "dst6_mode": 0,
			"src6_suffix": "", "dst6_suffix": "",
			"time": weeklyWindow(),
		},
		RequiredFields: []string{"tagname", "action", "protocol"},
	},
	"/security/app-protocols/professional/rules": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "",
			"app_proto": selectorPair(),
			"src_addr":  "", "dst_addr": "",
			"time": weeklyWindow(),
		},
		RequiredFields: []string{"tagname", "action", "prio", "app_proto"},
	},
	"/security/mac-rules": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "expires": 0, "comment": "",
			"time": weeklyWindow(),
		},
		RequiredFields: []string{"tagname", "mac"},
	},
	"/security/url-black/rules": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "",
			"domain":   selectorPair(),
			"src_addr": selectorPair(),
			// This endpoint is the one place where "time" is an empty
			// selector pair rather than the all-day weekly window every
			// other rule gets.
			"time": selectorPair(),
		},
		RequiredFields: []string{"tagname", "mode", "domain"},
	},
	"/security/url-keywords/rules": {
		CreateDefaults: map[string]any{"enabled": "yes"},
		RequiredFields: []string{"tagname", "mode", "src_url", "ori_keyword", "rep_keyword", "hit_rate", "prio"},
	},
	"/security/url-redirect/rules": {
		CreateDefaults: map[string]any{"enabled": "yes"},
		RequiredFields: []string{"tagname", "mode", "src_url", "dst_url", "hit_rate", "prio"},
	},
	"/security/url-replace/rules": {
		CreateDefaults: map[string]any{"enabled": "yes"},
		RequiredFields: []string{"tagname", "mode", "src_url", "param_keyword", "rep_keyword", "hit_rate", "prio"},
	},
	"/security/domain-blacklist/rules": {
		CreateDefaults: map[string]any{"enabled": "yes"},
		RequiredFields: []string{"tagname", "domain_group"},
	},
	"/security/peerconn/rules": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "",
			"src_addr": selectorPair(),
			"dst_port": selectorPair(),
		},
		RequiredFields: []string{"tagname", "limits", "protocol"},
	},
	"/security/terminals": {
		// No "enabled": the firmware rejects it here, and this endpoint has
		// no toggle verb either.
		RequiredFields: []string{"tagname", "mac"},
	},

	// --- routing --------------------------------------------------------

	// The routing endpoints return derived "_int" companions for several
	// fields. The firmware rejects a body carrying them, so they are dropped
	// on the way back in.
	"/routing/static-routes": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "ip_type": "4", "comment": "", "prio": 1,
		},
		DropSuffix:         []string{"_int"},
		NormalizeSelectors: true,
		RequiredFields: []string{
			"tagname", "dst_addr", "gateway", "netmask", "interface",
		},
	},

	// --- singleton configuration endpoints -------------------------------
	//
	// These have no collection and no id: the path addresses the one
	// document, so the update is a merge rather than a row edit.

	"/network/dns/config": {
		// A DNS change must carry the whole document: the firmware applies
		// its own defaults to anything absent, which can silently reset
		// caching, defence or proxy settings the operator never touched.
		InputFields: []string{
			"enabled", "forbid_dns_4a", "cache_ttl", "cachemode",
			"proxy_force", "proxy_force_dns", "dns1", "dns2", "query",
			"defense", "network", "query_args_ip", "query_head_ip",
		},
	},
	"/network/dhcp/access-control/mode": {
		InputFields: []string{"mode"},
	},
	"/network/dhcp6/access-control/mode": {
		InputFields: []string{"mode"},
	},
	"/security/mac-mode": {
		InputFields: []string{"acl_mac"},
	},
	"/security/advanced/config": {
		InputFields: []string{
			"noping_lan", "noping_wan", "notracert", "hijack_ping",
			"invalid", "dos_lan", "dos_lan_num", "tcp_mss", "tcp_mss_num",
			"limit_tcp2p", "limit_tcp2p_num", "limit_udp2p", "limit_udp2p_num",
		},
	},
	"/security/secondary-route/config": {
		InputFields:        []string{"nol2rt", "nol2rt_ip", "ttl_num", "time"},
		NormalizeSelectors: true,
	},
	"/system/cpufreq": {
		InputFields: []string{"mode", "turbo"},
	},

	"/routing/domain-rules": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "", "prio": 31,
			"domain":   selectorPair(),
			"src_addr": selectorPair(),
			"time":     weeklyWindow(),
		},
		RequiredFields: []string{"tagname", "interface"},
	},
	"/routing/five-tuple-rules": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "",
			"type": 0, "mode": 0, "prio": 31, "iface_band": 0,
			"src_addr_inv": 0, "dst_addr_inv": 0, "protocol": "any",
			"src_addr": selectorPair(), "dst_addr": selectorPair(),
			"src_port": selectorPair(), "dst_port": selectorPair(),
			"time": weeklyWindow(),
		},
		RequiredFields: []string{"tagname", "interface"},
	},
	"/routing/app-protocols": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "",
			"app_proto": selectorPair(),
			"mode":      0, "prio": 31, "iface_band": 0,
			"src_addr": selectorPair(),
			"time":     weeklyWindow(),
		},
		RequiredFields: []string{"tagname", "interface"},
	},
	"/routing/load-balance-rules": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "",
			"mode": 0,
			// weight is a string here, unlike most numeric fields.
			"weight":   "1",
			"isp_name": "all",
		},
		RequiredFields: []string{"tagname", "interface"},
	},
	"/routing/updown": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "", "protocol": "any",
			"src_addr": selectorPair(), "dst_addr": selectorPair(),
			"src_port": selectorPair(), "dst_port": selectorPair(),
			// No time field on this resource.
		},
		RequiredFields: []string{"tagname", "upiface", "downiface"},
	},

	// --- QoS ------------------------------------------------------------

	"/network/qos/ip": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "", "type": 0, "protocol": "any",
			"ip_addr":  selectorPair(),
			"src_port": selectorPair(), "dst_port": selectorPair(),
			"time": weeklyWindow(),
		},
		RequiredFields: []string{"tagname", "interface", "upload", "download"},
	},
	"/network/qos/mac": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "", "type": 0, "ip_type": "4",
			"time": weeklyWindow(),
		},
		RequiredFields: []string{"tagname", "interface", "upload", "download"},
	},

	// --- network --------------------------------------------------------

	"/network/nat/rules": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "", "protocol": "",
			"nat_addr": "", "nat_port": "",
			"src_addr": selectorPair(), "dst_addr": selectorPair(),
			"src_port": selectorPair(), "dst_port": selectorPair(),
			"src_addr_inv": 0, "dst_addr_inv": 0,
		},
		InputFields: []string{
			"tagname", "enabled", "action", "iinterface", "ointerface",
			"src_addr", "dst_addr", "nat_addr", "nat_port", "protocol",
			"comment", "src_port", "dst_port", "src_addr_inv", "dst_addr_inv",
		},
		RequiredFields: []string{"tagname", "action", "iinterface", "ointerface"},
	},
	"/network/dnat/rules": {
		CreateDefaults: map[string]any{
			"enabled":  "yes",
			"comment":  "",
			"src_addr": selectorPair(),
		},
		InputFields: []string{
			"id", "tagname", "enabled", "lan_addr", "lan_port", "protocol",
			"interface", "wan_port", "src_addr", "comment",
		},
		RequiredFields: []string{"tagname", "lan_addr", "lan_port", "wan_port", "protocol", "interface"},
		// The router returns lan_addr_int alongside lan_addr; it is
		// derived for display and is rejected on write.
		DropSuffix: []string{"_int"},
	},
	"/network/dmz/rules": {
		CreateDefaults: map[string]any{"enabled": "yes", "comment": ""},
		InputFields: []string{
			"id", "tagname", "enabled", "interface", "lan_addr", "protocol", "excl_port", "comment",
		},
		RequiredFields: []string{"tagname", "interface", "lan_addr", "protocol", "excl_port"},
		// The router returns lan_addr_int alongside lan_addr; it is
		// derived for display and is rejected on write.
		DropSuffix: []string{"_int"},
	},
	"/network/dns/proxy/rules": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "",
			// Note the plain string, not a selector pair.
			"src_addr": "",
		},
		RequiredFields: []string{"domain", "dns_addr", "parse_type"},
	},
	"/network/vlan": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "mac": "", "ip_mask": "", "comment": "",
		},
		// The firmware treats the VLAN's name and its rule name as the same
		// value, and copies one onto the other unconditionally — including
		// over a value the caller just set.
		Mirror: []MirrorRule{
			{From: "vlan_name", To: "tagname", Overwrite: true},
		},
		InputFields: []string{
			"vlan_id", "vlan_name", "tagname", "interface", "mac",
			"ip_addr", "netmask", "ip_mask", "enabled", "comment",
		},
		RequiredFields: []string{"vlan_id", "vlan_name", "interface", "ip_addr", "netmask"},
	},
	"/network/pppoe/services": {
		CreateDefaults: map[string]any{
			"server_name": "iKuai", "force_verify_name": 0,
			"nas_identifier": "iKuai", "nas_ip_address": "",
			"authport": 1812, "accountport": 1813,
			"rate_limit_lan": 1, "drop_client": 1,
			"force_pppoe": 0, "enhance_check": 1, "share_deny": 0,
			"bind_vlan": 0, "verify_vlan": 1, "bind_iface": 0,
			"mtu": 1480, "mru": 1480,
			"lcp_echo_interval": 10, "lcp_echo_failure": 3,
			"maxconnect": 0, "restart_timer": 0, "comment": "",
		},
		InputFields: []string{
			"enabled", "server_name", "force_verify_name", "server_ip",
			"dns1", "dns2", "authmode", "nas_identifier", "nas_ip_address",
			"radius_ip", "secret", "authport", "accountport", "addr_pool",
			"interface", "rate_limit_lan", "drop_client", "force_pppoe",
			"enhance_check", "share_deny", "bind_vlan", "verify_vlan",
			"bind_iface", "mtu", "mru", "lcp_echo_interval",
			"lcp_echo_failure", "maxconnect", "restart_timer",
			"restart_week", "restart_time", "comment",
		},
	},
	"/network/dhcp6/access-control/rules": {
		InputFields:    []string{"enabled", "mac", "tagname", "comment"},
		RequiredFields: []string{"tagname", "mac"},
	},
	// The IPv4 and IPv6 access-control rule sets take the same body; only
	// the path differs.
	"/network/dhcp/access-control/rules": {
		CreateDefaults: map[string]any{"enabled": "yes", "comment": ""},
		InputFields:    []string{"enabled", "mac", "tagname", "comment"},
		RequiredFields: []string{"tagname", "mac"},
	},

	// --- DHCP -----------------------------------------------------------

	// The DHCP service endpoint is the clearest case of a contract that
	// differs per verb: a create writes no opt_type* fields at all, while an
	// update must send twelve of them. Eleven default to 0 and opt_type121
	// defaults to 2. Sending a create body to the update path (or the
	// reverse) reconfigures DHCP options the caller never touched.
	"/network/dhcp/services": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "delay": 0,
			"check_addr_valid": 0, "check_relay_only": 0,
		},
		UpdateDefaults: map[string]any{
			"opt_type15": 0, "opt_type28": 0, "opt_type43": 0,
			"opt_type60": 0, "opt_type66": 0, "opt_type67": 0,
			"opt_type80": 0, "opt_type119": 0, "opt_type125": 0,
			"opt_type128": 0, "opt_type138": 0,
			// The only non-zero option default.
			"opt_type121": 2,
		},
		InputFields: []string{
			"enabled", "tagname", "interface", "phy_ifnames", "addr_pool",
			"exclude_pool", "netmask", "gateway", "dns1", "dns2", "wins1",
			"wins2", "domain", "next_server", "lease", "delay",
			"opt_type15", "opt15", "opt_type28", "opt28", "opt_type43",
			"opt43", "opt_type60", "opt60", "opt_type66", "opt66",
			"opt_type67", "opt67", "opt_type80", "opt80", "opt_type119",
			"opt119", "opt_type125", "opt125", "opt_type128", "opt128",
			"opt_type138", "opt138", "opt_type121", "opt121",
			"check_addr_valid", "check_relay_only",
		},
		RequiredFields: []string{
			"tagname", "interface", "phy_ifnames", "addr_pool",
			"netmask", "gateway", "lease",
		},
		// An option value that came back empty must be dropped rather than
		// re-sent: the router treats "" as "no option", so forwarding it
		// would silently clear the option the default is about to set.
		DropBlanks: []string{
			"opt_type15", "opt_type28", "opt_type43", "opt_type60",
			"opt_type66", "opt_type67", "opt_type80", "opt_type119",
			"opt_type125", "opt_type128", "opt_type138", "opt_type121",
		},
	},
	"/network/dhcp/static": {
		CreateDefaults: map[string]any{"enabled": "yes"},
		InputFields: []string{
			"enabled", "mac", "ip_addr", "interface", "gateway", "dns1",
			"dns2", "comment", "tagname", "hostname", "termname",
		},
		// The list returns ip_addr_int beside ip_addr; it is derived for
		// display and is rejected on write.
		DropSuffix:     []string{"_int"},
		RequiredFields: []string{"mac", "ip_addr", "interface", "tagname"},
	},

	// --- object groups --------------------------------------------------

	// Object groups carry their members in group_value, built with
	// ObjectGroup (or TimeObjectGroup for the time group). The display name
	// is group_name, not "name" — the one field every object endpoint
	// spells differently from what a caller would guess.
	//
	// The writable surface was confirmed with a create/read/update/delete
	// cycle on hardware: a row comes back as exactly
	// {group_name, group_value, id, ref_count}, so ref_count is a computed
	// reference counter and id identifies the row. Neither belongs in an
	// update body.
	"/ip-objects": {
		InputFields:    []string{"group_name", "group_value"},
		RequiredFields: []string{"group_name"},
	},
	"/ip6-objects": {
		InputFields:    []string{"group_name", "group_value"},
		RequiredFields: []string{"group_name"},
	},
	"/mac-objects": {
		InputFields:    []string{"group_name", "group_value"},
		RequiredFields: []string{"group_name"},
	},
	"/port-objects": {
		InputFields:    []string{"group_name", "group_value"},
		RequiredFields: []string{"group_name"},
	},
	"/proto-objects": {
		InputFields:    []string{"group_name", "group_value"},
		RequiredFields: []string{"group_name"},
	},
	"/domain-objects": {
		InputFields:    []string{"group_name", "group_value"},
		RequiredFields: []string{"group_name"},
	},
	"/time-objects": {
		InputFields:    []string{"group_name", "group_value"},
		RequiredFields: []string{"group_name"},
	},

	// --- wireless -------------------------------------------------------

	"/wireless/access-control/rules": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "mode": 0,
			"lmac": selectorPair(),
			// This resource uses a flat window plus a weekday string
			// instead of the nested weekly structure elsewhere.
			"lssid": "ALL", "lap": "ALL",
			"week": "1234567", "time": "00:00-23:59",
			"comment": "",
		},
		InputFields: []string{
			"enabled", "tagname", "mode", "lmac", "lssid", "lap", "week", "time", "comment",
		},
		RequiredFields: []string{"tagname"},
	},
	"/wireless/vlan/rules": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "lssid": "ALL",
			"lmac":    selectorPair(),
			"comment": "",
		},
		InputFields:    []string{"enabled", "tagname", "vlanid", "lmac", "lssid", "comment"},
		RequiredFields: []string{"tagname", "vlanid"},
	},

	// --- VPN ------------------------------------------------------------

	"/vpn/pptp/clients": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "",
			"server_port": 1723, "mtu": 1400, "mru": 1400,
			"check_link_mode": 3, "check_link_host": "www.baidu.com",
			"timing_rst_switch": 0, "timing_rst_week": "1234567",
			"timing_rst_time": "12:00", "cycle_rst_time": 0,
		},
		InputFields: []string{
			"id", "enabled", "name", "comment", "server", "server_port",
			"username", "passwd", "interface", "mtu", "mru", "check_link_mode",
			"check_link_host", "timing_rst_switch", "timing_rst_week",
			"timing_rst_time", "cycle_rst_time",
		},
		RequiredFields: []string{"name", "server", "username", "passwd", "interface"},
	},
	"/vpn/l2tp/clients": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "",
			"server_port": 1701, "mtu": 1400, "mru": 1400,
			"check_link_mode": 3, "check_link_host": "www.baidu.com",
			"timing_rst_switch": 0, "cycle_rst_time": 0,
		},
		InputFields: []string{
			"id", "enabled", "name", "comment", "server", "server_port",
			"username", "passwd", "ipsec_secret", "interface", "leftid",
			"rightid", "mtu", "mru", "check_link_mode", "check_link_host",
			"timing_rst_switch", "timing_rst_week", "timing_rst_time",
			"cycle_rst_time",
		},
		RequiredFields: []string{"name", "server", "username", "passwd", "interface"},
	},
	"/vpn/openvpn/clients": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "",
			// This endpoint mixes types: the port, MTU and compression
			// fields are strings while accept_push_route is a number.
			"remote_port": "1194", "method": 0,
			"proto": "udp", "dev_type": "tun", "cipher": "BF-CBC",
			"tun_mtu": "1400", "comp_lzo": "1",
			"accept_push_route": 1,
			"check_link_mode":   3, "check_link_host": "www.baidu.com",
			"timing_rst_switch": 0, "ca": "",
		},
		InputFields: []string{
			"id", "enabled", "name", "comment", "remote_addr", "remote_port",
			"method", "username", "password", "interface", "proto", "dev_type",
			"cipher", "tls_auth", "ca", "cert", "key", "accept_push_route",
			"route", "comp_lzo", "tun_mtu", "check_link_mode", "check_link_host",
			"timing_rst_switch", "timing_rst_week", "timing_rst_time", "extra_config",
		},
		RequiredFields: []string{"name", "remote_addr", "interface", "username", "password", "ca"},
	},
	"/vpn/ikev2/clients": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "",
			"authby":          "mschapv2",
			"check_link_mode": 3, "check_link_host": "www.baidu.com",
		},
		InputFields: []string{
			"id", "enabled", "name", "comment", "remote_addr", "interface",
			"authby", "secret", "leftid", "rightid", "username", "passwd",
			"check_link_mode", "check_link_host",
		},
		RequiredFields: []string{"name", "remote_addr", "interface", "leftid"},
	},
	"/vpn/ipsec/clients": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "",
			"keyexchange": "ikev2", "authby": "secret", "dpdaction": "none",
			"compress": "0", "ikelifetime": 3, "lifetime": 1,
			"ike_enc": "aes256", "ike_auth": "sha256", "ike_dh": "modp2048",
			"esp_enc": "aes256", "esp_auth": "sha256", "aggressive": "0",
		},
		InputFields: []string{
			"id", "name", "comment", "remote_addr", "authby", "leftsubnet",
			"rightsubnet", "interface", "enabled", "keyexchange", "aggressive",
			"ikelifetime", "ike_enc", "ike_auth", "ike_dh", "secret", "leftid",
			"rightid", "privatekey", "leftcert", "rightcert", "lifetime",
			"esp_enc", "esp_auth", "dpdaction", "dpddelay", "dpdtimeout", "compress",
		},
		RequiredFields: []string{"name", "interface", "leftsubnet", "rightsubnet"},
	},
	"/vpn/wireguard": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "interface": "auto",
			"local_listenport": 5000, "mtu": 1420,
		},
		InputFields: []string{
			"enabled", "name", "interface", "local_privatekey",
			"local_publickey", "local_address", "local_listenport", "mtu",
		},
		RequiredFields: []string{"name", "local_address", "local_privatekey", "local_publickey"},
	},
	// VPN server configuration. The path carries a /services suffix, which
	// the catalog originally omitted: verified on hardware, /vpn/pptp
	// answers 404 while /vpn/pptp/services returns the configuration.
	// WireGuard peers live under a tunnel, so their contract is keyed on the
	// templated path exactly as the catalog declares it.
	"/vpn/wireguard/{tunnel_id}/peers": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "", "keepalive": 0,
		},
		InputFields: []string{
			"enabled", "comment", "interface", "peer_publickey",
			"presharedkey", "allowips", "endpoint", "endpoint_port", "keepalive",
		},
		RequiredFields: []string{"peer_publickey", "allowips", "interface"},
	},

	"/vpn/pptp/services": {
		InputFields: []string{
			"enabled", "dns1", "dns2", "addr_pool", "open_mppe", "server_ip",
			"server_port", "mtu", "mru",
		},
	},
	"/vpn/l2tp/services": {
		InputFields: []string{
			"enabled", "server_ip", "server_port", "addr_pool", "dns1", "dns2",
			"mtu", "mru", "ipsec_secret", "leftid", "rightid", "force_ipsec",
		},
	},
	"/vpn/openvpn/services": {
		InputFields: []string{
			"enabled", "proto", "port", "subnet", "mask", "tun_mtu", "cipher",
			"comp_lzo", "dev_type", "topology", "method", "tls_auth", "ca",
			"cert", "key", "push_gateway", "push_route", "push_route_comment",
			"push_dns", "extra_config", "auth",
		},
	},
	"/vpn/ikev2/services": {
		InputFields: []string{
			"id", "enabled", "authby", "addrpool", "secret", "leftid", "rightid",
			"dns1", "dns2", "share_deny", "mtu", "privatekey", "leftcert",
			"aggressive", "keyexchange", "name",
		},
	},

	// --- auth -----------------------------------------------------------

	"/auth/users": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "comment": "",
			"ppptype":     "any",
			"packages":    0,
			"upload":      0,
			"download":    0,
			"start_time":  0,
			"expires":     0,
			"share":       1,
			"ip_type":     0,
			"auto_mac":    0,
			"auto_vlanid": 0,
			"bind_vlanid": "0",
			"bind_ifname": "any",
			// This endpoint is the one exception to the selector pair: its
			// src_addr holds a map and carries no "object" key. Verified
			// against hardware; do not normalize it.
			"src_addr": map[string]any{"custom": map[string]any{}},
		},
		InputFields: []string{
			"username", "passwd", "enabled", "ppptype", "packages", "upload",
			"download", "start_time", "expires", "share", "ip_type", "auto_mac",
			"auto_vlanid", "bind_vlanid", "pppname", "pppoev6_wan",
			"bind_ifname", "src_addr", "mac", "address", "name", "phone",
			"cardid", "comment",
		},
		RequiredFields: []string{"username", "passwd"},
	},
	"/auth/packages": {
		CreateDefaults: map[string]any{"comment": ""},
		InputFields:    []string{"packname", "packtime", "price", "up_speed", "down_speed", "comment"},
		RequiredFields: []string{"packname", "packtime", "price", "up_speed", "down_speed"},
	},

	// --- advanced -------------------------------------------------------

	"/advanced-service/ftp-users": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "upload": 0, "download": 0,
		},
		InputFields: []string{
			"enabled", "username", "passwd", "permission", "home_dir",
			"upload", "download", "tagname",
		},
		RequiredFields: []string{"username", "passwd", "permission", "home_dir"},
	},
	"/advanced-service/http-users": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "server_name": "", "access": 1,
		},
		InputFields: []string{
			"enabled", "tagname", "http_port", "server_name", "ssl_on",
			"autoindex", "download", "home_dir", "access",
		},
		RequiredFields: []string{"tagname", "http_port", "ssl_on", "autoindex", "download", "home_dir"},
	},
	"/advanced-service/samba-users": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "browseable": "",
		},
		InputFields: []string{
			"enabled", "username", "passwd", "name", "perm", "guest",
			"home_dir", "browseable", "tagname",
		},
		RequiredFields: []string{"name", "username", "passwd", "permission", "guest", "home_dir"},
	},
	"/advanced-service/ftp-config": {
		InputFields: []string{"open_ftp", "ftp_port", "ftp_access"},
	},
	"/advanced-service/samba-config": {
		InputFields: []string{"enabled", "workgroup", "wsdd2", "access", "interface"},
	},
	"/advanced-service/snmpd-config": {
		InputFields: []string{
			"enabled", "listen_port", "syslocation", "syscontact", "sysname",
			"version", "community", "source", "rw", "username", "security",
			"auth_proto", "auth_pass", "priv_proto", "priv_pass",
		},
	},

	// --- system ---------------------------------------------------------

	"/system/basic/config": {
		InputFields: []string{
			"hostname", "language", "time_zone", "time_zone_full", "switch_nat",
			"switch_ntp", "switch_ntpd", "switch_ntpserver", "ntpserver_list",
			"ntp_sync_cycle", "link_mode", "lan_nat", "backport", "listenport",
			"fast_nat",
		},
	},
	"/system/reboot-schedules": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "event": "reboot", "comment": "",
		},
		InputFields: []string{
			"enabled", "tagname", "event", "strategy", "cycle_time", "time", "comment",
		},
		RequiredFields: []string{"tagname", "time", "strategy", "cycle_time"},
	},
	"/system/remote-access": {
		InputFields: []string{
			"open_telnetd", "open_wanweb", "open_sshd", "sshd_port",
			"sshd_passwd", "http_port", "https_port", "force_https",
			"open_ftp", "ftp_port", "ftp_access",
		},
	},
	"/system/vrrp/config": {
		CreateDefaults: map[string]any{
			"method": 2, "domain": "www.baidu.com",
			"dns": "114.114.114.114", "interval": 3,
		},
		InputFields: []string{
			"enabled", "type", "prio", "method", "domain", "dns", "gateway",
			"interval", "ifnames", "auto_sync", "single_line", "ignore_wanstatus",
			"interfaces", "virtual_ips", "ht_iface", "remote_addr",
		},
	},
	"/system/alg": {
		InputFields: []string{
			"support_ftp", "support_tftp", "support_sip", "support_h323",
			"ftp_ports", "sip_ports", "tftp_ports",
		},
	},
	"/system/kernel-params": {
		InputFields: []string{
			"bbr", "syn_recv_timeout", "syn_send_timeout", "established_timeout",
			"fin_wait_timeout", "last_ack_timeout", "close_wait_timeout",
			"time_wait_timeout", "close_timeout", "udp_timeout",
			"udp_stream_timeout", "icmp_timeout",
		},
	},
	"/system/cpufreq/mode": {
		InputFields: []string{"mode", "turbo"},
	},
	"/system/backup-auto": {
		InputFields: []string{"id", "enabled", "strategy", "time", "cycle_time", "valid_days", "tagname"},
	},
	"/system/web-admin/groups": {
		InputFields: []string{"group_name", "ip_addr", "perm_config"},
		// The display name is what the caller sets; the firmware reads
		// group_name. See Skills/writing.md.
		RequiredFields: []string{"ip_addr", "perm_config"},
	},
	"/system/web-admin/accounts": {
		CreateDefaults: map[string]any{
			"enabled": "yes", "force": 0, "interval": 30,
			"sesstimeout": 120, "comment": "",
		},
		InputFields: []string{
			"username", "passwd", "enabled", "group_id", "force", "interval",
			"sesstimeout", "comment",
		},
		RequiredFields: []string{"username", "passwd", "group_id"},
	},
}

// selectorPair builds the empty {"custom": [...], "object": [...]} payload
// the firmware expects for an unset address, port, protocol or domain
// selector.
func selectorPair() map[string]any {
	return map[string]any{"custom": []any{}, "object": []any{}}
}

// weeklyWindow builds the all-day, every-week schedule the firmware applies
// to a rule when the caller supplies no window.
func weeklyWindow() map[string]any {
	return map[string]any{
		"custom": []any{map[string]any{
			"type":       "weekly",
			"weekdays":   "1234567",
			"start_time": "00:00",
			"end_time":   "23:59",
			"comment":    "",
		}},
		"object": []any{},
	}
}

// Schema returns the body contract for an endpoint path.
func Schema(path string) (V4Schema, bool) {
	s, ok := V4Schemas[path]
	return s, ok
}

// SchemaPaths lists every endpoint path that carries a body contract,
// sorted, so a caller can enumerate the coverage.
func SchemaPaths() []string {
	out := make([]string, 0, len(V4Schemas))
	for p := range V4Schemas {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// CreateBody builds a create payload for an endpoint: every field in the
// endpoint's CreateDefaults, with the caller's overrides applied on top,
// then the result normalized for the integer fields.
//
// The returned body is a deep copy, so mutating it — or mutating it again
// on a later call — cannot corrupt the shared contract.
//
// A field listed in the endpoint's RequiredFields that the caller neither
// supplied nor got a default for is reported as an error, because sending
// an incomplete create fails in a way that looks like a network problem.
//
// Endpoints with no recorded contract fall through to a plain merge, which
// keeps callers working against a router that is ahead of this table.
func CreateBody(path string, overrides map[string]any) (map[string]any, error) {
	body := map[string]any{}
	schema, known := V4Schemas[path]
	if known && len(schema.CreateDefaults) > 0 {
		body = deepCopyMap(schema.CreateDefaults)
	}
	for k, v := range overrides {
		body[k] = deepCopyValue(v)
	}

	if known {
		var missing []string
		for _, f := range schema.RequiredFields {
			if !bodyHasValue(body, f) {
				missing = append(missing, f)
			}
		}
		if len(missing) > 0 {
			return nil, NewError(
				fmt.Sprintf("%s create is missing required field(s): %s",
					path, strings.Join(missing, ", ")),
				WithCode(errCodeValidation),
				WithCause(ErrRequiredField),
			)
		}
	}
	return CoerceBody(body), nil
}

// UpdateFields returns the update allowlist for an endpoint, or nil when
// the endpoint imposes none. It lets a caller inspect the contract before
// deciding how to build a payload.
func UpdateFields(path string) []string {
	if s, ok := V4Schemas[path]; ok {
		return s.InputFields
	}
	return nil
}

// bodyHasValue reports whether a field counts as supplied. An empty string
// is treated as absent: the firmware rejects the write either way, and a
// clearer error here beats a confusing 3001 from the router.
func bodyHasValue(body map[string]any, field string) bool {
	v, ok := body[field]
	if !ok || v == nil {
		return false
	}
	if s, ok := v.(string); ok && strings.TrimSpace(s) == "" {
		return false
	}
	return true
}

// deepCopyValue copies the nested maps and slices a default body is made
// of, so callers never share structure with V4Schemas.
func deepCopyValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		return deepCopyMap(t)
	case []any:
		out := make([]any, len(t))
		for i, item := range t {
			out[i] = deepCopyValue(item)
		}
		return out
	default:
		return v
	}
}

func deepCopyMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = deepCopyValue(v)
	}
	return out
}
