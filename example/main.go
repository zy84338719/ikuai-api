// Example demonstrates how to use the iKuai SDK
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/service"
)

func main() {
	// Example 1: Create client and manually login
	fmt.Println("=== Example 1: Manual Login ===")
	manualLoginExample()

	// Example 2: Create client with auto login (recommended)
	fmt.Println("\n=== Example 2: Auto Login ===")
	autoLoginExample()

	// Example 3: Using services
	fmt.Println("\n=== Example 3: Using Services ===")
	servicesExample()

	// Example 4: Custom configuration
	fmt.Println("\n=== Example 4: Custom Configuration ===")
	customConfigExample()
}

func manualLoginExample() {
	// Replace with your router's address and credentials
	client := ikuaisdk.NewClient(
		"http://10.10.40.254",
		"admin",
		"password",
		ikuaisdk.WithTimeout(30*time.Second),
	)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Login(ctx); err != nil {
		log.Printf("Login failed: %v\n", err)
		return
	}

	fmt.Printf("Logged in successfully! iKuai version: %s\n", client.GetVersion())

	if err := client.Logout(ctx); err != nil {
		log.Printf("Logout failed: %v\n", err)
	}
}

func autoLoginExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create client with auto login
	client, err := ikuaisdk.NewClientWithLoginContext(
		ctx,
		"http://10.10.40.254",
		"admin",
		"password",
		ikuaisdk.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Printf("Failed to create client: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Printf("Connected to iKuai %s\n", client.GetVersion())
	fmt.Printf("Is logged in: %v\n", client.IsLoggedIn())
}

func servicesExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := ikuaisdk.NewClientWithLoginContext(
		ctx,
		"http://10.10.40.254",
		"admin",
		"password",
	)
	if err != nil {
		log.Printf("Failed to create client: %v\n", err)
		return
	}
	defer client.Close()

	// Create API client for accessing services
	api := service.NewAPIClient(client)

	// Get system information
	homepage, err := api.System().GetHomepage(ctx)
	if err != nil {
		log.Printf("Failed to get homepage: %v\n", err)
	} else {
		fmt.Printf("Hostname: %s\n", homepage.Hostname)
		fmt.Printf("Version: %s\n", homepage.VerInfo.Version)
		fmt.Printf("Uptime: %d seconds\n", homepage.Uptime)
		fmt.Printf("CPU: %v\n", homepage.CPU)
		fmt.Printf("Memory Used: %s\n", homepage.Memory.Used)
	}

	// Get LAN devices
	devices, err := api.Monitor().GetLanIP(ctx)
	if err != nil {
		log.Printf("Failed to get LAN devices: %v\n", err)
	} else {
		fmt.Printf("\nFound %d LAN devices:\n", len(devices))
		for i, device := range devices {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(devices)-5)
				break
			}
			hostname := device.Hostname
			if hostname == "" {
				hostname = "(unknown)"
			}
			fmt.Printf("  %s (%s): %s\n", hostname, device.Mac, device.IPAddr)
		}
	}

	// Get network interfaces
	ifaces, err := api.Monitor().GetInterfaces(ctx)
	if err != nil {
		log.Printf("Failed to get interfaces: %v\n", err)
	} else {
		fmt.Printf("\nNetwork interfaces:\n")
		for _, iface := range ifaces.GetIFaceStream() {
			fmt.Printf("  %s: %s (Up: %d, Down: %d)\n",
				iface.Interface, iface.Comment, iface.Upload, iface.Download)
		}
	}

	// Get WAN configuration
	wans, err := api.Network().GetWan(ctx)
	if err != nil {
		log.Printf("Failed to get WAN config: %v\n", err)
	} else {
		fmt.Printf("\nWAN configurations:\n")
		for _, wan := range wans {
			fmt.Printf("  %s: %s (Internet: %d)\n", wan.Name, wan.Comment, wan.Internet)
		}
	}
}

func customConfigExample() {
	// Create client with custom HTTP client for debugging
	// Note: This requires importing github.com/imroc/req/v3
	fmt.Println("Custom HTTP client configuration example:")
	fmt.Println(`
import "github.com/imroc/req/v3"

customReqClient := req.C().
    SetTimeout(60*time.Second).
    EnableInsecureSkipVerify().
    EnableDumpEachRequest()  // Enable request/response dump for debugging

client := ikuaisdk.NewClient(
    "http://192.168.1.1",
    "admin",
    "password",
    ikuaisdk.WithHTTPClient(customReqClient),
)
	`)

	// Example with cache enabled
	fmt.Println("\nExample with cache enabled:")
	fmt.Println(`
client, _ := ikuaisdk.NewClientWithLogin("http://192.168.1.1", "admin", "password")
client.EnableCache(5 * time.Minute)  // Cache responses for 5 minutes
defer client.DisableCache()
	`)
}
