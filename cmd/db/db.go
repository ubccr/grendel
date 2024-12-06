// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package db

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/internal/logger"
)

var (
	tags  []string
	log   = logger.GetLogger("DB")
	dbCmd = &cobra.Command{
		Use:   "db",
		Short: "Database commands",
		Long:  `Database commands`,
	}
)

func init() {
	cmd.Root.AddCommand(dbCmd)
}
