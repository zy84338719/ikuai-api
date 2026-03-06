package types

type UserManageShowResponse struct {
	BaseResponse
	Data struct {
		Data []UserManageItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []UserManageItem `json:"data"`
	} `json:"results,omitempty"`
}

type UserManageItem struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Group    string `json:"group"`
	Enabled  string `json:"enabled"`
	Comment  string `json:"comment"`
}

func (r *UserManageShowResponse) GetData() []UserManageItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type UserManageAddRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Group    string `json:"group"`
	Enabled  string `json:"enabled"`
	Comment  string `json:"comment"`
}

type UserManageEditRequest struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Group    string `json:"group"`
	Enabled  string `json:"enabled"`
	Comment  string `json:"comment"`
}

type OnlineMonitorShowResponse struct {
	BaseResponse
	Data struct {
		Data []OnlineMonitorItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []OnlineMonitorItem `json:"data"`
	} `json:"results,omitempty"`
}

type OnlineMonitorItem struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	IPAddr       string `json:"ip_addr"`
	MacAddr      string `json:"mac_addr"`
	OnlineTime   int64  `json:"online_time"`
	TotalTraffic int64  `json:"total_traffic"`
	SourceType   string `json:"source_type"`
}

func (r *OnlineMonitorShowResponse) GetData() []OnlineMonitorItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}
