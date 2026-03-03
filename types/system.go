package types

type HomepageShowResponse struct {
	BaseResponse
	Data struct {
		SysStat  HomepageSysStat `json:"sysstat"`
		ACStatus ACStatus        `json:"ac_status"`
	} `json:"Data"`
	Results *struct {
		SysStat  HomepageSysStat `json:"sysstat"`
		ACStatus ACStatus        `json:"ac_status"`
	} `json:"results,omitempty"`
}

type HomepageSysStat struct {
	CPU        []string   `json:"cpu"`
	CPUTemp    []int      `json:"cputemp"`
	Freq       []string   `json:"freq"`
	GWid       string     `json:"gwid"`
	Hostname   string     `json:"hostname"`
	LinkStatus int        `json:"link_status"`
	Memory     Memory     `json:"memory"`
	OnlineUser OnlineUser `json:"online_user"`
	Stream     Stream     `json:"stream"`
	Uptime     int        `json:"uptime"`
	VerInfo    VerInfo    `json:"verinfo"`
}

type Memory struct {
	Total     int64  `json:"total"`
	Available int64  `json:"available"`
	Free      int64  `json:"free"`
	Cached    int64  `json:"cached"`
	Buffers   int64  `json:"buffers"`
	Used      string `json:"used"`
}

type OnlineUser struct {
	Count         int `json:"count"`
	Count2G       int `json:"count_2g"`
	Count5G       int `json:"count_5g"`
	CountWired    int `json:"count_wired"`
	CountWireless int `json:"count_wireless"`
}

type Stream struct {
	ConnectNum int   `json:"connect_num"`
	Upload     int   `json:"upload"`
	Download   int   `json:"download"`
	TotalUp    int64 `json:"total_up"`
	TotalDown  int64 `json:"total_down"`
}

type VerInfo struct {
	ModelName    string `json:"modelname"`
	VerString    string `json:"verstring"`
	Version      string `json:"version"`
	BuildDate    int64  `json:"build_date"`
	Arch         string `json:"arch"`
	SysBit       string `json:"sysbit"`
	VerFlags     string `json:"verflags"`
	IsEnterprise int    `json:"is_enterprise"`
	SupportI18N  int    `json:"support_i18n"`
	SupportLCD   int    `json:"support_lcd"`
}

type ACStatus struct {
	APCount  int `json:"ap_count"`
	APOnline int `json:"ap_online"`
}

func (r *HomepageShowResponse) GetData() *HomepageSysStat {
	if r.Results != nil {
		return &r.Results.SysStat
	}
	return &r.Data.SysStat
}

func (r *HomepageShowResponse) GetACStatus() *ACStatus {
	if r.Results != nil {
		return &r.Results.ACStatus
	}
	return &r.Data.ACStatus
}

type UpgradeShowResponse struct {
	BaseResponse
	Data struct {
		Data UpgradeInfo `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data UpgradeInfo `json:"data"`
	} `json:"results,omitempty"`
}

type UpgradeInfo struct {
	SystemVer            string `json:"system_ver"`
	NewSystemVer         string `json:"new_system_ver"`
	BuildDate            string `json:"build_date"`
	NewBuildDate         string `json:"new_build_date"`
	VersionType          string `json:"version_type"`
	BootGuide            string `json:"bootguide"`
	UpdateContent        string `json:"update_content"`
	AutoUpgradeSec       int    `json:"auto_upgrade_sec"`
	AutoUpgradeLibDPI    int    `json:"auto_upgrade_lib_dpi"`
	AutoUpgradeLibIM     int    `json:"auto_upgrade_lib_im"`
	AutoUpgradeLibDomain int    `json:"auto_upgrade_lib_domain"`
	LibProtoVer          string `json:"libproto_ver"`
	NewLibProtoVer       string `json:"new_libproto_ver"`
	LibDomainVer         string `json:"libdomain_ver"`
	NewLibDomainVer      string `json:"new_libdomain_ver"`
	LibAuditVer          string `json:"libaudit_ver"`
	NewLibAuditVer       string `json:"new_libaudit_ver"`
	IgnoreUpgradeVer     string `json:"ignore_upgrade_ver"`
}

func (r *UpgradeShowResponse) GetData() *UpgradeInfo {
	if r.Results != nil {
		return &r.Results.Data
	}
	return &r.Data.Data
}

type BackupShowResponse struct {
	BaseResponse
	Data struct {
		Data []BackupItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []BackupItem `json:"data"`
	} `json:"results,omitempty"`
}

type BackupItem struct {
	ID        int    `json:"id"`
	Enabled   string `json:"enabled"`
	Strategy  string `json:"strategy"`
	CycleTime string `json:"cycle_time"`
	Time      string `json:"time"`
	ValidDays int    `json:"valid_days"`
	TagName   string `json:"tagname"`
}

func (r *BackupShowResponse) GetData() []BackupItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type WebUserShowResponse struct {
	BaseResponse
	Data struct {
		Data []WebUserItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []WebUserItem `json:"data"`
	} `json:"results,omitempty"`
}

type WebUserItem struct {
	ID            int    `json:"id"`
	Username      string `json:"username"`
	GroupID       int    `json:"group_id"`
	GroupName     string `json:"group_name"`
	Enabled       string `json:"enabled"`
	Comment       string `json:"comment"`
	IPAddr        string `json:"ip_addr"`
	Passwd        string `json:"passwd"`
	SessTimeout   int    `json:"sesstimeout"`
	PasswdTimeout int    `json:"passwd_timeout"`
	Force         int    `json:"force"`
	Interval      int    `json:"interval"`
	PermDefault   string `json:"perm_default"`
	PermConfig    string `json:"perm_config"`
	CustomChart   string `json:"custom_chart"`
	CustomPlugin  string `json:"custom_plugin"`
}

func (r *WebUserShowResponse) GetData() []WebUserItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}
