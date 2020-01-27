package bmc

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/cmd"
)

var (
	bmcUser     string
	bmcPassword string
	useIPMI     bool
	delay       int
	fanout      int
	bmcCmd      = &cobra.Command{
		Use:   "bmc",
		Short: "Query BMC devices",
		Long:  `Query BMC devices`,
	}
)

func init() {
	bmcCmd.PersistentFlags().String("user", "", "bmc user name")
	viper.BindPFlag("bmc.user", bmcCmd.PersistentFlags().Lookup("user"))
	bmcCmd.PersistentFlags().String("password", "", "bmc password")
	viper.BindPFlag("bmc.password", bmcCmd.PersistentFlags().Lookup("password"))
	bmcCmd.PersistentFlags().Int("delay", 0, "delay")
	viper.BindPFlag("bmc.delay", bmcCmd.PersistentFlags().Lookup("delay"))
	bmcCmd.PersistentFlags().Int("fanout", 1, "fanout")
	viper.BindPFlag("bmc.fanout", bmcCmd.PersistentFlags().Lookup("fanout"))
	bmcCmd.PersistentFlags().Bool("ipmi", false, "Use ipmi instead of redfish")
	viper.BindPFlag("bmc.ipmi", bmcCmd.PersistentFlags().Lookup("ipmi"))

	bmcCmd.PersistentPreRunE = func(command *cobra.Command, args []string) error {
		err := cmd.SetupLogging()
		if err != nil {
			return err
		}

		bmcUser = viper.GetString("bmc.user")
		if bmcUser == "" {
			return errors.New("please set bmc user")
		}
		bmcPassword = viper.GetString("bmc.password")
		if bmcPassword == "" {
			return errors.New("please set bmc password")
		}

		useIPMI = viper.GetBool("bmc.ipmi")
		delay = viper.GetInt("bmc.delay")
		fanout = viper.GetInt("bmc.fanout")

		return nil
	}

	cmd.Root.AddCommand(bmcCmd)
}
