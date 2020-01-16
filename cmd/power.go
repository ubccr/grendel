package cmd

import (
	"github.com/urfave/cli/v2"
)

func NewBMCCommand() *cli.Command {
	return &cli.Command{
		Name:        "bmc",
		Usage:       "BMC commands",
		Description: "BMC commands",
		Subcommands: []*cli.Command{
			NewBMCNetbootCommand(),
			NewBMCStatusCommand(),
		},
	}
}
