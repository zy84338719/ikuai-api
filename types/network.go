package types

type WanShowResponse struct {
	BaseResponse
	Data struct {
		Data []WanItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []WanItem `json:"data"`
	} `json:"results,omitempty"`
}

type WanItem struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	TagName          string `json:"tagname"`
	Comment          string `json:"comment"`
	BandEth          string `json:"bandeth"`
	BandIf           string `json:"bandif"`
	Internet         int    `json:"internet"`
	Policy           int    `json:"policy"`
	MTU              int    `json:"mtu"`
	MRU              int    `json:"mru"`
	PPPoEStatus      int    `json:"pppoe_status"`
	PPPoEIPAddr      string `json:"pppoe_ip_addr"`
	PPPoENetmask     string `json:"pppoe_netmask"`
	PPPoEGateway     string `json:"pppoe_gateway"`
	PPPoEDNS1        string `json:"pppoe_dns1"`
	PPPoEDNS2        string `json:"pppoe_dns2"`
	PPPoEMacRemote   string `json:"pppoe_macremote"`
	PPPoEUpdatetime  int64  `json:"pppoe_updatetime"`
	DHCPStatus       int    `json:"dhcp_status"`
	DHCPIPAddr       string `json:"dhcp_ip_addr"`
	DHCPNetmask      string `json:"dhcp_netmask"`
	DHCPGateway      string `json:"dhcp_gateway"`
	DHCPDNS1         string `json:"dhcp_dns1"`
	DHCPDNS2         string `json:"dhcp_dns2"`
	DHCPLease        int    `json:"dhcp_lease"`
	DHCPUpdatetime   int64  `json:"dhcp_updatetime"`
	Username         string `json:"username"`
	Passwd           string `json:"passwd"`
	Upload           int    `json:"upload"`
	Download         int    `json:"download"`
	QoSSwitch        int    `json:"qos_switch"`
	QoSUpload        int    `json:"qos_upload"`
	QoSDownload      int    `json:"qos_download"`
	DefaultRoute     int    `json:"default_route"`
	DiscAutoSwitch   int    `json:"disc_auto_switch"`
	CheckLinkMode    int    `json:"check_link_mode"`
	CheckLinkHost    string `json:"check_link_host"`
	LinkTime         string `json:"link_time"`
	LinkMode         int    `json:"linkmode"`
	EnableIPv6       int    `json:"enable_ipv6"`
	VLANInternetInfo string `json:"vlan_internet_info"`
	ModifiedTime     int64  `json:"modified_time"`
}

func (r *WanShowResponse) GetData() []WanItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type LanShowResponse struct {
	BaseResponse
	Data struct {
		Data []LanItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []LanItem `json:"data"`
	} `json:"results,omitempty"`
}

type LanItem struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	TagName    string `json:"tagname"`
	Comment    string `json:"comment"`
	IPMask     string `json:"ip_mask"`
	BandEth    string `json:"bandeth"`
	BandIf     string `json:"bandif"`
	DHCPServer int    `json:"dhcp_server"`
	LanVisit   int    `json:"lan_visit"`
	Policy     int    `json:"policy"`
	Vlan       string `json:"vlan"`
}

func (r *LanShowResponse) GetData() []LanItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type VLANShowResponse struct {
	BaseResponse
	Data struct {
		Data []VLANItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []VLANItem `json:"data"`
	} `json:"results,omitempty"`
}

type VLANItem struct {
	ID        int    `json:"id"`
	TagName   string `json:"tagname"`
	Interface string `json:"interface"`
	VlanID    int    `json:"vlan_id"`
	Comment   string `json:"comment"`
}

func (r *VLANShowResponse) GetData() []VLANItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type IPv6ShowResponse struct {
	BaseResponse
	Data struct {
		Data []IPv6Item `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []IPv6Item `json:"data"`
	} `json:"results,omitempty"`
}

type IPv6Item struct {
	ID             int    `json:"id"`
	TagName        string `json:"tagname"`
	Interface      string `json:"interface"`
	Enabled        string `json:"enabled"`
	Internet       string `json:"internet"`
	Prefix         string `json:"prefix"`
	PrefixHint     string `json:"prefix_hint"`
	IPv6Addr       string `json:"ipv6_addr"`
	IPv6Gateway    string `json:"ipv6_gateway"`
	LinkAddr       string `json:"link_addr"`
	DHCP6IPAddr    string `json:"dhcp6_ip_addr"`
	DHCP6IPGateway string `json:"dhcp6_ip_gateway"`
	DHCP6Prefix1   string `json:"dhcp6_prefix1"`
	DHCP6DNS1      string `json:"dhcp6_dns1"`
	DHCP6DNS2      string `json:"dhcp6_dns2"`
	PreferredLft   string `json:"preferred_lft"`
	ValidLft       string `json:"valid_lft"`
	ForceGenDuid   int    `json:"force_gen_duid"`
	ForcePrefix    int    `json:"force_prefix"`
}

func (r *IPv6ShowResponse) GetData() []IPv6Item {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type IPTVShowResponse struct {
	BaseResponse
	Data struct {
		Data []IPTVItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []IPTVItem `json:"data"`
	} `json:"results,omitempty"`
}

type IPTVItem struct {
	ID        int    `json:"id"`
	Enabled   string `json:"enabled"`
	Mode      int    `json:"mode"`
	WanIface  string `json:"wan_iface"`
	WanVlanID int    `json:"wan_vlanid"`
	LanIface  string `json:"lan_iface"`
	LanVlanID int    `json:"lan_vlanid"`
}

func (r *IPTVShowResponse) GetData() []IPTVItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type DDNSShowResponse struct {
	BaseResponse
	Data struct {
		Data []DDNSItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []DDNSItem `json:"data"`
	} `json:"results,omitempty"`
}

type DDNSItem struct {
	ID        int    `json:"id"`
	TagName   string `json:"tagname"`
	Enabled   string `json:"enabled"`
	Interface string `json:"interface"`
	Server    string `json:"server"`
	Domain    string `json:"domain"`
	Username  string `json:"username"`
	Passwd    string `json:"passwd"`
	IPAddress string `json:"ipaddress"`
	Result    string `json:"result"`
	Type      string `json:"type"`
	TopDomain string `json:"top_domain"`
	Mode      int    `json:"mode"`
	Location  int    `json:"location"`
	Line      string `json:"line"`
	CFToken   int    `json:"cf_token"`
	Account   string `json:"account"`
	Mac       string `json:"mac"`
	Duid      string `json:"duid"`
}

func (r *DDNSShowResponse) GetData() []DDNSItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}
