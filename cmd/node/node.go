// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/internal/logger"
)

var (
	tags    []string
	log     = logger.GetLogger("NODE")
	nodeCmd = &cobra.Command{
		Use:   "node",
		Short: "Node commands",
		Long:  `Node commands`,
	}
)

func init() {
	nodeCmd.PersistentFlags().StringSliceVarP(&tags, "tags", "t", []string{}, "filter by tags")
	cmd.Root.AddCommand(nodeCmd)
}
