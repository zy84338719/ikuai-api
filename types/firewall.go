package types

type ACLShowResponse struct {
	BaseResponse
	Data struct {
		Data []ACLItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []ACLItem `json:"data"`
	} `json:"results,omitempty"`
}

type ACLItem struct {
	ID       int    `json:"id"`
	TagName  string `json:"tagname"`
	Enabled  string `json:"enabled"`
	Comment  string `json:"comment"`
	SrcAddr  string `json:"src_addr"`
	DstAddr  string `json:"dst_addr"`
	SrcPort  string `json:"src_port"`
	DstPort  string `json:"dst_port"`
	Protocol string `json:"protocol"`
	Action   string `json:"action"`
	Week     string `json:"week"`
	Time     string `json:"time"`
}

func (r *ACLShowResponse) GetData() []ACLItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type DNATShowResponse struct {
	BaseResponse
	Data struct {
		Data []DNATItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []DNATItem `json:"data"`
	} `json:"results,omitempty"`
}

type DNATItem struct {
	ID         int                    `json:"id"`
	TagName    string                 `json:"tagname"`
	Enabled    string                 `json:"enabled"`
	Interface  string                 `json:"interface"`
	Protocol   string                 `json:"protocol"`
	WanPort    string                 `json:"wan_port"`
	LanAddr    string                 `json:"lan_addr"`
	LanPort    string                 `json:"lan_port"`
	Comment    string                 `json:"comment"`
	LanAddrInt int                    `json:"lan_addr_int"`
	SrcAddr    map[string]interface{} `json:"src_addr"`
}

func (r *DNATShowResponse) GetData() []DNATItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type ConnLimitShowResponse struct {
	BaseResponse
	Data struct {
		Data []ConnLimitItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []ConnLimitItem `json:"data"`
	} `json:"results,omitempty"`
}

type ConnLimitItem struct {
	ID      int    `json:"id"`
	TagName string `json:"tagname"`
	Enabled string `json:"enabled"`
	Comment string `json:"comment"`
	SrcAddr string `json:"src_addr"`
	ConnNum int    `json:"conn_num"`
	NewConn int    `json:"new_conn"`
}

func (r *ConnLimitShowResponse) GetData() []ConnLimitItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type DomainGroupShowResponse struct {
	BaseResponse
	Data struct {
		Data []DomainGroupItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []DomainGroupItem `json:"data"`
	} `json:"results,omitempty"`
}

type DomainGroupItem struct {
	ID      int    `json:"id"`
	TagName string `json:"tagname"`
	Comment string `json:"comment"`
	Domain  string `json:"domain"`
}

func (r *DomainGroupShowResponse) GetData() []DomainGroupItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type CustomISPShowResponse struct {
	BaseResponse
	Data struct {
		Data []CustomISPItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []CustomISPItem `json:"data"`
	} `json:"results,omitempty"`
}

type CustomISPItem struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	IPGroup string `json:"ipgroup"`
	Comment string `json:"comment"`
	Time    string `json:"time"`
}

func (r *CustomISPShowResponse) GetData() []CustomISPItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type StreamDomainShowResponse struct {
	BaseResponse
	Data struct {
		Data []StreamDomainItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []StreamDomainItem `json:"data"`
	} `json:"results,omitempty"`
}

type StreamDomainItem struct {
	ID        int    `json:"id"`
	Interface string `json:"interface"`
	SrcAddr   string `json:"src_addr"`
	Enabled   string `json:"enabled"`
	Week      string `json:"week"`
	Comment   string `json:"comment"`
	Domain    string `json:"domain"`
	Time      string `json:"time"`
}

func (r *StreamDomainShowResponse) GetData() []StreamDomainItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}
