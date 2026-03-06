//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/service"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== 测试 iKuai SDK 连接 ===")
	fmt.Println()

	addr := os.Getenv("IKUAI_TEST_ADDR")
	if addr == "" {
		addr = "192.168.1.1"
	}
	username := os.Getenv("IKUAI_TEST_USERNAME")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("IKUAI_TEST_PASSWORD")
	if password == "" {
		fmt.Println("❌ 请设置环境变量 IKUAI_TEST_PASSWORD")
		return
	}

	fmt.Println("1. 连接到路由器并登录...")
	client, err := ikuaisdk.NewClientWithLogin(
		addr,
		username,
		password,
		ikuaisdk.WithTimeout(30*time.Second),
	)
	if err != nil {
		fmt.Printf("❌ 登录失败: %v\n", err)
		return
	}
	defer client.Close()
	fmt.Printf("✅ 登录成功！iKuai 版本: %s\n\n", client.GetVersion())

	// 创建 API 客户端
	api := service.NewAPIClient(client)

	// 测试系统服务
	fmt.Println("2. 测试系统服务...")
	testSystemService(ctx, api)
	fmt.Println()

	// 测试监控服务
	fmt.Println("3. 测试监控服务...")
	testMonitorService(ctx, api)
	fmt.Println()

	// 测试网络服务
	fmt.Println("4. 测试网络服务...")
	testNetworkService(ctx, api)
	fmt.Println()

	// 测试防火墙服务
	fmt.Println("5. 测试防火墙服务...")
	testFirewallService(ctx, api)
	fmt.Println()

	fmt.Println("=== 测试完成 ===")
}

func testSystemService(ctx context.Context, api service.APIClient) {
	// 获取系统首页信息
	homepage, err := api.System().GetHomepage(ctx)
	if err != nil {
		fmt.Printf("  ❌ GetHomepage 失败: %v\n", err)
		return
	}
	fmt.Printf("  ✅ GetHomepage 成功\n")
	fmt.Printf("     - 版本: %s\n", homepage.VerInfo.Version)
	fmt.Printf("     - 主机名: %s\n", homepage.Hostname)
	fmt.Printf("     - 运行时间: %d 秒\n", homepage.Uptime)

	// 获取升级信息
	upgrade, err := api.System().GetUpgradeInfo(ctx)
	if err != nil {
		fmt.Printf("  ❌ GetUpgradeInfo 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ GetUpgradeInfo 成功\n")
		fmt.Printf("     - 当前版本: %s\n", upgrade.SystemVer)
	}

	// 获取备份列表
	backups, err := api.System().GetBackupList(ctx)
	if err != nil {
		fmt.Printf("  ❌ GetBackupList 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ GetBackupList 成功 (共 %d 个备份)\n", len(backups))
	}

	// 获取 Web 用户列表
	users, err := api.System().GetWebUsers(ctx)
	if err != nil {
		fmt.Printf("  ❌ GetWebUsers 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ GetWebUsers 成功 (共 %d 个用户)\n", len(users))
		for i, user := range users {
			if i < 3 { // 只显示前3个
				fmt.Printf("     - 用户: %s (ID: %d)\n", user.Username, user.ID)
			}
		}
	}
}

func testMonitorService(ctx context.Context, api service.APIClient) {
	// 获取局域网设备
	devices, err := api.Monitor().GetLanIP(ctx)
	if err != nil {
		fmt.Printf("  ❌ GetLanIP 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ GetLanIP 成功 (共 %d 个设备)\n", len(devices))
		for i, device := range devices {
			if i < 3 {
				fmt.Printf("     - %s (%s): %s\n", device.Hostname, device.Mac, device.IPAddr)
			}
		}
	}

	// 获取网络接口
	ifaces, err := api.Monitor().GetInterfaces(ctx)
	if err != nil {
		fmt.Printf("  ❌ GetInterfaces 失败: %v\n", err)
	} else {
		ifaceChecks := ifaces.GetIFaceCheck()
		fmt.Printf("  ✅ GetInterfaces 成功 (共 %d 个接口)\n", len(ifaceChecks))
		for i, iface := range ifaceChecks {
			if i < 3 {
				fmt.Printf("     - %s: %s\n", iface.Interface, iface.IPAddr)
			}
		}
	}

	// 获取系统监控数据
	system, err := api.Monitor().GetSystem(ctx)
	if err != nil {
		fmt.Printf("  ❌ GetSystem 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ GetSystem 成功 (共 %d 条数据)\n", len(system))
		if len(system) > 0 {
			data := system[0]
			fmt.Printf("     - CPU: %.1f%%\n", data.CPU)
			fmt.Printf("     - 内存使用: %d\n", data.MemoryUse)
		}
	}

	// 获取 ARP 表
	arp, err := api.Monitor().GetARP(ctx)
	if err != nil {
		fmt.Printf("  ❌ GetARP 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ GetARP 成功 (共 %d 条记录)\n", len(arp))
	}
}

func testNetworkService(ctx context.Context, api service.APIClient) {
	// 获取 WAN 配置
	wan, err := api.Network().GetWan(ctx)
	if err != nil {
		fmt.Printf("  ❌ GetWan 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ GetWan 成功 (共 %d 个WAN口)\n", len(wan))
		for i, w := range wan {
			if i < 3 {
				ip := w.DHCPIPAddr
				if ip == "" {
					ip = w.PPPoEIPAddr
				}
				fmt.Printf("     - %s: %s\n", w.Name, ip)
			}
		}
	}

	// 获取 LAN 配置
	lan, err := api.Network().GetLan(ctx)
	if err != nil {
		fmt.Printf("  ❌ GetLan 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ GetLan 成功 (共 %d 个LAN口)\n", len(lan))
		for i, l := range lan {
			if i < 3 {
				fmt.Printf("     - %s: %s\n", l.Name, l.IPMask)
			}
		}
	}

	// 获取 DHCP 配置
	dhcp, err := api.Network().GetDHCPD(ctx)
	if err != nil {
		fmt.Printf("  ❌ GetDHCPD 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ GetDHCPD 成功 (共 %d 个DHCP服务)\n", len(dhcp))
	}

	// 获取 DDNS 配置
	ddns, err := api.Network().GetDDNS(ctx)
	if err != nil {
		fmt.Printf("  ❌ GetDDNS 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ GetDDNS 成功 (共 %d 个DDNS)\n", len(ddns))
	}
}

func testFirewallService(ctx context.Context, api service.APIClient) {
	// 获取访问控制列表
	acl, err := api.Firewall().GetACL(ctx)
	if err != nil {
		fmt.Printf("  ❌ GetACL 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ GetACL 成功 (共 %d 条规则)\n", len(acl))
	}

	// 获取端口映射
	dnat, err := api.Firewall().GetDNAT(ctx)
	if err != nil {
		fmt.Printf("  ❌ GetDNAT 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ GetDNAT 成功 (共 %d 条映射)\n", len(dnat))
	}

	// 获取连接数限制
	connLimit, err := api.Firewall().GetConnLimit(ctx)
	if err != nil {
		fmt.Printf("  ❌ GetConnLimit 失败: %v\n", err)
	} else {
		fmt.Printf("  ✅ GetConnLimit 成功 (共 %d 条规则)\n", len(connLimit))
	}
}
