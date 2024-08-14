package bmc

import (
	"github.com/stmcginnis/gofish"
)

type Redfish struct {
	config  gofish.ClientConfig
	client  *gofish.APIClient
	service *gofish.Service
}

type System struct {
	Name           string   `json:"name"`
	HostName       string   `json:"host_name"`
	BIOSVersion    string   `json:"bios_version"`
	SerialNumber   string   `json:"serial_number"`
	Manufacturer   string   `json:"manufacturer"`
	Model          string   `json:"model"`
	PowerStatus    string   `json:"power_status"`
	Health         string   `json:"health"`
	TotalMemory    float32  `json:"total_memory"`
	ProcessorCount int      `json:"processor_count"`
	BootNext       string   `json:"boot_next"`
	BootOrder      []string `json:"boot_order"`
}

func NewRedfishClient(ip, user, pass string, insecure bool) (*Redfish, error) {
	endpoint := "https://" + ip

	config := gofish.ClientConfig{
		Endpoint: endpoint,
		Username: user,
		Password: pass,
		Insecure: insecure,
	}

	client, err := gofish.Connect(config)
	if err != nil {
		e := ParseRedfishError(err)
		// Try with default credentials
		if e.Code == "401" {
			config.Username = "root"
			config.Password = "calvin"
			client, err = gofish.Connect(config)
			if err != nil {
				log.Debug("default credentials failed")
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &Redfish{
		config:  config,
		client:  client,
		service: client.GetService(),
	}, nil
}
