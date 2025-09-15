// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/internal/logger"
)

var (
	log    = logger.GetLogger("CONFIG")
	cfgCmd = &cobra.Command{
		Use:   "config",
		Short: "Configuration commands",
	}
)

func init() {
	cmd.Root.AddCommand(cfgCmd)
}
