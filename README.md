# iKuai SDK

一个用于与 iKuai 路由器进行交互的 Go SDK，支持 v3 和 v4 版本的 iKuai OS。

## 特性

- 自动检测 iKuai OS 版本 (v3/v4)
- 统一的 API 接口，屏蔽版本差异
- 支持 Context 超时和取消
- 完善的错误处理
- Cookie 会话管理

## 架构设计

```
sdk/
├── client.go          # 核心客户端
├── auth.go            # 认证模块
├── types.go           # 数据类型定义
├── version.go         # 版本枚举
├── errors.go          # 错误处理
├── api/               # API 功能模块
│   ├── system.go      # 系统状态 API
│   ├── network.go     # 网络相关 API
│   ├── custom_isp.go  # 自定义运营商
│   └── stream_domain.go # 域名分流
└── internal/          # 内部工具
    └── util.go
```

## 安装

```bash
go get github.com/zy84338719/ikuai-api
```

## 快速开始

### 创建客户端并登录

```go
package main

import (
    "context"
    "fmt"
    "time"

    ikuaisdk "github.com/zy84338719/ikuai-api"
)

func main() {
    // 方式1: 创建客户端后手动登录
    client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password",
        ikuaisdk.WithTimeout(30*time.Second),
    )
    defer client.Close()

    if err := client.Login(context.Background()); err != nil {
        panic(err)
    }

    // 方式2: 创建客户端并自动登录
    client, err := ikuaisdk.NewClientWithLogin("http://192.168.1.1", "admin", "password")
    if err != nil {
        panic(err)
    }
    defer client.Close()

    fmt.Printf("Connected to iKuai %s\n", client.GetVersion())
}
```

### 获取系统信息

```go
var resp ikuaisdk.HomepageResponse
err := client.Call(ctx, "homepage", "show", nil, &resp)
if err != nil {
    panic(err)
}

data := resp.GetData()
fmt.Printf("Version: %s\n", data.SystemInfo.Version)
fmt.Printf("Hostname: %s\n", data.SystemInfo.HostName)
fmt.Printf("Uptime: %d\n", data.SystemInfo.Uptime)
```

### 获取网络接口

```go
var resp ikuaisdk.InterfaceShowResponse
err := client.Call(ctx, "monitor_iface", "show", nil, &resp)

for _, iface := range resp.GetData() {
    fmt.Printf("%s: %s\n", iface.Name, iface.IP)
}
```

### 获取局域网设备

```go
var resp ikuaisdk.LANIPShowResponse
err := client.Call(ctx, "monitor_lanip", "show", nil, &resp)

for _, device := range resp.GetData() {
    fmt.Printf("%s (%s): %s\n", device.Hostname, device.Mac, device.IP)
}
```

### 使用服务模块

```go
import "github.com/zy84338719/ikuai-api/service"

// 创建系统服务
systemSvc := api.NewSystemService(client)
homepage, err := systemSvc.GetHomepage(ctx)

// 创建网络服务
networkSvc := api.NewNetworkService(client)
devices, err := networkSvc.GetLANDevices(ctx)

// 创建自定义运营商服务
ispSvc := api.NewCustomISPService(client)
items, err := ispSvc.List(ctx)

// 创建日志服务
logSvc := api.NewLogService(client)
notice, err := logSvc.GetNotice(ctx)
sysEvent, err := logSvc.GetSysEvent(ctx)

// 创建 Docker 服务
dockerSvc := api.NewDockerService(client)
images, err := dockerSvc.GetImages(ctx)
containers, err := dockerSvc.GetContainers(ctx)
networks, err := dockerSvc.GetNetworks(ctx)
```

## 版本兼容性

SDK 自动检测 iKuai OS 版本，通过响应格式区分：

| 版本 | 成功标识 | 错误字段 | 数据字段 |
|------|---------|---------|---------|
| v3 | Result=10000/30000 | ErrMsg | Data |
| v4 | code=0 | message | results |

## 错误处理

```go
err := client.Login(ctx)
if err != nil {
    if ikuaisdk.IsSDKError(err) {
        code := ikuaisdk.GetErrorCode(err)
        switch code {
        case ikuaisdk.ErrCodeLoginFailed:
            fmt.Println("登录失败")
        case ikuaisdk.ErrCodeNotLoggedIn:
            fmt.Println("未登录")
        default:
            fmt.Printf("SDK错误: %v\n", err)
        }
    }
}
```

## 运行测试

### 单元测试

```bash
cd sdk
go test ./...
```

### 集成测试

