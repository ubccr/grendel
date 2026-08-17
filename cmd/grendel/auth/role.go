// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"github.com/spf13/cobra"
)

var (
	roleCmd = &cobra.Command{
		Use:   "role",
		Short: "Role commands",
	}
)

func init() {
	authCmd.AddCommand(roleCmd)
}
