// Example program demonstrating the v4 SDK.
//
// Usage:
//
//	IKUAI_BASE_URL=https://192.168.1.1 \
//	IKUAI_TOKEN=deadbeefcafebabe1234567890abcdef \
//	go run ./example
//
// The program lists online clients, prints the system overview, and
// (in dry-run mode) shows the request it would send to create a new
// auth user without actually doing so.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	ikuaiapi "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/service"
)

func main() {
	baseURL := os.Getenv("IKUAI_BASE_URL")
	token := os.Getenv("IKUAI_TOKEN")
	if baseURL == "" || token == "" {
		log.Fatal("set IKUAI_BASE_URL and IKUAI_TOKEN")
	}

	client, err := ikuaiapi.NewClient(baseURL,
		ikuaiapi.WithToken(token),
		ikuaiapi.WithTimeout(15*time.Second),
		ikuaiapi.WithLogger(func(format string, args ...any) {
			log.Printf("[ikuai] "+format, args...)
		}),
	)
	if err != nil {
		log.Fatalf("new client: %v", err)
	}
	defer client.Close()

	if err := ikuaiapi.ValidateToken(token); err != nil {
		log.Fatalf("token: %v", err)
	}

	api := service.NewAPIClient(client)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if os.Getenv("IKUAI_DRY_RUN") == "1" {
		// Replace client with a dry-run variant; no router traffic.
		dryClient, _ := ikuaiapi.NewClient(baseURL, ikuaiapi.WithToken(token), ikuaiapi.WithDryRun(true))
		api = service.NewAPIClient(dryClient)
		fmt.Println("# IKUAI_DRY_RUN=1 — printing requests without contacting the router")
	}

	// 1. System overview (single GET).
	system, err := api.Monitoring().GetMonitoringSystem(ctx)
	if err != nil {
		log.Fatalf("monitoring/system: %v", err)
	}
	fmt.Println("system overview:")
	prettyPrint(system)

	// 2. Online clients (single GET).
	clients, err := api.Monitoring().GetMonitoringClientsOnline(ctx)
	if err != nil {
		log.Fatalf("monitoring/clients-online: %v", err)
	}
	fmt.Println("\nclients online:")
	prettyPrint(clients)

	// 2b. Paginated list example: log/arp.
	arp, err := api.Log().ListLogArp(ctx,
		&service.LogArpListOptions{Page: 1, PageSize: 50, Order: "desc", OrderBy: "id"})
	if err != nil {
		log.Fatalf("log/arp: %v", err)
	}
	fmt.Println("\nlog/arp (first page, desc):")
	prettyPrint(arp)
	if err != nil {
		log.Fatalf("monitoring/clients-online: %v", err)
	}
	fmt.Println("\nclients (first page):")
	prettyPrint(clients)

	// 3. Auth users (paginated list).
	users, err := api.Auth().ListAuthUsers(ctx,
		&service.AuthUsersListOptions{Page: 1, PageSize: 50})
	if err != nil {
		log.Fatalf("auth/users: %v", err)
	}
	fmt.Println("\nauth users (first page):")
	prettyPrint(users)

	// 4. Network DHCP services (list).
	dhcp, err := api.Network().ListNetworkDhcpServices(ctx,
		&service.NetworkDhcpServicesListOptions{Page: 1, PageSize: 50})
	if err != nil {
		log.Fatalf("network/dhcp/services: %v", err)
	}
	fmt.Println("\ndhcp services (first page):")
	prettyPrint(dhcp)

	// 5. Catalog-driven call: get a single endpoint by (group, name).
	raw, err := api.Call(ctx, "interfaces", "wan-config", "GET", nil, nil)
	if err != nil {
		log.Fatalf("interfaces/wan-config: %v", err)
	}
	fmt.Println("\ninterfaces/wan-config:")
	prettyPrint(raw)

	// 6. Create dry-run (no router traffic) — only meaningful in dry-run mode.
	preview, err := api.Auth().CreateAuthUsers(ctx, map[string]any{
		"username": "alice",
		"password": "demo",
		"enabled":  "yes",
	})
	if err != nil {
		log.Fatalf("create auth user: %v", err)
	}
	fmt.Println("\ncreate preview (id=0 in dry-run):")
	fmt.Printf("  %+v\n", preview)
}

func prettyPrint(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}
