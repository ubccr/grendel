package cmd

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/spf13/viper"
	"github.com/ubccr/grendel/client"
	"github.com/ubccr/grendel/nodeset"
	"github.com/urfave/cli"
)

func NewHostShowCommand() cli.Command {
	return cli.Command{
		Name:        "show",
		Usage:       "Host show",
		Description: "Host show",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "nodeset",
				Usage: "Set of nodes to show",
			},
			cli.StringFlag{
				Name:  "grendel-endpoint",
				Usage: "grendel endpoint url",
			},
		},
		Action: func(c *cli.Context) error {
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
				return errors.New("No nodes in nodeset")
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
			if err := enc.Encode(hostList); err != nil {
				return err
			}

			return nil
		},
	}
}
