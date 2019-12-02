package cmd

import (
	"github.com/urfave/cli"

	"github.com/ubccr/grendel/model"
)

var DB model.Datastore

func NewServeCommand() cli.Command {
	return cli.Command{
		Name:        "serve",
		Usage:       "Start services",
		Description: "Start services",
		Subcommands: []cli.Command{
			NewDHCPCommand(),
			NewTFTPCommand(),
			NewAPICommand(),
			NewServeAllCommand(),
		},
	}
}
