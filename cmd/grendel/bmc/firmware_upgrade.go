// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	firmwareUpgradeFilter  []string
	firmwareUpgradeOptions *shared.TableOptions
	firmwareUpgradeColumns = []shared.Column{
		{Name: "Node"},
		{Name: "State"},
		{Name: "Message"},
	}
	firmwareUpgradeApplyUpdate       bool
	firmwareUpgradeCatalogFile       string
	firmwareUpgradeIpAddress         string
	firmwareUpgradeIgnoreCertWarning bool
	firmwareUpgradeRebootNeeded      bool
	firmwareUpgradeClearJobs         bool
	firmwareUpgradeShareName         string
	firmwareUpgradeShareType         string
	firmwareUpgradeCmd               = &cobra.Command{
		Use:   "upgrade <nodeset | all>",
		Short: "Upgrade firmware on Dell servers",
		Long:  "Upgrade firmware on Dell servers. Without --apply-update this only compares the installed firmware against the catalog, waiting for the comparison to finish, and bmc firmware check then lists the available updates",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			req := &client.BmcInstallFirmwareBody{
				ApplyUpdate:       client.NewOptBool(firmwareUpgradeApplyUpdate),
				CatalogFile:       client.NewOptString(firmwareUpgradeCatalogFile),
				IPAddress:         client.NewOptString(firmwareUpgradeIpAddress),
				IgnoreCertWarning: client.NewOptBool(firmwareUpgradeIgnoreCertWarning),
				Reboot:            client.NewOptBool(firmwareUpgradeRebootNeeded),
				ShareName:         client.NewOptString(firmwareUpgradeShareName),
				ShareType:         client.NewOptString(firmwareUpgradeShareType),
				ClearJobs:         client.NewOptBool(firmwareUpgradeClearJobs),
			}

			nodeset, all := selector(args[0])
			params := client.POSTV1BmcFirmwareInstallParams{
				Nodeset: nodeset,
				All:     all,
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}

			res, err := apiclient.API.POSTV1BmcFirmwareInstall(command.Context(), req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(firmwareUpgradeColumns, firmwareUpgradeFilter).Options(*firmwareUpgradeOptions).Empty("-")
			for _, r := range res.Results.Value {
				result := r.Value
				t.AppendRow(
					result.Node.Value,
					colorState(result.Success.Value),
					result.Message.Value,
				)
			}
			t.Print()

			s := res.Summary.Value
			return summaryLine(s.Total.Value, s.Succeeded.Value, s.Failed.Value, s.FailedNodes.Value, s.DurationMs.Value)
		},
	}
)

func init() {
	firmwareUpgradeOptions = shared.RegisterTableFlags(firmwareUpgradeCmd)
	firmwareUpgradeCmd.Flags().StringSliceVar(&firmwareUpgradeFilter, "filter", nil, "filter table columns by name")
	firmwareUpgradeCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(firmwareUpgradeColumns))

	firmwareUpgradeCmd.Flags().BoolVarP(&firmwareUpgradeApplyUpdate, "apply-update", "a", false, "By default only check for updates, do not queue them. Pass this flag to apply available updates.")
	firmwareUpgradeCmd.Flags().StringVar(&firmwareUpgradeCatalogFile, "catalog-file", "", "Update catalog name. Defaults to Catalog.xml")
	firmwareUpgradeCmd.Flags().StringVarP(&firmwareUpgradeIpAddress, "ip-address", "i", "downloads.dell.com", "IP or Domain name of share")
	firmwareUpgradeCmd.Flags().BoolVar(&firmwareUpgradeIgnoreCertWarning, "ignore-cert-warning", true, "Pass this flag to ignore invalid certs")
	firmwareUpgradeCmd.Flags().BoolVarP(&firmwareUpgradeRebootNeeded, "reboot", "r", false, "Reboot arg will immediately reboot the node when needed")
	firmwareUpgradeCmd.Flags().BoolVar(&firmwareUpgradeClearJobs, "clear-jobs", false, "Clear all jobs in the job queue before upgrading. apply-update must be true")
	firmwareUpgradeCmd.Flags().StringVar(&firmwareUpgradeShareName, "share-name", "", "Path of the catalog's directory on the share, or the NFS export")
	firmwareUpgradeCmd.Flags().StringVar(&firmwareUpgradeShareType, "share-type", "HTTPS", "Valid options: HTTPS, HTTP or NFS")
	firmwareUpgradeCmd.MarkFlagRequired("share-name")

	firmwareCmd.AddCommand(firmwareUpgradeCmd)
}
