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