```bash
# 使用默认测试机器
go test -tags=integration ./...

# 或指定测试机器
IKUAI_TEST_ADDR=10.10.30.254 \
IKUAI_TEST_USERNAME=zhangyi \
IKUAI_TEST_PASSWORD=REDACTED \
go test -tags=integration -v ./...
```

## API 列表

### 核心方法

| 方法 | 说明 |
|------|------|
| `NewClient()` | 创建客户端 |
| `NewClientWithLogin()` | 创建客户端并自动登录 |
| `Login()` | 登录 |
| `Logout()` | 登出 |
| `Call()` | 调用 API |
| `GetVersion()` | 获取检测到的版本 |
| `IsLoggedIn()` | 检查登录状态 |

### 配置选项

| 选项 | 说明 |
|------|------|
| `WithTimeout()` | 设置超时时间 |
| `WithInsecureSkipVerify()` | 跳过 SSL 验证 |
| `WithHTTPClient()` | 自定义 HTTP 客户端 |

### 服务模块

#### MonitorService - 监控服务
| 方法 | 说明 |
|------|------|
| `GetLanIP()` | 获取局域网设备列表 |
| `GetLanIPv6()` | 获取局域网 IPv6 设备列表 |
| `GetInterfaces()` | 获取网络接口信息 |
| `GetSystem()` | 获取系统监控数据（CPU、内存、连接数等） |
| `GetARP()` | 获取 ARP 表 |

#### SystemService - 系统服务
| 方法 | 说明 |
|------|------|
| `GetHomepage()` | 获取首页系统状态 |
| `GetUpgradeInfo()` | 获取升级信息 |
| `GetBackupList()` | 获取备份列表 |
| `GetWebUsers()` | 获取 Web 用户列表 |

#### NetworkService - 网络服务
| 方法 | 说明 |
|------|------|
| `GetWan()` | 获取 WAN 配置 |
| `GetLan()` | 获取 LAN 配置 |
| `GetVLAN()` | 获取 VLAN 配置 |
| `GetIPv6()` | 获取 IPv6 配置 |
| `GetIPTV()` | 获取 IPTV 配置 |
| `GetDDNS()` | 获取 DDNS 配置 |
| `GetDHCPD()` | **新增** - 获取 DHCP 服务器配置 |
| `GetStaticBindings()` | **新增** - 获取 DHCP 静态绑定 |
| `GetLeases()` | **新增** - 获取 DHCP 租约信息 |

#### FirewallService - 防火墙服务
| 方法 | 说明 |
|------|------|
| `GetACL()` | 获取访问控制列表 |
| `GetDNAT()` | 获取端口映射 |
| `GetConnLimit()` | 获取连接数限制 |
| `GetDomainGroups()` | 获取域名分组 |
| `GetCustomISP()` | 获取自定义运营商 |
| `GetStreamDomain()` | 获取域名分流规则 |

#### VPNService - VPN 服务
| 方法 | 说明 |
|------|------|
| `GetPPTPClients()` | 获取 PPTP 客户端 |
| `GetL2TPClients()` | 获取 L2TP 客户端 |

#### LogService - 日志服务
| 方法 | 说明 |
|------|------|
| `GetNotice()` | 获取通知日志 |
| `GetWanPPPoE()` | 获取 PPPoE 日志 |
| `GetDHCPD()` | 获取 DHCP 日志 |
| `GetARP()` | 获取 ARP 日志 |
| `GetDDNS()` | 获取 DDNS 日志 |
| `GetWebAdmin()` | 获取 Web 管理日志 |
| `GetSysEvent()` | 获取系统事件日志 |

#### DockerService - Docker 服务
| 方法 | 说明 |
|------|------|
| `GetImages()` | 获取 Docker 镜像列表 |
| `GetContainers()` | 获取 Docker 容器列表 |
| `GetNetworks()` | 获取 Docker 网络列表 |
| `GetComposes()` | 获取 Docker Compose 项目列表 |

#### VMService - 虚拟机服务
| 方法 | 说明 |
|------|------|
| `List()` | 获取虚拟机列表 |
| `Add()` | 添加虚拟机 |
| `Edit()` | 编辑虚拟机配置 |
| `Del()` | 删除虚拟机 |
| `Start()` | 启动虚拟机 |
| `Stop()` | 停止虚拟机 |
| `Restart()` | 重启虚拟机 |

#### UPnPService - UPnP 服务
| 方法 | 说明 |
|------|------|
| `List()` | 获取 UPnP 映射列表 |
| `Add()` | 添加 UPnP 映射 |
| `Edit()` | 编辑 UPnP 映射 |
| `Del()` | 删除 UPnP 映射 |

## License

MIT
