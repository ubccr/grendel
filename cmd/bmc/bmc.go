// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/internal/logger"
)

var (
	tags   []string
	log    = logger.GetLogger("BMC")
	bmcCmd = &cobra.Command{
		Use:   "bmc",
		Short: "BMC commands",
	}
)

func init() {
	bmcCmd.PersistentFlags().StringSliceVarP(&tags, "tags", "t", []string{}, "Filter by tags")

	cmd.Root.AddCommand(bmcCmd)
}
