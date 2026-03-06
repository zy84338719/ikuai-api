# Service API

这个包提供了 ikuai API 的服务层接口，采用接口抽象设计，方便其他方对接和实现。

## 设计理念

### 接口抽象

所有服务都通过接口定义，而不是具体的结构体，这样可以：

1. **易于测试**：可以轻松创建 mock 实现进行单元测试
2. **易于扩展**：其他实现可以无缝替换默认实现
3. **解耦合**：调用方只依赖接口，不依赖具体实现

### 文件结构

```
service/
├── interface.go     # 所有服务接口定义
├── client.go        # APIClient 统一入口
├── monitor.go       # 监控服务实现
├── system.go        # 系统服务实现
├── network.go       # 网络服务实现
├── firewall.go      # 防火墙服务实现
├── vpn.go          # VPN 服务实现
├── log.go          # 日志服务实现
├── docker.go       # Docker 服务实现
├── vm.go           # 虚拟机服务实现
├── upnp.go         # UPnP 服务实现
├── traffic.go      # 流量统计服务实现
├── appcontrol.go   # 应用管控服务实现
├── usermanage.go   # 用户管理服务实现
└── onlinemonitor.go # 在线监控服务实现
```

## 快速开始

### 使用默认实现

```go
package main

import (
    "context"
    "fmt"

    ikuaisdk "github.com/zy84338719/ikuai-api"
    "github.com/zy84338719/ikuai-api/service"
)

func main() {
    // 创建客户端并登录
    client, err := ikuaisdk.NewClientWithLogin("http://192.168.1.1", "admin", "admin123")
    if err != nil {
        panic(err)
    }
    defer client.Close()

    // 创建 API 客户端
    api := service.NewAPIClient(client)

    // 使用系统服务
    homepage, err := api.System().GetHomepage(context.Background())
    if err != nil {
        panic(err)
    }
    fmt.Printf("Version: %s\n", homepage.VerInfo.Version)

    // 使用监控服务
    devices, err := api.Monitor().GetLanIP(context.Background())
    if err != nil {
        panic(err)
    }
    fmt.Printf("Found %d devices\n", len(devices))
}
```

### 单独使用某个服务

如果只需要使用某个特定的服务，可以单独创建：

```go
client, _ := ikuaisdk.NewClientWithLogin("http://192.168.1.1", "admin", "admin123")
defer client.Close()

// 只创建监控服务
monitorSvc := service.NewMonitorService(client)
devices, _ := monitorSvc.GetLanIP(context.Background())
```

## 接口说明

### APIClient

统一入口，提供所有服务的访问：

```go
type APIClient interface {
    Monitor() MonitorService
    System() SystemService
    Network() NetworkService
    Firewall() FirewallService
    VPN() VPNService
    Log() LogService
    Docker() DockerService
    VM() VMService
    UPnP() UPnPService
    Traffic() TrafficService
    AppControl() AppControlService
    UserManage() UserManageService
    OnlineMonitor() OnlineMonitorService
}
```

### MonitorService

监控服务，提供系统和网络监控数据：

```go
type MonitorService interface {
    GetLanIP(ctx context.Context) ([]types.MonitorLanIPItem, error)
    GetLanIPv6(ctx context.Context) ([]types.MonitorLanIPv6Item, error)
    GetInterfaces(ctx context.Context) (*types.MonitorIFaceShowResponse, error)
    GetSystem(ctx context.Context) ([]types.MonitorSystemItem, error)
    GetARP(ctx context.Context) ([]types.ARPItem, error)
}
```

### SystemService

系统服务，提供系统配置和状态：

```go
type SystemService interface {
    GetHomepage(ctx context.Context) (*types.HomepageSysStat, error)
    GetUpgradeInfo(ctx context.Context) (*types.UpgradeInfo, error)
    GetBackupList(ctx context.Context) ([]types.BackupItem, error)
    GetWebUsers(ctx context.Context) ([]types.WebUserItem, error)
}
```

### NetworkService

网络服务，提供网络配置：

```go
type NetworkService interface {
    GetWan(ctx context.Context) ([]types.WanItem, error)
    GetLan(ctx context.Context) ([]types.LanItem, error)
    GetVLAN(ctx context.Context) ([]types.VLANItem, error)
    GetIPv6(ctx context.Context) ([]types.IPv6Item, error)
    GetIPTV(ctx context.Context) ([]types.IPTVItem, error)
    GetDDNS(ctx context.Context) ([]types.DDNSItem, error)
    GetDHCPD(ctx context.Context) ([]types.DHCPDItem, error)
    GetStaticBindings(ctx context.Context) ([]types.DHCPStaticItem, error)
    GetLeases(ctx context.Context) ([]types.DHCPLeaseItem, error)
}
```

### FirewallService

防火墙服务：

