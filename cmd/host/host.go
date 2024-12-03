// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package host

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/internal/logger"
)

var (
	tags    []string
	log     = logger.GetLogger("HOST")
	hostCmd = &cobra.Command{
		Use:   "host",
		Short: "Host commands",
		Long:  `Host commands`,
	}
)

func init() {
	hostCmd.PersistentFlags().StringSliceVarP(&tags, "tags", "t", []string{}, "filter by tags")
	cmd.Root.AddCommand(hostCmd)
}
