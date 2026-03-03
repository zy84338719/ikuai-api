package types

type DHCPDShowResponse struct {
	BaseResponse
	Data struct {
		Data []DHCPDItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []DHCPDItem `json:"data"`
	} `json:"results,omitempty"`
}

type DHCPDItem struct {
	ID        int    `json:"id"`
	Interface string `json:"interface"`
	TagName   string `json:"tagname"`
	AddrPool  string `json:"addr_pool"`
	Netmask   string `json:"netmask"`
	Gateway   string `json:"gateway"`
	Lease     int    `json:"lease"`
	Delay     int    `json:"delay"`
	DNS1      string `json:"dns1"`
	DNS2      string `json:"dns2"`
	Enabled   string `json:"enabled"`
	Comment   string `json:"comment"`
	Available int    `json:"available"`
	Status    int    `json:"status"`
}

func (r *DHCPDShowResponse) GetData() []DHCPDItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type DHCPStaticShowResponse struct {
	BaseResponse
	Data struct {
		Data []DHCPStaticItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []DHCPStaticItem `json:"data"`
	} `json:"results,omitempty"`
}

type DHCPStaticItem struct {
	ID           int    `json:"id"`
	Interface    string `json:"interface"`
	IPAddr       string `json:"ip_addr"`
	Mac          string `json:"mac"`
	Hostname     string `json:"hostname"`
	TermName     string `json:"termname"`
	StaticStatus int    `json:"static_status"`
	StartTime    int64  `json:"start_time"`
	EndTime      int64  `json:"end_time"`
	Timeout      int    `json:"timeout"`
	Status       int    `json:"status"`
}

func (r *DHCPStaticShowResponse) GetData() []DHCPStaticItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type DHCPLeaseShowResponse struct {
	BaseResponse
	Data struct {
		Data []DHCPLeaseItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []DHCPLeaseItem `json:"data"`
	} `json:"results,omitempty"`
}

type DHCPLeaseItem struct {
	ID           int    `json:"id"`
	Interface    string `json:"interface"`
	IPAddr       string `json:"ip_addr"`
	Mac          string `json:"mac"`
	Hostname     string `json:"hostname"`
	TermName     string `json:"termname"`
	StaticStatus int    `json:"static_status"`
	StartTime    int64  `json:"start_time"`
	EndTime      int64  `json:"end_time"`
	Timeout      int    `json:"timeout"`
	Status       int    `json:"status"`
}

func (r *DHCPLeaseShowResponse) GetData() []DHCPLeaseItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}
