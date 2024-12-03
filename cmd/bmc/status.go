// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/internal/bmc"
)

var (
	statusLong bool
	statusCmd  = &cobra.Command{
		Use:   "status",
		Short: "Check BMC status",
		Long:  `Check BMC status`,
		RunE: func(command *cobra.Command, args []string) error {
			return cmdStatus()
		},
	}
)

func init() {
	statusCmd.Flags().BoolVar(&statusLong, "long", false, "Display long format")
	bmcCmd.AddCommand(statusCmd)
}

func cmdStatus() error {
	job := bmc.NewJob()
	output, err := job.BmcStatus(hostList)
	if err != nil {
		return err
	}

	if statusLong {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "    ")

		err := enc.Encode(output)
		if err != nil {
			return err
		}
	} else {
		for _, o := range output {

			if !statusLong {
				fmt.Printf("%s\t%s\t%s\n", o.Name, o.PowerStatus, o.BIOSVersion)
				continue
			}
		}

	}

	return nil
}
