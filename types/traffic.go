package types

type TrafficRealtimeShowResponse struct {
	BaseResponse
	Data struct {
		Data []TrafficRealtimeItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []TrafficRealtimeItem `json:"data"`
	} `json:"results,omitempty"`
}

type TrafficRealtimeItem struct {
	ID            int    `json:"id"`
	Interface     string `json:"interface"`
	Upload        int64  `json:"upload"`
	Download      int64  `json:"download"`
	UploadSpeed   int64  `json:"upload_speed"`
	DownloadSpeed int64  `json:"download_speed"`
	Timestamp     int64  `json:"timestamp"`
}

func (r *TrafficRealtimeShowResponse) GetData() []TrafficRealtimeItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type TrafficHistoryShowResponse struct {
	BaseResponse
	Data struct {
		Data []TrafficHistoryItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []TrafficHistoryItem `json:"data"`
	} `json:"results,omitempty"`
}

type TrafficHistoryItem struct {
	ID        int    `json:"id"`
	Interface string `json:"interface"`
	Upload    int64  `json:"upload"`
	Download  int64  `json:"download"`
	Timestamp int64  `json:"timestamp"`
	Date      string `json:"date"`
}

type TrafficHistoryRequest struct {
	Hours int `json:"hours"`
}

func (r *TrafficHistoryShowResponse) GetData() []TrafficHistoryItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}
