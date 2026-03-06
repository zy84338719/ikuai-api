package types

// UserManageShowResponse - 用户管理响应
type UserManageShowResponse struct {
	BaseResponse
	Data struct {
		Data []UserManageItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []UserManageItem `json:"data"`
	} `json:"results,omitempty"`
}

// UserManageItem - 用户管理项
type UserManageItem struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	Password    string `json:"passwd"`
	GroupName   string `json:"group_name"`
	GroupID     int    `json:"group_id"`
	Enabled     string `json:"enabled"`
	Comment     string `json:"comment"`
	IPAddr      string `json:"ip_addr"`
	MacAddr     string `json:"mac_addr"`
	BindMac     int    `json:"bind_mac"`
	BindIP      int    `json:"bind_ip"`
	SessTimeout int    `json:"sesstimeout"`
	MaxConn     int    `json:"max_conn"`
	UploadLimit int64  `json:"upload_limit"`
	DownloadLimit int64 `json:"download_limit"`
	ExpireTime  int64  `json:"expire_time"`
	CreateTime  int64  `json:"create_time"`
	LastLogin   int64  `json:"last_login"`
	Online      int    `json:"online"`
}

func (r *UserManageShowResponse) GetData() []UserManageItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

// UserManageAddRequest - 添加用户请求
type UserManageAddRequest struct {
	Username      string `json:"username"`
	Password      string `json:"passwd"`
	GroupName     string `json:"group_name"`
	Enabled       string `json:"enabled"`
	Comment       string `json:"comment"`
	IPAddr        string `json:"ip_addr,omitempty"`
	MacAddr       string `json:"mac_addr,omitempty"`
	BindMac       int    `json:"bind_mac,omitempty"`
	BindIP        int    `json:"bind_ip,omitempty"`
	SessTimeout   int    `json:"sesstimeout,omitempty"`
	MaxConn       int    `json:"max_conn,omitempty"`
	UploadLimit   int64  `json:"upload_limit,omitempty"`
	DownloadLimit int64  `json:"download_limit,omitempty"`
	ExpireTime    int64  `json:"expire_time,omitempty"`
}

// UserManageEditRequest - 编辑用户请求
type UserManageEditRequest struct {
	ID            int    `json:"id"`
	Username      string `json:"username,omitempty"`
	Password      string `json:"passwd,omitempty"`
	GroupName     string `json:"group_name,omitempty"`
	Enabled       string `json:"enabled,omitempty"`
	Comment       string `json:"comment,omitempty"`
	IPAddr        string `json:"ip_addr,omitempty"`
	MacAddr       string `json:"mac_addr,omitempty"`
	BindMac       int    `json:"bind_mac,omitempty"`
	BindIP        int    `json:"bind_ip,omitempty"`
	SessTimeout   int    `json:"sesstimeout,omitempty"`
	MaxConn       int    `json:"max_conn,omitempty"`
	UploadLimit   int64  `json:"upload_limit,omitempty"`
	DownloadLimit int64  `json:"download_limit,omitempty"`
	ExpireTime    int64  `json:"expire_time,omitempty"`
}
