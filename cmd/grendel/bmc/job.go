// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/spf13/cobra"
)

var (
	jobCmd = &cobra.Command{
		Use:   "job",
		Short: "BMC job commands",
	}
)

func init() {
	bmcCmd.AddCommand(jobCmd)
}
