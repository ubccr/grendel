// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
)

var (
	authCmd = &cobra.Command{
		Use:   "auth",
		Short: "Auth commands",
	}
)

func init() {
	cmd.Root.AddCommand(authCmd)
}
