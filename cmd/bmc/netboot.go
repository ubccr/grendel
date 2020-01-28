package bmc

import (
	"errors"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/client"
	"github.com/ubccr/grendel/nodeset"
)

var (
	reboot     bool
	netbootCmd = &cobra.Command{
		Use:   "netboot",
		Short: "Set hosts to PXE netboot",
		Long:  `Set hosts to PXE netboot`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			ns, err := nodeset.NewNodeSet(strings.Join(args, ","))
			if err != nil {
				return err
			}

			if ns.Len() == 0 {
				return errors.New("Node nodes in nodeset")
			}

			return runNetboot(ns)
		},
	}
)

func init() {
	netbootCmd.Flags().BoolVarP(&reboot, "reboot", "r", false, "Reboot nodes")
	bmcCmd.AddCommand(netbootCmd)
}

func runNetboot(ns *nodeset.NodeSet) error {
	gc, err := client.NewClient()
	if err != nil {
		return err
	}

	hostList, err := gc.FindHosts(ns)
	if err != nil {
		return err
	}

	if len(hostList) == 0 {
		return errors.New("No hosts found")
	}

	delay := viper.GetInt("bmc.delay")
	runner := NewJobRunner(viper.GetInt("bmc.fanout"))
	for _, host := range hostList {
		runner.RunNetBoot(host, reboot)
		time.Sleep(time.Duration(delay) * time.Second)
	}

	runner.Wait()

	return nil
}
