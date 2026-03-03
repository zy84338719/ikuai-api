package types

type SyslogShowResponse struct {
	BaseResponse
	Data struct {
		Data []SyslogItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []SyslogItem `json:"data"`
	} `json:"results,omitempty"`
}

type SyslogItem struct {
	ID        int    `json:"id"`
	Timestamp int    `json:"timestamp"`
	Content   string `json:"content"`
	Interface string `json:"interface,omitempty"`
}

func (r *SyslogShowResponse) GetData() []SyslogItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type SyslogNoticeShowResponse struct {
	BaseResponse
	Data struct {
		Data []SyslogItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []SyslogItem `json:"data"`
	} `json:"results,omitempty"`
}

func (r *SyslogNoticeShowResponse) GetData() []SyslogItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type SyslogWanPPPoEShowResponse struct {
	BaseResponse
	Data struct {
		Data []SyslogWanPPPoEItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []SyslogWanPPPoEItem `json:"data"`
	} `json:"results,omitempty"`
}

type SyslogWanPPPoEItem struct {
	ID        int    `json:"id"`
	Timestamp int    `json:"timestamp"`
	Content   string `json:"content"`
	Interface string `json:"interface"`
}

func (r *SyslogWanPPPoEShowResponse) GetData() []SyslogWanPPPoEItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type SyslogDHCPDShowResponse struct {
	BaseResponse
	Data struct {
		Data []SyslogItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []SyslogItem `json:"data"`
	} `json:"results,omitempty"`
}

func (r *SyslogDHCPDShowResponse) GetData() []SyslogItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type SyslogARPShowResponse struct {
	BaseResponse
	Data struct {
		Data []SyslogItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []SyslogItem `json:"data"`
	} `json:"results,omitempty"`
}

func (r *SyslogARPShowResponse) GetData() []SyslogItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type SyslogDDNSShowResponse struct {
	BaseResponse
	Data struct {
		Data []SyslogItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []SyslogItem `json:"data"`
	} `json:"results,omitempty"`
}

func (r *SyslogDDNSShowResponse) GetData() []SyslogItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type SyslogWebAdminShowResponse struct {
	BaseResponse
	Data struct {
		Data []SyslogItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []SyslogItem `json:"data"`
	} `json:"results,omitempty"`
}

func (r *SyslogWebAdminShowResponse) GetData() []SyslogItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type SyslogSysEventShowResponse struct {
	BaseResponse
	Data struct {
		Data []SyslogItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []SyslogItem `json:"data"`
	} `json:"results,omitempty"`
}

func (r *SyslogSysEventShowResponse) GetData() []SyslogItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}
