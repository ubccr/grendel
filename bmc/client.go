package bmc

import (
	"github.com/spf13/viper"
	"github.com/stmcginnis/gofish"
)

func NewClient(ip string) (*Redfish2, error) {

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

	// Try with default credentials
	// if err != nil && err.Error() == "Invalid username or password" {
	// client, err := gofish.ConnectDefault(endpoint)

	// }

	if err != nil {
		return nil, err
	}

	return &Redfish2{config: config, client: client, service: client.Service}, nil
}
