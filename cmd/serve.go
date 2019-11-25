package cmd

import (
	"github.com/urfave/cli"
)

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
