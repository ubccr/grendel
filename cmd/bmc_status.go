package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ubccr/grendel/api"
	"github.com/ubccr/grendel/client"
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
				Name:     "grendel-endpoint",
				Required: true,
				Usage:    "grendel endpoint url",
			},
			cli.IntFlag{
				Name:  "delay",
				Value: 0,
				Usage: "delay",
			},
		},
		Action: runBMCStatus,
	}
}

func runBMCStatus(c *cli.Context) error {
	grendelEndpoint := c.String("grendel-endpoint")

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
		Nodeset: ns,
		Delay:   c.Int("delay"),
	}

	res, err := gc.BMCStatus(params)
	if err != nil {
		return err
	}

	for host, system := range res {
		fmt.Printf("%s\n", strings.Join([]string{
			host,
			system.PowerStatus,
			system.BIOSVersion,
			system.SerialNumber,
			system.Health,
			strings.Join(system.BootOrder, ","),
			system.BootNext,
		}, "\t"))
	}

	return nil
}
