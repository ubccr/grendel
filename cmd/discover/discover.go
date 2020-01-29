package discover

import (
	"fmt"
	"net"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/firmware"
)

var (
	subnetStr     string
	noProvision   bool
	subnet        net.IP
	firmwareBuild firmware.Build
	discoverCmd   = &cobra.Command{
		Use:   "discover",
		Short: "Auto-discover commands",
		Long:  `Auto-discover commands`,
	}
)

func init() {
	discoverCmd.PersistentFlags().StringP("domain", "d", "", "domain name")
	viper.BindPFlag("discovery.domain", discoverCmd.PersistentFlags().Lookup("domain"))
	discoverCmd.PersistentFlags().String("firmware", "", "firmware")
	viper.BindPFlag("discovery.firmware", discoverCmd.PersistentFlags().Lookup("firmware"))

	discoverCmd.PersistentFlags().BoolVar(&noProvision, "disable-provision", false, "don't set host to provision")
	discoverCmd.PersistentFlags().StringVarP(&subnetStr, "subnet", "s", "", "subnet to use for auto ip assignment (/24)")
	discoverCmd.MarkFlagRequired("subnet")

	discoverCmd.PersistentPreRunE = func(command *cobra.Command, args []string) error {
		err := cmd.SetupLogging()
		if err != nil {
			return err
		}

		subnet = net.IPv4(0, 0, 0, 0)
		if subnetStr != "" {
			subnet = net.ParseIP(subnetStr)
			if subnet == nil || subnet.To4() == nil {
				return fmt.Errorf("Invalid IPv4 subnet address: %s", subnetStr)
			}
		}

		firmwareStr := viper.GetString("discovery.firmware")
		if firmwareStr != "" {
			firmwareBuild = firmware.NewFromString(firmwareStr)
			if firmwareBuild.IsNil() {
				return fmt.Errorf("Invalid firmware build: %s", firmwareStr)
			}
		}

		return nil
	}

	cmd.Root.AddCommand(discoverCmd)
}
