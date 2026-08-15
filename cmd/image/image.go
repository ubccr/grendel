// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package image

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
)

var (
	imageCmd = &cobra.Command{
		Use:     "image",
		Short:   "Boot Image commands",
		Aliases: []string{"images"},
	}
)

func init() {
	cmd.Root.AddCommand(imageCmd)
}
