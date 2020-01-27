package bmc

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/korovkin/limiter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/client"
	"github.com/ubccr/grendel/cmd"
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
	limit := limiter.NewConcurrencyLimiter(viper.GetInt("bmc.fanout"))
	for _, host := range hostList {
		limit.Execute(func() {
			sysmgr, err := systemMgr(host)
			if err != nil {
				cmd.Log.WithFields(logrus.Fields{
					"err":  err,
					"name": host.Name,
					"ID":   host.ID,
				}).Error("Failed to connect to BMC")
				return
			}
			defer sysmgr.Logout()

			err = sysmgr.EnablePXE()
			if err != nil {
				cmd.Log.WithFields(logrus.Fields{
					"err":  err,
					"name": host.Name,
					"ID":   host.ID,
				}).Error("Failed to enabel PXE on next boot")
				return
			}

			if reboot {
				err = sysmgr.PowerCycle()
				if err != nil {
					cmd.Log.WithFields(logrus.Fields{
						"err":  err,
						"name": host.Name,
						"ID":   host.ID,
					}).Error("Failed to power cycle node")
					return
				}
			}

			fmt.Printf("%s: OK\n", host.Name)
		})

		time.Sleep(time.Duration(delay) * time.Second)
	}

	limit.Wait()

	return nil
}
