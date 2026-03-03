package main

import (
	"context"
	"fmt"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/service"
)

func main() {
	client, err := ikuaisdk.NewClientWithLogin("http://192.168.1.1", "admin", "admin123")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	api := service.NewAPIClient(client)

	homepage, err := api.System().GetHomepage(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("iKuai Version: %s\n", homepage.VerInfo.Version)
	fmt.Printf("Hostname: %s\n", homepage.Hostname)
	fmt.Printf("Uptime: %d seconds\n", homepage.Uptime)

	devices, err := api.Monitor().GetLanIP(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d LAN devices\n", len(devices))
}
