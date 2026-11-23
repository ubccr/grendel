// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	biosSetApplyTime string
	biosSetCmd       = ackCommand(&cobra.Command{
		Use:   "set <nodeset | all> <Attribute=Value>...",
		Short: "Stage BIOS attribute changes",
		Long:  "Stage BIOS attribute changes. Values are converted to the type each attribute holds, and a node is left unchanged if any value does not fit it. Use 'bmc bios describe' to list an attribute's allowed values",
		Args:  cobra.MinimumNArgs(2),
	}, func(command *cobra.Command, args []string) (*client.BmcResponseModelBmcAck, error) {
		attrs, err := parseAssignments(args[1:])
		if err != nil {
			return nil, err
		}

		req := &client.BmcSetBiosBody{Attributes: client.NewOptBmcSetBiosBodyAttributes(attrs)}
		if biosSetApplyTime != "" {
			req.ApplyTime = client.NewOptString(biosSetApplyTime)
		}

		return apiclient.API.POSTV1BmcBios(command.Context(), req, client.POSTV1BmcBiosParams(selected(args[0])))
	})
)

func init() {
	biosSetCmd.Flags().StringVar(&biosSetApplyTime, "apply-time", "", "when the BMC applies the change: Immediate, OnReset, AtMaintenanceWindowStart or InMaintenanceWindowOnReset. Defaults to the BMC's own default. Immediate can reboot the host")
	biosSetCmd.RegisterFlagCompletionFunc("apply-time", cobra.FixedCompletions([]string{"Immediate", "OnReset", "AtMaintenanceWindowStart", "InMaintenanceWindowOnReset"}, cobra.ShellCompDirectiveNoFileComp))

	biosCmd.AddCommand(biosSetCmd)
}
