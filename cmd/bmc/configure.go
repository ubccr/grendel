// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/internal/bmc"
)

var (
	configureCmd = &cobra.Command{
		Use:   "configure",
		Short: "Set iDRAC to Auto configure",
		Long:  `Set iDRAC to Auto configure`,
		RunE: func(command *cobra.Command, args []string) error {
			return runConf()
		},
	}
)

func init() {
	bmcCmd.AddCommand(configureCmd)
}

func runConf() error {
	job := bmc.NewJob()

	output, err := job.BmcAutoConfigure(hostList)
	if err != nil {
		return err
	}

	bmc.PrintStatusCli(output)

	return nil
}
