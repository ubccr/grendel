// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/spf13/cobra"
)

var (
	configureCmd = &cobra.Command{
		Use:   "configure",
		Short: "Configure BMC",
	}
)

func init() {
	bmcCmd.AddCommand(configureCmd)
}
