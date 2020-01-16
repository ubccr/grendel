package cmd

import (
	"github.com/urfave/cli/v2"

	"github.com/ubccr/grendel/model"
)

var DB model.Datastore

func NewServeCommand() *cli.Command {
	return &cli.Command{
		Name:        "serve",
		Usage:       "Start services",
		Description: "Start services",
		Subcommands: []*cli.Command{
			NewDHCPCommand(),
			NewTFTPCommand(),
			NewDNSCommand(),
			NewAPICommand(),
			NewServeAllCommand(),
		},
	}
}
