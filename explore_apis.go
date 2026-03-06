//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"time"

	ikuaisdk "github.com/zy84338719/ikuai-api"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== 探索 iKuai API 接口 ===")
	fmt.Println()

	// 创建客户端并登录
	client, err := ikuaisdk.NewClientWithLogin(
		"http://10.10.30.254",
		"zhangyi",
		"REDACTED",
		ikuaisdk.WithTimeout(30*time.Second),
	)
	if err != nil {
		fmt.Printf("❌ 登录失败: %v\n", err)
		return
	}
	defer client.Close()
	fmt.Printf("✅ 登录成功！\n\n")

	// 探索可能的 funcName
	funcNames := []string{
		// 已知的
		"monitor_lanip",
		"monitor_iface",
		"monitor_system",
		"arp",
		"homepage",
		"upgrade",
		"backup",
		"webuser",
		"wan",
		"lan",
		"vlan",
		"ipv6",
		"iptv",
		"ddns",
		"dhcpd",
		"dhcp_static",
		"dhcp_lease",
		"acl",
		"dnat",
		"conn_limit",
		"domain_group",
		"custom_isp",
		"stream_domain",
		"pptp_client",
		"l2tp_client",
		"syslog",
		"syslog_wan_pppoe",
		"syslog_dhcpd",
		"syslog_arp",
		"syslog_ddns",
		"syslog_webadmin",
		"syslog_sysevent",
		"docker_image",
		"docker_container",
		"docker_network",
		"docker_compose",
		"qemu",
		"upnp",

		// 新增探索
		"dns",
		"dns_forward",
		"dns_static",
		"route_static",
		"route_policy",
		"flow_control",
		"bandwidth",
		"qos",
		"upnp_client",
		"vpn_server",
		"ipsec",
		"openvpn",
		"wireguard",
		"user",
		"group",
		"schedule",
		"time",
		"ntp",
		"logo",
		"theme",
		"language",
		"network_tool",
		"ping",
		"traceroute",
		"nslookup",
		"tcpdump",
		"system_tool",
		"reboot",
		"reset",
		"upgrade_online",
		"upgrade_offline",
		"backup_restore",
		"backup_download",
		"license",
		"about",
	}

	// 尝试每个 funcName
	for _, funcName := range funcNames {
		testAPI(client, ctx, funcName)
	}

	fmt.Println("\n=== 探索完成 ===")
}

func testAPI(client *ikuaisdk.Client, ctx context.Context, funcName string) {
	var result map[string]interface{}
	err := client.Call(ctx, funcName, "show", nil, &result)

	if err == nil {
		fmt.Printf("✅ %s - 可用\n", funcName)
	} else {
		// 检查是否是真正的错误还是只是没有数据
		if ikuaisdk.IsSDKError(err) {
			code := ikuaisdk.GetErrorCode(err)
			if code != ikuaisdk.ErrCodeRequestFailed {
				// 其他错误，可能接口存在但权限不足等
				fmt.Printf("⚠️  %s - 存在但出错: %v\n", funcName, err)
			}
		}
		// 不打印失败的，避免输出太多
	}
}
