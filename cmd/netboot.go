package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"github.com/ubccr/grendel/api"
	"github.com/ubccr/grendel/bmc"
	"github.com/ubccr/grendel/client"
	"github.com/ubccr/grendel/nodeset"
	"github.com/urfave/cli"
)

func NewNetbootCommand() cli.Command {
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

	params := api.NetbootParams{
		Username: c.String("bmc-user"),
		Password: c.String("bmc-pass"),
		IPMI:     c.Bool("ipmi"),
		Reboot:   c.Bool("reboot"),
		Nodeset:  ns,
		Delay:    c.Int("delay"),
	}

	res, err := gc.Netboot(params)
	if err != nil {
		return err
	}

	for host, err := range res {
		fmt.Printf("%s: %s\n", host, err)
	}

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
		defer redfish.Logout()
		sysmgr = redfish
	}

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
