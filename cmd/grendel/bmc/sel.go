// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/spf13/cobra"
)

var (
	selCmd = &cobra.Command{
		Use:   "sel",
		Short: "BMC system event log commands",
	}
)

func init() {
	bmcCmd.AddCommand(selCmd)
}
