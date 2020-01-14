package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/korovkin/limiter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/bmc"
	"github.com/ubccr/grendel/client"
	"github.com/ubccr/grendel/nodeset"
	"github.com/urfave/cli"
)

func NewBMCNetbootCommand() cli.Command {
	return cli.Command{
		Name:        "netboot",
		Usage:       "Enable PXE and reboot host",
		Description: "Enabel PXE and reboot host",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "nodeset",
				Usage: "Set of nodes to netboot",
			},
			cli.StringFlag{
				Name:  "grendel-endpoint",
				Usage: "grendel endpoint url",
			},
			cli.StringFlag{
				Name:  "bmc-endpoint",
				Usage: "bmc endpoint url",
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
			cli.BoolFlag{
				Name:  "reboot",
				Usage: "Reboot nodes",
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
		Action: runNetboot,
	}
}

func runNetboot(c *cli.Context) error {
	if !c.IsSet("bmc-endpoint") && !c.IsSet("nodeset") {
		return errors.New("Either nodeset or bmc-endpoint is required")
	}

	if c.IsSet("bmc-endpoint") {
		return netbootUsingEndpoint(c.String("bmc-endpoint"), c.String("bmc-user"), c.String("bmc-pass"), c.Bool("ipmi"), c.Bool("reboot"))
	}

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

	limit := limiter.NewConcurrencyLimiter(c.Int("fanout"))
	for _, host := range hostList {
		limit.Execute(func() {
			bmcIntf := host.InterfaceBMC()
			if bmcIntf == nil {
				log.WithFields(log.Fields{
					"name": host.Name,
					"ID":   host.ID,
				}).Error("BMC interface not found")
				return
			}

			bmcAddress := bmcIntf.FQDN
			if bmcAddress == "" {
				bmcAddress = bmcIntf.IP.String()
			}

			if bmcAddress == "" {
				log.WithFields(log.Fields{
					"name": host.Name,
					"ID":   host.ID,
				}).Error("BMC address not set")
				return
			} else if !c.Bool("ipmi") {
				bmcAddress = fmt.Sprintf("https://%s", bmcAddress)
			}

			err = netbootUsingEndpoint(bmcAddress, bmcUsername, bmcPassword, c.Bool("ipmi"), c.Bool("reboot"))
			if err != nil {
				log.WithFields(log.Fields{
					"err":  err,
					"name": host.Name,
					"ID":   host.ID,
				}).Error("Failed to netboot host")
				return
			}

			fmt.Printf("%s: OK\n", host.Name)
		})

		time.Sleep(time.Duration(c.Int("delay")) * time.Second)
	}

	limit.Wait()

	return nil
}

func netbootUsingEndpoint(endpoint, user, pass string, useIPMI, reboot bool) error {
	var sysmgr bmc.SystemManager
	var err error

	if useIPMI {
		ipmi, err := bmc.NewIPMI(endpoint, user, pass, 623)
		if err != nil {
			return err
		}
		sysmgr = ipmi
	} else {
		redfish, err := bmc.NewRedfish(endpoint, user, pass, true)
		if err != nil {
			return err
		}
		sysmgr = redfish
	}

	defer sysmgr.Logout()

	err = sysmgr.EnablePXE()
	if err != nil {
		return err
	}

	if reboot {
		err = sysmgr.PowerCycle()
		if err != nil {
			return err
		}
	}

	return nil
}
