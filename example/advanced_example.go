// Package main demonstrates advanced usage of iKuai SDK
//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/service"
	"github.com/zy84338719/ikuai-api/utils"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== iKuai SDK 高级示例 ===")
	fmt.Println()

	// 示例1: 基本连接和系统监控
	demoBasicConnection(ctx)

	// 示例2: 防火墙管理
	demoFirewallManagement(ctx)

	// 示例3: 网络配置
	demoNetworkConfig(ctx)

	// 示例4: 工具函数
	demoUtilities()

	// 示例5: 错误处理和重试
	demoErrorHandling(ctx)
}

func demoBasicConnection(ctx context.Context) {
	fmt.Println("【示例1】基本连接和监控")
	fmt.Println("----------------------------------------")

	client, err := createClient()
	if err != nil {
		log.Printf("❌ 连接失败: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Printf("✅ 连接成功 - iKuai 版本: %s\n\n", client.GetVersion())

	// 获取系统信息
	api := service.NewAPIClient(client)

	homepage, err := api.System().GetHomepage(ctx)
	if err == nil {
		fmt.Printf("主机名: %s\n", homepage.Hostname)
		fmt.Printf("运行时间: %d 秒\n\n", homepage.Uptime)
	}

	// 获取局域网设备
	devices, err := api.Monitor().GetLanIP(ctx)
	if err == nil {
		fmt.Printf("局域网设备: %d 个\n", len(devices))
		if len(devices) > 0 {
			fmt.Printf("  - 第一个设备: %s (%s)\n\n", devices[0].Hostname, devices[0].IPAddr)
		}
	}
}

func demoFirewallManagement(ctx context.Context) {
	fmt.Println("【示例2】防火墙管理")
	fmt.Println("----------------------------------------")

	client, err := createClient()
	if err != nil {
		log.Printf("❌ 连接失败: %v\n", err)
		return
	}
	defer client.Close()

	api := service.NewAPIClient(client)

	// 查询 ACL 规则
	acls, err := api.Firewall().GetACL(ctx)
	if err != nil {
		log.Printf("❌ 获取 ACL 失败: %v\n", err)
	} else {
		fmt.Printf("ACL 规则数量: %d\n", len(acls))
	}

	// 查询端口映射
	dnats, err := api.Firewall().GetDNAT(ctx)
	if err != nil {
		log.Printf("❌ 获取 DNAT 失败: %v\n", err)
	} else {
		fmt.Printf("端口映射数量: %d\n\n", len(dnats))
	}

	// 示例: 添加规则（已注释，避免实际执行）
	/*
		// 添加 ACL 规则
		aclReq := &types.ACLAddRequest{
			TagName:  "test-rule",
			Enabled:  "yes",
			Comment:  "测试规则",
			SrcAddr:  "192.168.1.0/24",
			DstAddr:  "any",
			Protocol: "tcp",
			DstPort:  "80",
			Action:   "accept",
		}
		id, err := api.Firewall().AddACL(ctx, aclReq)
		if err != nil {
			log.Printf("添加失败: %v\n", err)
		} else {
			fmt.Printf("✅ 添加成功, ID: %d\n", id)
			// 删除规则
			api.Firewall().DelACL(ctx, id)
		}
	*/
}

func demoNetworkConfig(ctx context.Context) {
	fmt.Println("【示例3】网络配置管理")
	fmt.Println("----------------------------------------")

	client, err := createClient()
	if err != nil {
		log.Printf("❌ 连接失败: %v\n", err)
		return
	}
	defer client.Close()

	api := service.NewAPIClient(client)

	// 获取 WAN 配置
	wans, err := api.Network().GetWan(ctx)
	if err == nil {
		fmt.Printf("WAN 接口: %d 个\n", len(wans))
		for i, wan := range wans {
			if i < 2 {
				ip := wan.DHCPIPAddr
				if ip == "" {
					ip = wan.PPPoEIPAddr
				}
				fmt.Printf("  - %s: %s\n", wan.Name, ip)
			}
		}
		fmt.Println()
	}

	// 获取 DNS 静态解析
	dnsStatic, err := api.Network().GetDNSStatic(ctx)
	if err == nil {
		fmt.Printf("DNS 静态解析: %d 条\n", len(dnsStatic))
		for i, dns := range dnsStatic {
			if i < 3 {
				fmt.Printf("  - %s -> %s\n", dns.Domain, dns.IPAddr)
			}
		}
		fmt.Println()
	}

	// 获取静态路由
	routes, err := api.Network().GetRouteStatic(ctx)
	if err == nil {
		fmt.Printf("静态路由: %d 条\n\n", len(routes))
	}

	// 示例: 添加 DNS 静态解析（已注释）
	/*
		dnsReq := &types.DNSStaticAddRequest{
			TagName: "test-dns",
			Enabled: "yes",
			Comment: "测试",
			Domain:  "test.local",
			IPAddr:  "192.168.1.100",
		}
		id, err := api.Network().AddDNSStatic(ctx, dnsReq)
		if err != nil {
			log.Printf("添加失败: %v\n", err)
		} else {
			fmt.Printf("✅ 添加成功, ID: %d\n", id)
		}
	*/
}

func demoUtilities() {
	fmt.Println("【示例4】工具函数")
	fmt.Println("----------------------------------------")

	// IP 验证
	ips := []string{"192.168.1.1", "256.1.1.1", "10.0.0.1"}
	for _, ip := range ips {
		if utils.IsValidIPv4(ip) {
			fmt.Printf("✅ %s - 有效的 IPv4 地址\n", ip)
		} else {
			fmt.Printf("❌ %s - 无效的 IPv4 地址\n", ip)
		}
	}

	// MAC 验证
	macs := []string{"00:11:22:33:44:55", "invalid-mac"}
	for _, mac := range macs {
		if utils.IsValidMAC(mac) {
			fmt.Printf("✅ %s - 有效的 MAC 地址\n", mac)
		} else {
			fmt.Printf("❌ %s - 无效的 MAC 地址\n", mac)
		}
	}

	// 端口验证
	ports := []int{80, 443, 65535, 70000}
	for _, port := range ports {
		if utils.IsValidPort(port) {
			fmt.Printf("✅ 端口 %d - 有效\n", port)
		} else {
			fmt.Printf("❌ 端口 %d - 无效\n", port)
		}
	}
	fmt.Println()

	// 速率限制器
	fmt.Println("速率限制器示例 (10 req/s):")
	rateLimiter := utils.NewRateLimiter(10)
	defer rateLimiter.Stop()

	start := time.Now()
	for i := 0; i < 5; i++ {
		rateLimiter.Wait(context.Background())
		fmt.Printf("  请求 %d 已发送\n", i+1)
	}
	elapsed := time.Since(start)
	fmt.Printf("耗时: %v (预期约 400ms)\n\n", elapsed)
}

func demoErrorHandling(ctx context.Context) {
	fmt.Println("【示例5】错误处理和重试")
	fmt.Println("----------------------------------------")

	// 重试示例
	fmt.Println("重试机制示例:")
	attempts := 0
	err := utils.Retry(ctx, 3, 100*time.Millisecond, func() error {
		attempts++
		fmt.Printf("  尝试 %d...\n", attempts)
		if attempts < 2 {
			return fmt.Errorf("模拟临时错误")
		}
		fmt.Printf("  ✅ 第 %d 次尝试成功\n", attempts)
		return nil
	})
	if err != nil {
		log.Printf("❌ 重试失败: %v\n", err)
	}
	fmt.Println()

	// 指数退避重试
	fmt.Println("指数退避重试示例:")
	attempts = 0
	err = utils.RetryWithBackoff(ctx, 3, 50*time.Millisecond, func() error {
		attempts++
		delay := 50 * time.Millisecond * time.Duration(1<<(attempts-1))
		if delay > 200*time.Millisecond {
			delay = 200 * time.Millisecond
		}
		fmt.Printf("  尝试 %d (延迟 ~%v)...\n", attempts, delay)
		if attempts < 3 {
			return fmt.Errorf("模拟持续错误")
		}
		return nil
	})
	if err != nil {
		fmt.Printf("  ❌ 重试失败 (预期): %v\n\n", err)
	}

	// SDK 错误处理
	client, err := createClient()
	if err != nil {
		return
	}
	defer client.Close()

	fmt.Println("SDK 错误处理示例:")
	err = client.Call(ctx, "invalid_api", "show", nil, nil)
	if err != nil {
		if ikuaisdk.IsSDKError(err) {
			code := ikuaisdk.GetErrorCode(err)
			fmt.Printf("  SDK 错误码: %d\n", code)
			fmt.Printf("  错误信息: %v\n", err)
		} else {
			fmt.Printf("  其他错误: %v\n", err)
		}
	}
}

func createClient() (*ikuaisdk.Client, error) {
	// 注意: 实际使用时请从环境变量或配置文件读取
	addr := "192.168.1.1"
	username := "admin"
	password := "your-password"

	return ikuaisdk.NewClientWithLogin(
		addr,
		username,
		password,
		ikuaisdk.WithTimeout(30*time.Second),
	)
}
