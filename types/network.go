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

// DNSForwardShowResponse - DNS转发响应
type DNSForwardShowResponse struct {
	BaseResponse
	Data struct {
		Data []DNSForwardItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []DNSForwardItem `json:"data"`
	} `json:"results,omitempty"`
}

func (r *DNSForwardShowResponse) GetData() []DNSForwardItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

// DNSForwardItem - DNS转发项
type DNSForwardItem struct {
	ID        int    `json:"id"`
	TagName   string `json:"tagname"`
	Enabled   string `json:"enabled"`
	Comment   string `json:"comment"`
	Interface string `json:"interface"`
	Domain    string `json:"domain"`
	ForwardIP string `json:"forward_ip"`
}

// DNSStaticShowResponse - DNS静态解析响应
type DNSStaticShowResponse struct {
	BaseResponse
	Data struct {
		Data []DNSStaticItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []DNSStaticItem `json:"data"`
	} `json:"results,omitempty"`
}

func (r *DNSStaticShowResponse) GetData() []DNSStaticItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

// DNSStaticItem - DNS静态解析项
type DNSStaticItem struct {
	ID      int    `json:"id"`
	TagName string `json:"tagname"`
	Enabled string `json:"enabled"`
	Comment string `json:"comment"`
	Domain  string `json:"domain"`
	IPAddr  string `json:"ip_addr"`
}

// RouteStaticShowResponse - 静态路由响应
type RouteStaticShowResponse struct {
	BaseResponse
	Data struct {
		Data []RouteStaticItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []RouteStaticItem `json:"data"`
	} `json:"results,omitempty"`
}

func (r *RouteStaticShowResponse) GetData() []RouteStaticItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

// RouteStaticItem - 静态路由项
type RouteStaticItem struct {
	ID        int    `json:"id"`
	TagName   string `json:"tagname"`
	Enabled   string `json:"enabled"`
	Comment   string `json:"comment"`
	DstAddr   string `json:"dst_addr"`
	Gateway   string `json:"gateway"`
	Interface string `json:"interface"`
	Metric    int    `json:"metric"`
}

// RoutePolicyShowResponse - 策略路由响应
type RoutePolicyShowResponse struct {
	BaseResponse
	Data struct {
		Data []RoutePolicyItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []RoutePolicyItem `json:"data"`
	} `json:"results,omitempty"`
}

func (r *RoutePolicyShowResponse) GetData() []RoutePolicyItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

// RoutePolicyItem - 策略路由项
type RoutePolicyItem struct {
	ID          int    `json:"id"`
	TagName     string `json:"tagname"`
	Enabled     string `json:"enabled"`
	Comment     string `json:"comment"`
	SrcAddr     string `json:"src_addr"`
	DstAddr     string `json:"dst_addr"`
	Interface   string `json:"interface"`
	Gateway     string `json:"gateway"`
	Description string `json:"description"`
}

// FlowControlShowResponse - 流控响应
type FlowControlShowResponse struct {
	BaseResponse
	Data struct {
		Data []FlowControlItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []FlowControlItem `json:"data"`
	} `json:"results,omitempty"`
}

func (r *FlowControlShowResponse) GetData() []FlowControlItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

// FlowControlItem - 流控项
type FlowControlItem struct {
	ID       int    `json:"id"`
	TagName  string `json:"tagname"`
	Enabled  string `json:"enabled"`
	Comment  string `json:"comment"`
	SrcAddr  string `json:"src_addr"`
	DstAddr  string `json:"dst_addr"`
	Protocol string `json:"protocol"`
	SrcPort  string `json:"src_port"`
	DstPort  string `json:"dst_port"`
	Upload   int64  `json:"upload"`
	Download int64  `json:"download"`
	Priority int    `json:"priority"`
}

// BandwidthShowResponse - 带宽控制响应
type BandwidthShowResponse struct {
	BaseResponse
	Data struct {
		Data []BandwidthItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []BandwidthItem `json:"data"`
	} `json:"results,omitempty"`
}

func (r *BandwidthShowResponse) GetData() []BandwidthItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

// BandwidthItem - 带宽控制项
type BandwidthItem struct {
	ID          int    `json:"id"`
	TagName     string `json:"tagname"`
	Enabled     string `json:"enabled"`
	Comment     string `json:"comment"`
	SrcAddr     string `json:"src_addr"`
	DstAddr     string `json:"dst_addr"`
	MinUpload   int64  `json:"min_upload"`
	MaxUpload   int64  `json:"max_upload"`
	MinDownload int64  `json:"min_download"`
	MaxDownload int64  `json:"max_download"`
}

// QoSShowResponse - QoS响应
type QoSShowResponse struct {
	BaseResponse
	Data struct {
		Data []QoSItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []QoSItem `json:"data"`
	} `json:"results,omitempty"`
}

func (r *QoSShowResponse) GetData() []QoSItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

// QoSItem - QoS项
type QoSItem struct {
	ID       int    `json:"id"`
	TagName  string `json:"tagname"`
	Enabled  string `json:"enabled"`
	Comment  string `json:"comment"`
	SrcAddr  string `json:"src_addr"`
	DstAddr  string `json:"dst_addr"`
	Protocol string `json:"protocol"`
	SrcPort  string `json:"src_port"`
	DstPort  string `json:"dst_port"`
	Priority int    `json:"priority"`
	DSCP     int    `json:"dscp"`
}

type DNSStaticAddRequest struct {
	TagName string `json:"tagname"`
	Enabled string `json:"enabled"`
	Comment string `json:"comment"`
	Domain  string `json:"domain"`
	IPAddr  string `json:"ip_addr"`
}

type DNSStaticEditRequest struct {
	ID      int    `json:"id"`
	TagName string `json:"tagname"`
	Enabled string `json:"enabled"`
	Comment string `json:"comment"`
	Domain  string `json:"domain"`
	IPAddr  string `json:"ip_addr"`
}

type RouteStaticAddRequest struct {
	TagName   string `json:"tagname"`
	Enabled   string `json:"enabled"`
	Comment   string `json:"comment"`
	DstAddr   string `json:"dst_addr"`
	Gateway   string `json:"gateway"`
	Interface string `json:"interface"`
	Metric    int    `json:"metric"`
}

type RouteStaticEditRequest struct {
	ID        int    `json:"id"`
	TagName   string `json:"tagname"`
	Enabled   string `json:"enabled"`
	Comment   string `json:"comment"`
	DstAddr   string `json:"dst_addr"`
	Gateway   string `json:"gateway"`
	Interface string `json:"interface"`
	Metric    int    `json:"metric"`
}

type DHCPStaticAddRequest struct {
	TagName string `json:"tagname"`
	Comment string `json:"comment"`
	Mac     string `json:"mac"`
	IPAddr  string `json:"ip_addr"`
	Enabled string `json:"enabled"`
}

type DHCPStaticEditRequest struct {
	ID      int    `json:"id"`
	TagName string `json:"tagname"`
	Comment string `json:"comment"`
	Mac     string `json:"mac"`
	IPAddr  string `json:"ip_addr"`
	Enabled string `json:"enabled"`
}
