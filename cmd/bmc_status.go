package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/korovkin/limiter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/bmc"
	"github.com/ubccr/grendel/client"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
	"github.com/urfave/cli"
)

func NewBMCStatusCommand() cli.Command {
	return cli.Command{
		Name:        "status",
		Usage:       "Show bmc status",
		Description: "Show bmc status",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:     "nodeset",
				Required: true,
				Usage:    "Set of nodes to netboot",
			},
			cli.StringFlag{
				Name:  "grendel-endpoint",
				Usage: "grendel endpoint url",
			},
			cli.StringFlag{
				Name:  "bmc-user",
				Usage: "BMC Username",
			},
			cli.StringFlag{
				Name:  "bmc-pass",
				Usage: "BMC Password",
			},
			cli.BoolFlag{
				Name:  "ipmi",
				Usage: "Use ipmi instead of redfish",
			},
			cli.IntFlag{
				Name:  "delay",
				Value: 0,
				Usage: "delay",
			},
			cli.IntFlag{
				Name:  "fanout",
				Value: 1,
				Usage: "fanout",
			},
		},
		Action: runBMCStatus,
	}
}

func runBMCStatus(c *cli.Context) error {
	grendelEndpoint := viper.GetString("grendel_endpoint")
	if c.IsSet("grendel-endpoint") {
		grendelEndpoint = c.String("grendel-endpoint")
	}

	if grendelEndpoint == "" {
		return errors.New("Please set grendel-endpoint")
	}

	bmcUsername := viper.GetString("bmc_user")
	if c.IsSet("bmc-user") {
		bmcUsername = c.String("bmc-user")
	}

	bmcPassword := viper.GetString("bmc_pass")
	if c.IsSet("bmc-pass") {
		bmcPassword = c.String("bmc-pass")
	}

	if bmcUsername == "" || bmcPassword == "" {
		return errors.New("Please set bmc_user and bmc_password")
	}

	ns, err := nodeset.NewNodeSet(c.String("nodeset"))
	if err != nil {
		return err
	}

	if ns.Len() == 0 {
		return errors.New("Node nodes in nodeset")
	}

	gc, err := client.NewClient(grendelEndpoint, "", "", "", true)
	if err != nil {
		return err
	}

	hostList, err := gc.HostFind(ns)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")

	limit := limiter.NewConcurrencyLimiter(c.Int("fanout"))
	for _, host := range hostList {
		limit.Execute(func() {
			sysmgr, err := systemMgr(host, bmcUsername, bmcPassword, c.Bool("ipmi"))
			if err != nil {
				log.WithFields(log.Fields{
					"err":  err,
					"name": host.Name,
					"ID":   host.ID,
				}).Error("Failed to connect to BMC")
				return
			}
			defer sysmgr.Logout()

			system, err := sysmgr.GetSystem()
			if err != nil {
				log.WithFields(log.Fields{
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
				log.WithFields(log.Fields{
					"err":  err,
					"name": host.Name,
					"ID":   host.ID,
				}).Error("Failed to encode json")
			}
		})

		time.Sleep(time.Duration(c.Int("delay")) * time.Second)
	}

	limit.Wait()

	return nil
}

func systemMgr(host *model.Host, bmcUsername, bmcPass string, useIPMI bool) (bmc.SystemManager, error) {
	bmcIntf := host.InterfaceBMC()
	if bmcIntf == nil {
		return nil, errors.New("BMC interface not found")
	}

	bmcAddress := bmcIntf.FQDN
	if bmcAddress == "" {
		bmcAddress = bmcIntf.IP.String()
	}

	if bmcAddress == "" {
		return nil, errors.New("BMC address not set")
	}

	if useIPMI {
		ipmi, err := bmc.NewIPMI(bmcAddress, bmcUsername, bmcPass, 623)
		if err != nil {
			return nil, err
		}

		return ipmi, nil
	}

	redfish, err := bmc.NewRedfish(fmt.Sprintf("https://%s", bmcAddress), bmcUsername, bmcPass, true)
	if err != nil {
		return nil, err
	}

	return redfish, nil
}
