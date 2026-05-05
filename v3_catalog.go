package ikuaisdk

type V3Endpoint struct {
	Group        string
	Name         string
	V4Name       string
	FuncName     string
	Actions      []string
	DefaultParam interface{}
	Supported    bool
	Notes        string
}

type V3CompatibilityStatus struct {
	V4Endpoint V4Endpoint
	V3Endpoint V3Endpoint
	Supported  bool
	Notes      string
}

func (e V3Endpoint) Supports(action string) bool {
	if !e.Supported {
		return false
	}
	for _, supported := range e.Actions {
		if supported == action {
			return true
		}
	}
	return false
}

var V3EndpointCatalog = []V3Endpoint{
	{Group: "monitoring", Name: "homepage", V4Name: "system", FuncName: "homepage", Actions: []string{"show"}, DefaultParam: map[string]string{"TYPE": "sysstat"}, Supported: true},
	{Group: "monitoring", Name: "system", V4Name: "system-load", FuncName: "monitor_system", Actions: []string{"show"}, Supported: true},
	{Group: "monitoring", Name: "clients-online", V4Name: "clients-online", FuncName: "monitor_lanip", Actions: []string{"show"}, Supported: true},
	{Group: "monitoring", Name: "clients-ip6-online", V4Name: "clients-ip6-online", FuncName: "monitor_lanipv6", Actions: []string{"show"}, Supported: true},
	{Group: "monitoring", Name: "interfaces", V4Name: "interfaces-status", FuncName: "monitor_iface", Actions: []string{"show"}, DefaultParam: map[string]string{"TYPE": "all"}, Supported: true},
	{Group: "monitoring", Name: "arp", V4Name: "arp", FuncName: "arp", Actions: []string{"show"}, Supported: true},
	{Group: "monitoring", Name: "traffic-realtime", V4Name: "clients-traffic-load", FuncName: "traffic_realtime", Actions: []string{"show"}, Supported: true},
	{Group: "monitoring", Name: "traffic-history", V4Name: "traffic-load", FuncName: "traffic_history", Actions: []string{"show"}, Supported: true},

	{Group: "network", Name: "wan", V4Name: "wan-config", FuncName: "wan", Actions: []string{"show"}, Supported: true},
	{Group: "network", Name: "lan", V4Name: "lan-config", FuncName: "lan", Actions: []string{"show"}, Supported: true},
	{Group: "network", Name: "vlan", V4Name: "vlan", FuncName: "vlan", Actions: []string{"show", "add", "edit", "del"}, Supported: true},
	{Group: "network", Name: "ipv6", V4Name: "ipv6", FuncName: "ipv6", Actions: []string{"show"}, Supported: true},
	{Group: "network", Name: "iptv", V4Name: "iptv", FuncName: "iptv", Actions: []string{"show"}, Supported: true},
	{Group: "network", Name: "ddns", V4Name: "ddns", FuncName: "ddns", Actions: []string{"show"}, Supported: true},
	{Group: "network", Name: "dhcp-services", V4Name: "dhcp-services", FuncName: "dhcpd", Actions: []string{"show"}, Supported: true},
	{Group: "network", Name: "dhcp-static", V4Name: "dhcp-static", FuncName: "dhcp_static", Actions: []string{"show", "add", "edit", "del"}, Supported: true},
	{Group: "network", Name: "dhcp-clients", V4Name: "dhcp-clients", FuncName: "dhcp_lease", Actions: []string{"show"}, Supported: true},
	{Group: "network", Name: "dns-forward", V4Name: "dns-config", FuncName: "dns_forward", Actions: []string{"show"}, Supported: true},
	{Group: "network", Name: "dns-static", V4Name: "dns-proxy-rules", FuncName: "dns_static", Actions: []string{"show", "add", "edit", "del"}, Supported: true},
	{Group: "network", Name: "pppoe-services", V4Name: "pppoe-services", FuncName: "wan", Actions: []string{"show"}, Supported: true, Notes: "v3 PPPoE fields are embedded in wan rows"},
	{Group: "network", Name: "qos", V4Name: "qos-ip", FuncName: "qos", Actions: []string{"show"}, Supported: true},
	{Group: "network", Name: "bandwidth", V4Name: "qos-mac", FuncName: "bandwidth", Actions: []string{"show"}, Supported: true},
	{Group: "network", Name: "flow-control", V4Name: "flow-control", FuncName: "flow_control", Actions: []string{"show"}, Supported: true},

	{Group: "routing", Name: "static-routes", V4Name: "static-routes", FuncName: "route_static", Actions: []string{"show", "add", "edit", "del"}, Supported: true},
	{Group: "routing", Name: "policy-routes", V4Name: "five-tuple-rules", FuncName: "route_policy", Actions: []string{"show"}, Supported: true},
	{Group: "routing", Name: "domain-rules", V4Name: "domain-rules", FuncName: "stream_domain", Actions: []string{"show"}, Supported: true},

	{Group: "security", Name: "acl-rules", V4Name: "acl-rules", FuncName: "acl", Actions: []string{"show", "add", "edit", "del"}, Supported: true},
	{Group: "security", Name: "dnat-rules", V4Name: "dnat-rules", FuncName: "dnat", Actions: []string{"show", "add", "edit", "del"}, Supported: true},
	{Group: "security", Name: "peerconn-rules", V4Name: "peerconn-rules", FuncName: "conn_limit", Actions: []string{"show", "add", "edit", "del"}, Supported: true},
	{Group: "security", Name: "domain-groups", V4Name: "domain-objects", FuncName: "domain_group", Actions: []string{"show"}, Supported: true},
	{Group: "security", Name: "custom-isp", V4Name: "ip-objects", FuncName: "custom_isp", Actions: []string{"show"}, Supported: true},
	{Group: "security", Name: "app-control", V4Name: "app-protocols-professional-rules", FuncName: "app_control", Actions: []string{"show", "add", "edit", "del"}, Supported: true},

	{Group: "vpn", Name: "pptp-clients", V4Name: "pptp-clients", FuncName: "pptp_client", Actions: []string{"show", "add", "edit", "del"}, Supported: true},
	{Group: "vpn", Name: "l2tp-clients", V4Name: "l2tp-clients", FuncName: "l2tp_client", Actions: []string{"show", "add", "edit", "del"}, Supported: true},

	{Group: "system", Name: "basic-config", V4Name: "basic-config", FuncName: "homepage", Actions: []string{"show"}, DefaultParam: map[string]string{"TYPE": "sysstat"}, Supported: true},
	{Group: "system", Name: "upgrade", V4Name: "upgrade", FuncName: "upgrade", Actions: []string{"show"}, Supported: true},
	{Group: "system", Name: "backup", V4Name: "backup", FuncName: "backup", Actions: []string{"show"}, Supported: true},
	{Group: "system", Name: "web-users", V4Name: "web-users", FuncName: "webuser", Actions: []string{"show"}, Supported: true},
	{Group: "system", Name: "upnp", V4Name: "upnp", FuncName: "upnp", Actions: []string{"show", "add", "edit", "del"}, Supported: true},

	{Group: "users", Name: "accounts", V4Name: "users", FuncName: "user_manage", Actions: []string{"show", "add", "edit", "del"}, Supported: true},
	{Group: "users", Name: "online", V4Name: "online-users", FuncName: "online_monitor", Actions: []string{"show"}, Supported: true},

	{Group: "log", Name: "notice", V4Name: "notice", FuncName: "syslog-notice", Actions: []string{"show"}, Supported: true},
	{Group: "log", Name: "pppoe", V4Name: "pppoe", FuncName: "syslog-wanpppoe", Actions: []string{"show"}, Supported: true},
	{Group: "log", Name: "dhcp", V4Name: "dhcp", FuncName: "syslog-dhcpd", Actions: []string{"show"}, Supported: true},
	{Group: "log", Name: "arp", V4Name: "arp", FuncName: "syslog-arp", Actions: []string{"show"}, Supported: true},
	{Group: "log", Name: "ddns", V4Name: "ddns", FuncName: "syslog-ddns", Actions: []string{"show"}, Supported: true},
	{Group: "log", Name: "web-activity", V4Name: "web-activity", FuncName: "syslog-webadmin", Actions: []string{"show"}, Supported: true},
	{Group: "log", Name: "system", V4Name: "system", FuncName: "syslog-sysevent", Actions: []string{"show"}, Supported: true},

	{Group: "advanced", Name: "docker-images", V4Name: "docker-images", FuncName: "docker_image", Actions: []string{"show"}, Supported: true},
	{Group: "advanced", Name: "docker-containers", V4Name: "docker-containers", FuncName: "docker_container", Actions: []string{"show"}, Supported: true},
	{Group: "advanced", Name: "docker-networks", V4Name: "docker-networks", FuncName: "docker_network", Actions: []string{"show"}, Supported: true},
	{Group: "advanced", Name: "docker-composes", V4Name: "docker-composes", FuncName: "docker_compose", Actions: []string{"show"}, Supported: true},
	{Group: "advanced", Name: "vm", V4Name: "vm", FuncName: "qemu", Actions: []string{"show", "add", "edit", "del", "start", "stop", "restart"}, Supported: true},

	{Group: "vpn", Name: "openvpn-clients", V4Name: "openvpn-clients", Supported: false, Notes: "not exposed by the current v3 Action/call catalog"},
	{Group: "vpn", Name: "ikev2-clients", V4Name: "ikev2-clients", Supported: false, Notes: "not exposed by the current v3 Action/call catalog"},
	{Group: "vpn", Name: "ipsec-clients", V4Name: "ipsec-clients", Supported: false, Notes: "not exposed by the current v3 Action/call catalog"},
	{Group: "vpn", Name: "wireguard", V4Name: "wireguard", Supported: false, Notes: "not exposed by the current v3 Action/call catalog"},
	{Group: "wireless", Name: "access-control-rules", V4Name: "access-control-rules", Supported: false, Notes: "v3 wireless ACL endpoint is not known yet"},
}

