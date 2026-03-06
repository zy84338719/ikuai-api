package types

// OnlineMonitorShowResponse - 在线用户监控响应
type OnlineMonitorShowResponse struct {
	BaseResponse
	Data struct {
		Data []OnlineMonitorItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []OnlineMonitorItem `json:"data"`
	} `json:"results,omitempty"`
}

// OnlineMonitorItem - 在线用户监控项
type OnlineMonitorItem struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	IPAddr      string `json:"ip_addr"`
	MacAddr     string `json:"mac_addr"`
	Interface   string `json:"interface"`
	LoginTime   int64  `json:"login_time"`
	LastActive  int64  `json:"last_active"`
	OnlineTime  int64  `json:"online_time"`
	Upload      int64  `json:"upload"`
	Download    int64  `json:"download"`
	ConnNum     int    `json:"conn_num"`
	DeviceType  string `json:"device_type"`
	Hostname    string `json:"hostname"`
	GroupName   string `json:"group_name"`
	GroupID     int    `json:"group_id"`
}

func (r *OnlineMonitorShowResponse) GetData() []OnlineMonitorItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}
