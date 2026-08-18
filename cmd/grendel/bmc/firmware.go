// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var (
	firmwareCmd = &cobra.Command{
		Use:   "firmware",
		Short: "BMC Firmware commands",
	}
)

func colorVersion(v1, v2 string) string {
	if v1 != v2 {
		return text.FgHiRed.Sprint(v2)
	}
	return text.FgHiGreen.Sprint(v1)
}

func init() {
	bmcCmd.AddCommand(firmwareCmd)
}
