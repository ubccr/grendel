package cmd

import (
	certstrap "github.com/square/certstrap/cmd"
	"github.com/square/certstrap/depot"
	"github.com/urfave/cli"
)

func NewCertsCommand() cli.Command {
	return cli.Command{
		Name:        "certs",
		Usage:       "Certificate Authority Operations",
		Description: "Certificate Authority Operations",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "depot-path",
				Value: depot.DefaultFileDepotDir,
				Usage: "Location to store certificates, keys and other files.",
			},
		},
		Subcommands: []cli.Command{
			certstrap.NewInitCommand(),
			certstrap.NewCertRequestCommand(),
			certstrap.NewSignCommand(),
			certstrap.NewRevokeCommand(),
		},
		Before: func(c *cli.Context) error {
			certstrap.InitDepot(c.String("depot-path"))
			return nil
		},
	}
}