```go
type FirewallService interface {
    GetACL(ctx context.Context) ([]types.ACLItem, error)
    GetDNAT(ctx context.Context) ([]types.DNATItem, error)
    GetConnLimit(ctx context.Context) ([]types.ConnLimitItem, error)
    GetDomainGroups(ctx context.Context) ([]types.DomainGroupItem, error)
    GetCustomISP(ctx context.Context) ([]types.CustomISPItem, error)
    GetStreamDomain(ctx context.Context) ([]types.StreamDomainItem, error)
}
```

### VPNService

VPN 服务：

```go
type VPNService interface {
    GetPPTPClients(ctx context.Context) ([]types.PPTPClientItem, error)
    GetL2TPClients(ctx context.Context) ([]types.L2TPClientItem, error)
}
```

### LogService

日志服务：

```go
type LogService interface {
    GetNotice(ctx context.Context) ([]types.SyslogItem, error)
    GetWanPPPoE(ctx context.Context) ([]types.SyslogWanPPPoEItem, error)
    GetDHCPD(ctx context.Context) ([]types.SyslogItem, error)
    GetARP(ctx context.Context) ([]types.SyslogItem, error)
    GetDDNS(ctx context.Context) ([]types.SyslogItem, error)
    GetWebAdmin(ctx context.Context) ([]types.SyslogItem, error)
    GetSysEvent(ctx context.Context) ([]types.SyslogItem, error)
}
```

### DockerService

Docker 服务：

```go
type DockerService interface {
    GetImages(ctx context.Context) ([]types.DockerImageItem, error)
    GetContainers(ctx context.Context) ([]types.DockerContainerItem, error)
    GetNetworks(ctx context.Context) ([]types.DockerNetworkItem, error)
    GetComposes(ctx context.Context) ([]types.DockerComposeItem, error)
}
```

### VMService

虚拟机服务：

```go
type VMService interface {
    List(ctx context.Context) ([]types.QemuVM, error)
    Add(ctx context.Context, req *types.QemuAddRequest) (int, error)
    Edit(ctx context.Context, req *types.QemuEditRequest) error
    Del(ctx context.Context, id int) error
    Start(ctx context.Context, id int) error
    Stop(ctx context.Context, id int) error
    Restart(ctx context.Context, id int) error
}
```

### UPnPService

UPnP 服务：

```go
type UPnPService interface {
    List(ctx context.Context) ([]types.UPnPItem, error)
    Add(ctx context.Context, req *types.UPnPAddRequest) (int, error)
    Edit(ctx context.Context, req *types.UPnPEditRequest) error
    Del(ctx context.Context, id int) error
}
```

### TrafficService

流量统计服务：

```go
type TrafficService interface {
    GetRealtime(ctx context.Context) ([]types.TrafficRealtimeItem, error)
    GetHistory(ctx context.Context, hours int64) ([]types.TrafficHistoryItem, error)
}
```

### AppControlService

应用管控服务：

```go
type AppControlService interface {
    GetAppControl(ctx context.Context) ([]types.AppControlItem, error)
    AddAppControl(ctx context.Context, req *types.AppControlAddRequest) (int, error)
    EditAppControl(ctx context.Context, req *types.AppControlEditRequest) error
    DelAppControl(ctx context.Context, id int) error
}
```

### UserManageService

用户管理服务：

```go
type UserManageService interface {
    GetUsers(ctx context.Context) ([]types.UserManageItem, error)
    AddUser(ctx context.Context, req *types.UserManageAddRequest) (int, error)
    EditUser(ctx context.Context, req *types.UserManageEditRequest) error
    DelUser(ctx context.Context, id int) error
}
```

### OnlineMonitorService

在线监控服务：

```go
type OnlineMonitorService interface {
    GetOnlineUsers(ctx context.Context) ([]types.OnlineMonitorItem, error)
}
```

## 自定义实现

如果需要自定义实现（例如 mock 测试或适配其他系统），只需实现对应的接口：

```go
type MockMonitorService struct{}

func (m *MockMonitorService) GetLanIP(ctx context.Context) ([]types.MonitorLanIPItem, error) {
    return []types.MonitorLanIPItem{
        {IP: "192.168.1.100", Mac: "00:11:22:33:44:55", Hostname: "device1"},
    }, nil
}

// ... 实现其他方法

// 使用自定义实现
func main() {
    var monitorSvc service.MonitorService = &MockMonitorService{}
    devices, _ := monitorSvc.GetLanIP(context.Background())
    fmt.Printf("Mock devices: %v\n", devices)
}
```

## 最佳实践

1. **使用 APIClient**：推荐使用 `NewAPIClient` 获取所有服务，避免重复创建
2. **接口依赖**：在业务代码中依赖接口而非具体实现，便于测试和替换
3. **Context 传递**：所有方法都支持 context，可以实现超时控制和取消
4. **错误处理**：统一返回 error 接口，调用方需要检查错误

## 示例

完整示例请参考 `example/main.go`
