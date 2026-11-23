// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var indicatorCmd = ackCommand(&cobra.Command{
	Use:       "indicator {on | off} {nodeset | all}",
	Aliases:   []string{"led", "locate"},
	Short:     "Turn the location indicator on or off",
	Long:      "Turn the location indicator (identify LED) on or off, so the nodes can be found in the rack",
	Args:      cobra.ExactArgs(2),
	ValidArgs: []string{"on", "off"},
}, func(command *cobra.Command, args []string) (*client.BmcResponseModelBmcAck, error) {
	if args[0] != "on" && args[0] != "off" {
		return nil, fmt.Errorf("invalid state %q, must be on or off", args[0])
	}

	req := &client.BmcIndicatorBody{State: client.NewOptString(args[0])}

	return apiclient.API.POSTV1BmcIndicator(command.Context(), req, client.POSTV1BmcIndicatorParams(selected(args[1])))
})

func init() {
	bmcCmd.AddCommand(indicatorCmd)
}
