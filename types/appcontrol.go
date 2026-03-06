package types

// AppControlShowResponse - 应用管控响应
type AppControlShowResponse struct {
	BaseResponse
	Data struct {
		Data []AppControlItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []AppControlItem `json:"data"`
	} `json:"results,omitempty"`
}

// AppControlItem - 应用管控项
type AppControlItem struct {
	ID          int    `json:"id"`
	TagName     string `json:"tagname"`
	Enabled     string `json:"enabled"`
	Comment     string `json:"comment"`
	SrcAddr     string `json:"src_addr"`
	DstAddr     string `json:"dst_addr"`
	Week        string `json:"week"`
	Time        string `json:"time"`
	AppGroup    string `json:"app_group"`
	AppName     string `json:"app_name"`
	Action      string `json:"action"`
	Protocol    string `json:"protocol"`
	SrcPort     string `json:"src_port"`
	DstPort     string `json:"dst_port"`
	Priority    int    `json:"priority"`
	Description string `json:"description"`
}

func (r *AppControlShowResponse) GetData() []AppControlItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

// AppControlAddRequest - 添加应用管控请求
type AppControlAddRequest struct {
	TagName  string `json:"tagname"`
	Enabled  string `json:"enabled"`
	Comment  string `json:"comment"`
	SrcAddr  string `json:"src_addr"`
	DstAddr  string `json:"dst_addr"`
	Week     string `json:"week"`
	Time     string `json:"time"`
	AppGroup string `json:"app_group"`
	AppName  string `json:"app_name"`
	Action   string `json:"action"`
	Priority int    `json:"priority"`
}

// AppControlEditRequest - 编辑应用管控请求
type AppControlEditRequest struct {
	ID       int    `json:"id"`
	TagName  string `json:"tagname"`
	Enabled  string `json:"enabled"`
	Comment  string `json:"comment"`
	SrcAddr  string `json:"src_addr"`
	DstAddr  string `json:"dst_addr"`
	Week     string `json:"week"`
	Time     string `json:"time"`
	AppGroup string `json:"app_group"`
	AppName  string `json:"app_name"`
	Action   string `json:"action"`
	Priority int    `json:"priority"`
}
