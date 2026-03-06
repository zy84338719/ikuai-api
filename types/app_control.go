package types

type AppControlShowResponse struct {
	BaseResponse
	Data struct {
		Data []AppControlItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []AppControlItem `json:"data"`
	} `json:"results,omitempty"`
}

type AppControlItem struct {
	ID        int    `json:"id"`
	TagName   string `json:"tagname"`
	Enabled   string `json:"enabled"`
	Comment   string `json:"comment"`
	AppName   string `json:"app_name"`
	Action    string `json:"action"`
	SrcAddr   string `json:"src_addr"`
	DstAddr   string `json:"dst_addr"`
	TimeRange string `json:"time_range"`
	Week      string `json:"week"`
}

func (r *AppControlShowResponse) GetData() []AppControlItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type AppControlAddRequest struct {
	TagName   string `json:"tagname"`
	Enabled   string `json:"enabled"`
	Comment   string `json:"comment"`
	AppName   string `json:"app_name"`
	Action    string `json:"action"`
	SrcAddr   string `json:"src_addr"`
	DstAddr   string `json:"dst_addr"`
	TimeRange string `json:"time_range"`
	Week      string `json:"week"`
}

type AppControlEditRequest struct {
	ID        int    `json:"id"`
	TagName   string `json:"tagname"`
	Enabled   string `json:"enabled"`
	Comment   string `json:"comment"`
	AppName   string `json:"app_name"`
	Action    string `json:"action"`
	SrcAddr   string `json:"src_addr"`
	DstAddr   string `json:"dst_addr"`
	TimeRange string `json:"time_range"`
	Week      string `json:"week"`
}
