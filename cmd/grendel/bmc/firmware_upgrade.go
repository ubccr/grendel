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
		{Name: "Status"},
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
		Use:   "upgrade <nodeset>",
		Short: "Upgrade firmware on Dell servers",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset := args[0]

			req := client.BmcDellInstallFromRepoRequest{
				ApplyUpdate:       client.NewOptBool(firmwareUpgradeApplyUpdate),
				CatalogFile:       client.NewOptString(firmwareUpgradeCatalogFile),
				IPAddress:         client.NewOptString(firmwareUpgradeIpAddress),
				IgnoreCertWarning: client.NewOptBool(firmwareUpgradeIgnoreCertWarning),
				RebootNeeded:      client.NewOptBool(firmwareUpgradeRebootNeeded),
				ShareName:         client.NewOptString(firmwareUpgradeShareName),
				ShareType:         client.NewOptString(firmwareUpgradeShareType),
				ClearJobQueue:     client.NewOptBool(firmwareUpgradeClearJobs),
			}
			params := client.POSTV1BmcUpgradeDellInstallfromrepoParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			jobs, err := apiclient.API.POSTV1BmcUpgradeDellInstallfromrepo(command.Context(), &req, params)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := shared.NewTable(firmwareUpgradeColumns, firmwareUpgradeFilter).Options(*firmwareUpgradeOptions).Empty("-")

			for _, job := range jobs {
				t.AppendRow(
					job.Host.Value,
					job.Status.Value,
					job.Msg.Value,
				)
			}

			t.Print()
			return nil
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
	firmwareUpgradeCmd.Flags().StringVar(&firmwareUpgradeShareName, "share-name", "", "")
	firmwareUpgradeCmd.Flags().StringVar(&firmwareUpgradeShareType, "share-type", "HTTPS", "Valid options: HTTPS, HTTP, NFS, or CIFS")

	firmwareCmd.AddCommand(firmwareUpgradeCmd)
}
