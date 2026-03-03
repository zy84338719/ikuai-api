package types

type MonitorLanIPShowResponse struct {
	BaseResponse
	Data struct {
		Data []MonitorLanIPItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []MonitorLanIPItem `json:"data"`
	} `json:"results,omitempty"`
}

type MonitorLanIPItem struct {
	ID           int         `json:"id"`
	IPAddr       string      `json:"ip_addr"`
	Mac          string      `json:"mac"`
	Hostname     string      `json:"hostname"`
	Interface    string      `json:"interface"`
	Comment      string      `json:"comment"`
	Upload       int         `json:"upload"`
	Download     int         `json:"download"`
	TotalUp      int64       `json:"total_up"`
	TotalDown    int64       `json:"total_down"`
	ConnectNum   int         `json:"connect_num"`
	ClientType   string      `json:"client_type"`
	ClientVendor string      `json:"client_vendor"`
	ClientModel  string      `json:"client_model"`
	Uptime       string      `json:"uptime"`
	PPType       string      `json:"ppptype"`
	StaticStatus int         `json:"static_status"`
	Timestamp    int         `json:"timestamp"`
	DeviceIcon   string      `json:"device_icon"`
	VendorIcon   string      `json:"vendor_icon"`
	Bssid        string      `json:"bssid"`
	Ssid         string      `json:"ssid"`
	Frequencies  string      `json:"frequencies"`
	Signal       interface{} `json:"signal"`
	ApName       string      `json:"apname"`
	ApMac        string      `json:"apmac"`
	AuthType     int         `json:"auth_type"`
	Channel      string      `json:"channel"`
	DTalkName    string      `json:"dtalk_name"`
	LinkAddr     string      `json:"link_addr"`
	Reject       int         `json:"reject"`
	UpRate       string      `json:"uprate"`
	DownRate     string      `json:"downrate"`
	WebID        int         `json:"webid"`
	Username     string      `json:"username"`
	VlanID       int         `json:"vlan_id"`
	AcGid        int         `json:"ac_gid"`
	IPAddrInt    int64       `json:"ip_addr_int"`
	TodayTotal   int         `json:"today_total"`
	Enc          string      `json:"enc"`
	TermName     string      `json:"termname"`
	UplinkAddr   string      `json:"uplink_addr"`
	UplinkDev    string      `json:"uplink_dev"`
	IPv4GNames   string      `json:"ipv4_gnames"`
	IPv6GNames   string      `json:"ipv6_gnames"`
	MacGNames    string      `json:"mac_gnames"`
	ClientTypeID int         `json:"client_typeid"`
	DeviceType   string      `json:"device_type"`
}

func (r *MonitorLanIPShowResponse) GetData() []MonitorLanIPItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type MonitorLanIPv6ShowResponse struct {
	BaseResponse
	Data struct {
		Data []MonitorLanIPv6Item `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []MonitorLanIPv6Item `json:"data"`
	} `json:"results,omitempty"`
}

type MonitorLanIPv6Item struct {
	MonitorLanIPItem
	LinkAddrV6 string `json:"link_addr"`
}

func (r *MonitorLanIPv6ShowResponse) GetData() []MonitorLanIPv6Item {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type MonitorIFaceShowResponse struct {
	BaseResponse
	Data struct {
		IFaceCheck  []IFaceCheck  `json:"iface_check"`
		IFaceStream []IFaceStream `json:"iface_stream"`
	} `json:"Data"`
	Results *struct {
		IFaceCheck  []IFaceCheck  `json:"iface_check"`
		IFaceStream []IFaceStream `json:"iface_stream"`
	} `json:"results,omitempty"`
}

type IFaceCheck struct {
	ID              int    `json:"id"`
	Interface       string `json:"interface"`
	ParentInterface string `json:"parent_interface"`
	IPAddr          string `json:"ip_addr"`
	Gateway         string `json:"gateway"`
	Internet        string `json:"internet"`
	UpdateTime      string `json:"updatetime"`
	AutoSwitch      string `json:"auto_switch"`
	Result          string `json:"result"`
	ErrMsg          string `json:"errmsg"`
	Comment         string `json:"comment"`
}

type IFaceStream struct {
	Interface   string `json:"interface"`
	Comment     string `json:"comment"`
	IPAddr      string `json:"ip_addr"`
	ConnectNum  string `json:"connect_num"`
	Upload      int    `json:"upload"`
	Download    int    `json:"download"`
	TotalUp     int64  `json:"total_up"`
	TotalDown   int64  `json:"total_down"`
	UpDropped   int    `json:"updropped"`
	DownDropped int    `json:"downdropped"`
	UpPacked    int    `json:"uppacked"`
	DownPacked  int    `json:"downpacked"`
}

func (r *MonitorIFaceShowResponse) GetIFaceCheck() []IFaceCheck {
	if r.Results != nil {
		return r.Results.IFaceCheck
	}
	return r.Data.IFaceCheck
}

func (r *MonitorIFaceShowResponse) GetIFaceStream() []IFaceStream {
	if r.Results != nil {
		return r.Results.IFaceStream
	}
	return r.Data.IFaceStream
}

type MonitorSystemShowResponse struct {
	BaseResponse
	Data struct {
		Data []MonitorSystemItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []MonitorSystemItem `json:"data"`
	} `json:"results,omitempty"`
}

type MonitorSystemItem struct {
	ID               int     `json:"id"`
	Timestamp        int     `json:"timestamp"`
	CPU              float64 `json:"cpu"`
	CPUTemp1         int     `json:"cputemp1"`
	CPUTemp2         int     `json:"cputemp2"`
	Memory           int64   `json:"memory"`
	MemoryUse        int     `json:"memory_use"`
	DiskSpaceUse     int     `json:"disk_space_use"`
	DiskSpaceUsed    int     `json:"disk_space_used"`
	ConnNum          int     `json:"conn_num"`
	OnTerminal       int     `json:"on_terminal"`
	WiredTerminal    int     `json:"wired_terminal"`
	WirelessTerminal int     `json:"wireless_terminal"`
	MaxUpload        int64   `json:"max_upload"`
	MaxDownload      int64   `json:"max_download"`
	MaxSize          int64   `json:"maxsize"`
	RxPackages       int64   `json:"rx_packages"`
	TxPackages       int64   `json:"tx_packages"`
}

func (r *MonitorSystemShowResponse) GetData() []MonitorSystemItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type ARPShowResponse struct {
	BaseResponse
	Data struct {
		Data []ARPItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []ARPItem `json:"data"`
	} `json:"results,omitempty"`
}

type ARPItem struct {
	ID        int    `json:"id"`
	Interface string `json:"interface"`
	IPAddr    string `json:"ip_addr"`
	IPAddrInt int64  `json:"ip_addr_int"`
	Mac       string `json:"mac"`
	TagName   string `json:"tagname"`
	TermName  string `json:"termname"`
	Comment   string `json:"comment"`
	BindState string `json:"bind_state"`
	BindType  string `json:"bind_type"`
}

func (r *ARPShowResponse) GetData() []ARPItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}
