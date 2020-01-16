package cmd

import (
	"github.com/urfave/cli/v2"
)

func NewHostCommand() *cli.Command {
	return &cli.Command{
		Name:        "host",
		Usage:       "Host commands",
		Description: "Host commands",
		Subcommands: []*cli.Command{
			NewHostDiscoverCommand(),
			NewHostShowCommand(),
			NewBMCCommand(),
		},
	}
}
