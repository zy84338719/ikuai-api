package types

type UPnPShowResponse struct {
	BaseResponse
	Data struct {
		Data []UPnPItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []UPnPItem `json:"data"`
	} `json:"results,omitempty"`
}

type UPnPItem struct {
	ID            int    `json:"id"`
	Interface     string `json:"interface"`
	Protocol      string `json:"protocol"`
	ExternalPort  int    `json:"external_port"`
	InternalIP    string `json:"internal_ip"`
	InternalPort  int    `json:"internal_port"`
	Description   string `json:"description"`
	Enabled       string `json:"enabled"`
	Comment       string `json:"comment"`
	LeaseDuration int    `json:"lease_duration"`
}

func (r *UPnPShowResponse) GetData() []UPnPItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type UPnPAddRequest struct {
	Interface    string `json:"interface"`
	Protocol     string `json:"protocol"`
	ExternalPort int    `json:"external_port"`
	InternalIP   string `json:"internal_ip"`
	InternalPort int    `json:"internal_port"`
	Description  string `json:"description"`
	Enabled      string `json:"enabled"`
	Comment      string `json:"comment,omitempty"`
}

type UPnPEditRequest struct {
	ID           int    `json:"id"`
	Interface    string `json:"interface,omitempty"`
	Protocol     string `json:"protocol,omitempty"`
	ExternalPort int    `json:"external_port,omitempty"`
	InternalIP   string `json:"internal_ip,omitempty"`
	InternalPort int    `json:"internal_port,omitempty"`
	Description  string `json:"description,omitempty"`
	Enabled      string `json:"enabled,omitempty"`
	Comment      string `json:"comment,omitempty"`
}

type UPnPDelRequest struct {
	ID int `json:"id"`
}
