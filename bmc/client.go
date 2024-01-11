package bmc

import (
	"github.com/gofiber/fiber/v2/log"
	"github.com/spf13/viper"
	"github.com/stmcginnis/gofish"
)

type Redfish struct {
	config  gofish.ClientConfig
	client  *gofish.APIClient
	service *gofish.Service
}

type System struct {
	Name           string   `json:"name"`
	BIOSVersion    string   `json:"bios_version"`
	SerialNumber   string   `json:"serial_number"`
	Manufacturer   string   `json:"manufacturer"`
	PowerStatus    string   `json:"power_status"`
	Health         string   `json:"health"`
	TotalMemory    float32  `json:"total_memory"`
	ProcessorCount int      `json:"processor_count"`
	BootNext       string   `json:"boot_next"`
	BootOrder      []string `json:"boot_order"`
}

func NewRedfishClient(ip string) (*Redfish, error) {
	user := viper.GetString("bmc.user")
	pass := viper.GetString("bmc.password")
	viper.SetDefault("bmc.insecure", true)
	insecure := viper.GetBool("bmc.insecure")

	endpoint := "https://" + ip

	config := gofish.ClientConfig{
		Endpoint: endpoint,
		Username: user,
		Password: pass,
		Insecure: insecure,
	}

	client, err := gofish.Connect(config)
	e, err := ParseRedfishError(err)
	if err != nil {
		return nil, err
	}
	// Try with default credentials
	if e.Code == "401" {
		config.Username = "root"
		config.Password = "calvin"
		client, err = gofish.Connect(config)
		if err != nil {
			log.Debug("default credentials failed")
			return nil, err
		}
	}

	return &Redfish{
		config:  config,
		client:  client,
		service: client.Service,
	}, nil
}
