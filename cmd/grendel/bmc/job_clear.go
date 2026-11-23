// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var jobClearCmd = ackCommand(&cobra.Command{
	Use:   "clear <nodeset | all> [JID]...",
	Short: "Clear all jobs or by JID",
	Long:  `Clear all jobs or by JID. Clears the whole queue when no JID is given`,
	Args:  cobra.MinimumNArgs(1),
}, func(command *cobra.Command, args []string) (*client.BmcResponseModelBmcAck, error) {
	req := &client.BmcClearJobsBody{}
	if jids := args[1:]; len(jids) > 0 {
		req.Ids = client.NewOptNilStringArray(jids)
	} else {
		req.All = client.NewOptBool(true)
	}

	return apiclient.API.POSTV1BmcJobsClear(command.Context(), req, client.POSTV1BmcJobsClearParams(selected(args[0])))
})

func init() {
	jobCmd.AddCommand(jobClearCmd)
}
