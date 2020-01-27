package bmc

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/korovkin/limiter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/bmc"
	"github.com/ubccr/grendel/client"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/nodeset"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Check BMC status",
		Long:  `Check BMC status`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			ns, err := nodeset.NewNodeSet(strings.Join(args, ","))
			if err != nil {
				return err
			}

			if ns.Len() == 0 {
				return errors.New("Node nodes in nodeset")
			}

			return runStatus(ns)
		},
	}
)

func init() {
	bmcCmd.AddCommand(statusCmd)
}

func runStatus(ns *nodeset.NodeSet) error {
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

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")

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

			system, err := sysmgr.GetSystem()
			if err != nil {
				cmd.Log.WithFields(logrus.Fields{
					"err":  err,
					"name": host.Name,
					"ID":   host.ID,
				}).Error("Failed to fetch system info from BMC")
				return
			}

			if system.Name == "" {
				system.Name = host.Name
			}

			rec := make(map[string]*bmc.System, 1)
			rec[host.Name] = system

			if err := enc.Encode(rec); err != nil {
				cmd.Log.WithFields(logrus.Fields{
					"err":  err,
					"name": host.Name,
					"ID":   host.ID,
				}).Error("Failed to encode json")
			}
		})

		time.Sleep(time.Duration(delay) * time.Second)
	}

	limit.Wait()

	return nil
}
