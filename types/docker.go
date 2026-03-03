package types

type DockerImageShowResponse struct {
	BaseResponse
	Data struct {
		Data []DockerImageItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []DockerImageItem `json:"data"`
	} `json:"results,omitempty"`
}

type DockerImageContainer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

type DockerImageItem struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Tag            string                 `json:"tag"`
	Size           int64                  `json:"size"`
	Created        int64                  `json:"created"`
	Install        int64                  `json:"install"`
	Status         int                    `json:"status"`
	ImageLogo      string                 `json:"image_logo"`
	ContainerCount int                    `json:"container_count"`
	Containers     []DockerImageContainer `json:"containers"`
}

func (r *DockerImageShowResponse) GetData() []DockerImageItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type DockerContainerShowResponse struct {
	BaseResponse
	Data struct {
		Data []DockerContainerItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []DockerContainerItem `json:"data"`
	} `json:"results,omitempty"`
}

type DockerPortMapping struct {
	IP          string `json:"IP"`
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort"`
	Type        string `json:"Type"`
}

type DockerContainerItem struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Image      string              `json:"image"`
	ImageLogo  string              `json:"image_logo"`
	State      string              `json:"state"`
	Status     string              `json:"status"`
	Created    int64               `json:"created"`
	AutoStart  string              `json:"auto_start"`
	Cmd        string              `json:"cmd"`
	Comment    string              `json:"comment"`
	CPUUsed    string              `json:"cpu_used"`
	Memory     int64               `json:"memory"`
	MemUsed    int64               `json:"memused"`
	Gateway    string              `json:"gateway"`
	IPAddr     string              `json:"ipaddr"`
	IP6Addr    string              `json:"ip6addr"`
	IP6Gateway string              `json:"ip6gateway"`
	Mac        string              `json:"mac"`
	Interface  string              `json:"interface"`
	Env        string              `json:"env"`
	Mounts     string              `json:"mounts"`
	Ports      []DockerPortMapping `json:"ports"`
	ExitCode   int                 `json:"exitcode"`
	UpSpeed    int                 `json:"up_speed"`
	DownSpeed  int                 `json:"down_speed"`
}

func (r *DockerContainerShowResponse) GetData() []DockerContainerItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type DockerNetworkShowResponse struct {
	BaseResponse
	Data struct {
		Data []DockerNetworkItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []DockerNetworkItem `json:"data"`
	} `json:"results,omitempty"`
}

type DockerNetworkItem struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Comment    string   `json:"comment"`
	Subnet     string   `json:"subnet"`
	Subnet6    string   `json:"subnet6"`
	Gateway    string   `json:"gateway"`
	Gateway6   string   `json:"gateway6"`
	Containers []string `json:"containers"`
}

func (r *DockerNetworkShowResponse) GetData() []DockerNetworkItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}

type DockerComposeShowResponse struct {
	BaseResponse
	Data struct {
		Data []DockerComposeItem `json:"data"`
	} `json:"Data"`
	Results *struct {
		Data []DockerComposeItem `json:"data"`
	} `json:"results,omitempty"`
}

type DockerComposeItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Created int64  `json:"created"`
	Path    string `json:"path"`
}

func (r *DockerComposeShowResponse) GetData() []DockerComposeItem {
	if r.Results != nil {
		return r.Results.Data
	}
	return r.Data.Data
}
