# iKuai SDK

[![Go Report Card](https://goreportcard.com/badge/github.com/zy84338719/ikuai-api)](https://goreportcard.com/report/github.com/zy84338719/ikuai-api)
[![GoDoc](https://godoc.org/github.com/zy84338719/ikuai-api?status.svg)](https://godoc.org/github.com/zy84338719/ikuai-api)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

一个用于与 iKuai 路由器进行交互的 Go SDK，支持 v3 和 v4 版本的 iKuai OS。

## ✨ 特性

- 🚀 **基于 reqv3 HTTP 客户端** - 使用强大的 [github.com/imroc/req/v3](https://github.com/imroc/req)，提供更好的调试、重试、HTTP2/3 支持
- 🔌 **接口抽象设计** - 所有服务通过接口定义，方便测试和扩展
- 🔄 **自动版本检测** - 自动检测 iKuai OS 版本 (v3/v4)，统一 API 接口
- 🛡️ **完善的错误处理** - 统一的错误类型和错误码
- ⏱️ **Context 支持** - 完整的超时和取消支持
- 🍪 **Cookie 会话管理** - 自动管理登录会话

## 📦 架构设计

```
sdk/
├── client.go              # 核心客户端（基于 reqv3）
├── auth.go                # 认证模块
├── errors.go              # 错误处理
├── version.go             # 版本枚举
├── internal/              # 内部工具
│   └── util.go
├── types/                 # 数据类型定义
│   ├── base.go
│   ├── system.go
│   ├── network.go
│   ├── monitor.go
│   ├── firewall.go
│   ├── vpn.go
│   ├── docker.go
│   ├── vm.go
│   └── upnp.go
├── service/               # 服务层（接口设计）
│   ├── interface.go       # 所有服务接口定义
│   ├── client.go          # APIClient 统一入口
│   ├── monitor.go         # 监控服务
│   ├── system.go          # 系统服务
│   ├── network.go         # 网络服务
│   ├── firewall.go        # 防火墙服务
│   ├── vpn.go             # VPN 服务
│   ├── log.go             # 日志服务
│   ├── docker.go          # Docker 服务
│   ├── vm.go              # 虚拟机服务
│   └── upnp.go            # UPnP 服务
└── example/               # 示例代码
    └── main.go
```

### 核心设计理念

#### 1. 接口抽象
所有服务都通过接口定义，而不是具体的结构体，这样可以：
- ✅ **易于测试** - 可以轻松创建 mock 实现进行单元测试
- ✅ **易于扩展** - 其他实现可以无缝替换默认实现
- ✅ **解耦合** - 调用方只依赖接口，不依赖具体实现

#### 2. 统一入口
通过 `APIClient` 提供所有服务的统一访问入口，避免重复创建服务实例。

## 📥 安装

```bash
go get github.com/zy84338719/ikuai-api
```

## 🚀 快速开始

### 创建客户端并登录

```go
package main

import (
    "context"
    "fmt"
    "time"

    ikuaisdk "github.com/zy84338719/ikuai-api"
    "github.com/zy84338719/ikuai-api/service"
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

    // 方式2: 创建客户端并自动登录（推荐）
    client, err := ikuaisdk.NewClientWithLogin("http://192.168.1.1", "admin", "password")
    if err != nil {
        panic(err)
    }
    defer client.Close()

    fmt.Printf("Connected to iKuai %s\n", client.GetVersion())
}
```

### 使用服务层 API（推荐）

```go
// 创建 API 客户端
api := service.NewAPIClient(client)

// 获取系统信息
homepage, err := api.System().GetHomepage(context.Background())
if err != nil {
    panic(err)
}
fmt.Printf("Version: %s\n", homepage.VerInfo.Version)
fmt.Printf("Hostname: %s\n", homepage.Hostname)
fmt.Printf("Uptime: %d seconds\n", homepage.Uptime)

// 获取局域网设备
devices, err := api.Monitor().GetLanIP(context.Background())
if err != nil {
    panic(err)
}
for _, device := range devices {
    fmt.Printf("%s (%s): %s\n", device.Hostname, device.Mac, device.IP)
}

// 获取网络接口
interfaces, err := api.Monitor().GetInterfaces(context.Background())
if err != nil {
    panic(err)
}
for _, iface := range interfaces.GetData() {
    fmt.Printf("%s: %s\n", iface.Name, iface.IP)
}
```

### 使用底层 Call 方法

```go
// 直接调用任意 API
var resp types.HomepageShowResponse
err := client.Call(ctx, "homepage", "show", nil, &resp)
if err != nil {
    panic(err)
}

data := resp.GetData()
fmt.Printf("Version: %s\n", data.VerInfo.Version)
```

### 单独使用某个服务

```go
// 只创建需要的服务
monitorSvc := service.NewMonitorService(client)
devices, err := monitorSvc.GetLanIP(context.Background())
```

## 🔧 HTTP 客户端配置

SDK 使用 [reqv3](https://github.com/imroc/req) 作为 HTTP 客户端，提供强大的功能：

### 配置选项

```go
client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password",
    ikuaisdk.WithTimeout(60*time.Second),              // 设置超时时间
    ikuaisdk.WithInsecureSkipVerify(true),             // 跳过 SSL 验证
)

// 如果需要更高级的配置，可以传入自定义的 req.Client
customReqClient := req.C().
    SetTimeout(60*time.Second).
    EnableInsecureSkipVerify().
    EnableDumpEachRequest()  // 启用请求/响应 dump，便于调试

client := ikuaisdk.NewClient("http://192.168.1.1", "admin", "password",
    ikuaisdk.WithHTTPClient(customReqClient),
)
```

### reqv3 的优势

相比标准库 `net/http`，reqv3 提供：

- ✅ **更简单的 API** - 链式调用，代码更简洁
- ✅ **自动重试** - 支持自定义重试策略
- ✅ **调试友好** - 内置请求/响应 dump 和日志
- ✅ **HTTP/2 & HTTP/3** - 自动选择最优协议
- ✅ **自动解码** - 智能检测并解码响应
- ✅ **更好的错误处理** - 统一的错误类型

## 📚 服务列表

### MonitorService - 监控服务
获取系统和网络监控数据

| 方法 | 说明 |
|------|------|
| `GetLanIP()` | 获取局域网设备列表 |
| `GetLanIPv6()` | 获取局域网 IPv6 设备列表 |
| `GetInterfaces()` | 获取网络接口信息 |
| `GetSystem()` | 获取系统监控数据（CPU、内存、连接数等） |
| `GetARP()` | 获取 ARP 表 |

### SystemService - 系统服务
系统配置和状态管理

| 方法 | 说明 |
|------|------|
| `GetHomepage()` | 获取首页系统状态 |
| `GetUpgradeInfo()` | 获取升级信息 |
| `GetBackupList()` | 获取备份列表 |
| `GetWebUsers()` | 获取 Web 用户列表 |

### NetworkService - 网络服务
网络配置管理

| 方法 | 说明 |
|------|------|
| `GetWan()` | 获取 WAN 配置 |
| `GetLan()` | 获取 LAN 配置 |
| `GetVLAN()` | 获取 VLAN 配置 |
| `GetIPv6()` | 获取 IPv6 配置 |
| `GetIPTV()` | 获取 IPTV 配置 |
| `GetDDNS()` | 获取 DDNS 配置 |
| `GetDHCPD()` | 获取 DHCP 服务器配置 |
| `GetStaticBindings()` | 获取 DHCP 静态绑定 |
| `GetLeases()` | 获取 DHCP 租约信息 |

### FirewallService - 防火墙服务
防火墙规则管理

| 方法 | 说明 |
|------|------|
| `GetACL()` | 获取访问控制列表 |
| `GetDNAT()` | 获取端口映射 |
| `GetConnLimit()` | 获取连接数限制 |
| `GetDomainGroups()` | 获取域名分组 |
| `GetCustomISP()` | 获取自定义运营商 |
| `GetStreamDomain()` | 获取域名分流规则 |

### VPNService - VPN 服务
VPN 客户端管理

| 方法 | 说明 |
|------|------|
| `GetPPTPClients()` | 获取 PPTP 客户端 |
| `GetL2TPClients()` | 获取 L2TP 客户端 |

### LogService - 日志服务
系统日志查询

| 方法 | 说明 |
|------|------|
| `GetNotice()` | 获取通知日志 |
| `GetWanPPPoE()` | 获取 PPPoE 日志 |
| `GetDHCPD()` | 获取 DHCP 日志 |
| `GetARP()` | 获取 ARP 日志 |
| `GetDDNS()` | 获取 DDNS 日志 |
| `GetWebAdmin()` | 获取 Web 管理日志 |
| `GetSysEvent()` | 获取系统事件日志 |

### DockerService - Docker 服务
Docker 容器管理

| 方法 | 说明 |
|------|------|
| `GetImages()` | 获取 Docker 镜像列表 |
| `GetContainers()` | 获取 Docker 容器列表 |
| `GetNetworks()` | 获取 Docker 网络列表 |
| `GetComposes()` | 获取 Docker Compose 项目列表 |

### VMService - 虚拟机服务
虚拟机管理

| 方法 | 说明 |
|------|------|
| `List()` | 获取虚拟机列表 |
| `Add()` | 添加虚拟机 |
| `Edit()` | 编辑虚拟机配置 |
| `Del()` | 删除虚拟机 |
| `Start()` | 启动虚拟机 |
| `Stop()` | 停止虚拟机 |
| `Restart()` | 重启虚拟机 |

### UPnPService - UPnP 服务
UPnP 映射管理

| 方法 | 说明 |
|------|------|
| `List()` | 获取 UPnP 映射列表 |
| `Add()` | 添加 UPnP 映射 |
| `Edit()` | 编辑 UPnP 映射 |
| `Del()` | 删除 UPnP 映射 |

## 📋 核心方法

| 方法 | 说明 |
|------|------|
| `NewClient()` | 创建客户端 |
| `NewClientWithLogin()` | 创建客户端并自动登录 |
| `Login()` | 登录 |
| `Logout()` | 登出 |
| `Call()` | 调用 API |
| `GetVersion()` | 获取检测到的版本 |
| `IsLoggedIn()` | 检查登录状态 |

## 🛡️ 错误处理

SDK 提供了统一的错误处理机制：

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
        case ikuaisdk.ErrCodeRequestFailed:
            fmt.Println("请求失败")
        case ikuaisdk.ErrCodeInvalidResponse:
            fmt.Println("响应格式错误")
        default:
            fmt.Printf("SDK错误: %v\n", err)
        }
    } else {
        fmt.Printf("其他错误: %v\n", err)
    }
}
```

## 🧪 运行测试

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
IKUAI_TEST_ADDR=192.168.1.1 \
IKUAI_TEST_USERNAME=admin \
IKUAI_TEST_PASSWORD=admin123 \
go test -tags=integration -v ./...
```

## 📖 更多文档

- [Service 层详细文档](service/README.md) - 接口设计、自定义实现、最佳实践
- [reqv3 官方文档](https://req.cool) - HTTP 客户端完整功能

## 📝 版本兼容性

SDK 自动检测 iKuai OS 版本，通过响应格式区分：

| 版本 | 成功标识 | 错误字段 | 数据字段 |
|------|---------|---------|---------|
| v3 | Result=10000/30000 | ErrMsg | Data |
| v4 | code=0 | message | results |

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 License

[MIT](LICENSE)
