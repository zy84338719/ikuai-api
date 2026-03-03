package types

type QemuShowResponse struct {
	BaseResponse
	Data struct {
		Data []QemuVM `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []QemuVM `json:"data"`
	} `json:"results,omitempty"`
}

type QemuVM struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	TagName    string `json:"tagname"`
	Comment    string `json:"comment"`
	System     string `json:"system"`
	CPUCores   int    `json:"cpu_cores"`
	CPUUsage   int    `json:"cpu_usage"`
	MemSize    int    `json:"mem_size"`
	Status     int    `json:"status"`
	Enabled    string `json:"enabled"`
	AutoStart  int    `json:"auto_start"`
	Iso        string `json:"iso"`
	VDisk      string `json:"vdisk"`
	BrName     string `json:"brname"`
	VNCPort    int    `json:"vnc_port"`
	VNCPwd     string `json:"vnc_pwd"`
	VNCAcl     int    `json:"vnc_acl"`
	Accel      int    `json:"accel"`
	UEFI       int    `json:"uefi"`
	USB        string `json:"usb"`
	PartName   string `json:"partname"`
	ElapseTime int64  `json:"elapse_time"`
}

func (r *QemuShowResponse) GetData() []QemuVM {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type QemuAddRequest struct {
	Name      string `json:"name"`
	TagName   string `json:"tagname,omitempty"`
	Comment   string `json:"comment,omitempty"`
	System    string `json:"system"`
	CPUCores  int    `json:"cpu_cores"`
	MemSize   int    `json:"mem_size"`
	Iso       string `json:"iso,omitempty"`
	VDisk     string `json:"vdisk,omitempty"`
	BrName    string `json:"brname"`
	VNCPort   int    `json:"vnc_port,omitempty"`
	VNCPwd    string `json:"vnc_pwd,omitempty"`
	AutoStart int    `json:"auto_start,omitempty"`
	Accel     int    `json:"accel,omitempty"`
	UEFI      int    `json:"uefi,omitempty"`
	USB       string `json:"usb,omitempty"`
}

type QemuEditRequest struct {
	ID        int    `json:"id"`
	Name      string `json:"name,omitempty"`
	TagName   string `json:"tagname,omitempty"`
	Comment   string `json:"comment,omitempty"`
	CPUCores  int    `json:"cpu_cores,omitempty"`
	MemSize   int    `json:"mem_size,omitempty"`
	Iso       string `json:"iso,omitempty"`
	VDisk     string `json:"vdisk,omitempty"`
	BrName    string `json:"brname,omitempty"`
	VNCPort   int    `json:"vnc_port,omitempty"`
	VNCPwd    string `json:"vnc_pwd,omitempty"`
	AutoStart int    `json:"auto_start,omitempty"`
}

type QemuDelRequest struct {
	ID int `json:"id"`
}

type QemuStartRequest struct {
	ID int `json:"id"`
}

type QemuStopRequest struct {
	ID int `json:"id"`
}

type QemuRestartRequest struct {
	ID int `json:"id"`
}
