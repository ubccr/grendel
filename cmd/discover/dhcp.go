package discover

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/dhcp"
)

var (
	dhcpCmd = &cobra.Command{
		Use:   "dhcp",
		Short: "Auto-discover hosts from dhcp",
		Long:  `Auto-discover hosts from dhcp`,
		RunE: func(command *cobra.Command, args []string) error {
			netmask := net.IPv4Mask(255, 255, 255, 0)
			subnet := net.ParseIP(subnetStr)
			if subnet == nil || subnet.To4() == nil {
				return fmt.Errorf("Invalid IPv4 subnet address: %s", subnetStr)
			}

			if len(args) == 0 {
				return errors.New("Please provide a nodeset")
			}

			return dhcp.RunDiscovery(viper.GetString("discovery.listen"), strings.Join(args, ","), subnet, netmask, nil)
		},
	}
)

func init() {
	switchCmd.Flags().StringP("listen", "l", "0.0.0.0:67", "address to run discovery DHCP server")
	viper.BindPFlag("discovery.listen", switchCmd.Flags().Lookup("listen"))

	discoverCmd.AddCommand(dhcpCmd)
}
