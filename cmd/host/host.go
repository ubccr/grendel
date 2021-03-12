// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package host

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/logger"
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
