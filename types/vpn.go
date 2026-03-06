package types

type PPTPClientShowResponse struct {
	BaseResponse
	Data struct {
		Data []PPTPClientItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []PPTPClientItem `json:"data"`
	} `json:"results,omitempty"`
}

type PPTPClientItem struct {
	ID        int    `json:"id"`
	TagName   string `json:"tagname"`
	Enabled   string `json:"enabled"`
	Comment   string `json:"comment"`
	Server    string `json:"server"`
	Username  string `json:"username"`
	Passwd    string `json:"passwd"`
	Interface string `json:"interface"`
}

func (r *PPTPClientShowResponse) GetData() []PPTPClientItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type L2TPClientShowResponse struct {
	BaseResponse
	Data struct {
		Data []L2TPClientItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []L2TPClientItem `json:"data"`
	} `json:"results,omitempty"`
}

type L2TPClientItem struct {
	ID        int    `json:"id"`
	TagName   string `json:"tagname"`
	Enabled   string `json:"enabled"`
	Comment   string `json:"comment"`
	Server    string `json:"server"`
	Username  string `json:"username"`
	Passwd    string `json:"passwd"`
	Interface string `json:"interface"`
	Secret    string `json:"secret"`
}

func (r *L2TPClientShowResponse) GetData() []L2TPClientItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type PPTPClientAddRequest struct {
	TagName   string `json:"tagname"`
	Enabled   string `json:"enabled"`
	Comment   string `json:"comment"`
	Server    string `json:"server"`
	Username  string `json:"username"`
	Passwd    string `json:"passwd"`
	Interface string `json:"interface"`
}

type PPTPClientEditRequest struct {
	ID        int    `json:"id"`
	TagName   string `json:"tagname"`
	Enabled   string `json:"enabled"`
	Comment   string `json:"comment"`
	Server    string `json:"server"`
	Username  string `json:"username"`
	Passwd    string `json:"passwd"`
	Interface string `json:"interface"`
}

type L2TPClientAddRequest struct {
	TagName   string `json:"tagname"`
	Enabled   string `json:"enabled"`
	Comment   string `json:"comment"`
	Server    string `json:"server"`
	Username  string `json:"username"`
	Passwd    string `json:"passwd"`
	Interface string `json:"interface"`
	Secret    string `json:"secret"`
}

type L2TPClientEditRequest struct {
	ID        int    `json:"id"`
	TagName   string `json:"tagname"`
	Enabled   string `json:"enabled"`
	Comment   string `json:"comment"`
	Server    string `json:"server"`
	Username  string `json:"username"`
	Passwd    string `json:"passwd"`
	Interface string `json:"interface"`
	Secret    string `json:"secret"`
}
