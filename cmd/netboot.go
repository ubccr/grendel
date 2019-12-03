package cmd

import (
	"github.com/ubccr/grendel/bmc"
	"github.com/urfave/cli"
)

func NewNetbootCommand() cli.Command {
	return cli.Command{
		Name:        "netboot",
		Usage:       "Enable PXE and reboot host",
		Description: "Enabel PXE and reboot host",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "mac",
				Usage: "mac address",
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
			cli.IntFlag{
				Name:  "ipmi-port",
				Value: 623,
				Usage: "IPMI port",
			},
			cli.BoolFlag{
				Name:  "ipmi",
				Usage: "Use ipmi instead of redfish",
			},
		},
		Action: runNetboot,
	}
}

func runNetboot(c *cli.Context) error {
	var sysmgr bmc.SystemManager
	var err error

	if c.Bool("ipmi") {
		ipmi, err := bmc.NewIPMI(c.String("bmc-endpoint"), c.String("bmc-user"), c.String("bmc-pass"), c.Int("ipmi-port"))
		if err != nil {
			return err
		}
		sysmgr = ipmi
	} else {
		redfish, err := bmc.NewRedfish(c.String("bmc-endpoint"), c.String("bmc-user"), c.String("bmc-pass"), true)
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

	err = sysmgr.PowerCycle()
	if err != nil {
		return err
	}

	return nil
}