func V3EndpointsByGroup(group string) []V3Endpoint {
	group = normalizeV3Name(group)
	var endpoints []V3Endpoint
	for _, endpoint := range V3EndpointCatalog {
		if normalizeV3Name(endpoint.Group) == group {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

func V3EndpointByName(name string) (V3Endpoint, bool) {
	name = normalizeV3Name(name)
	for _, endpoint := range V3EndpointCatalog {
		if normalizeV3Name(endpoint.Name) == name || normalizeV3Name(endpoint.V4Name) == name {
			return endpoint, true
		}
	}
	return V3Endpoint{}, false
}

func V3CompatibilityForV4Catalog() []V3CompatibilityStatus {
	statuses := make([]V3CompatibilityStatus, 0, len(V4EndpointCatalog))
	for _, v4Endpoint := range V4EndpointCatalog {
		v3Endpoint, ok := V3EndpointByName(v4Endpoint.Name)
		if !ok {
			statuses = append(statuses, V3CompatibilityStatus{
				V4Endpoint: v4Endpoint,
				Supported:  false,
				Notes:      "no known v3 Action/call mapping",
			})
			continue
		}
		statuses = append(statuses, V3CompatibilityStatus{
			V4Endpoint: v4Endpoint,
			V3Endpoint: v3Endpoint,
			Supported:  v3Endpoint.Supported,
			Notes:      v3Endpoint.Notes,
		})
	}
	return statuses
}
