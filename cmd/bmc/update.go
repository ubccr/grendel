// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	firmwareCmd = &cobra.Command{
		Use:   "firmware",
		Short: "BMC Firmware commands",
		Long:  `BMC Firmware commands`,
	}

	firmwareCheckCmd = &cobra.Command{
		Use:   "check <nodeset>",
		Short: "Check for updates on Dell servers",
		Long: `Check for updates on Dell servers
Must run bmc upgrade <nodeset> to populate firmware data`,
		Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			nodeset := args[0]

			if nodeset == "all" {
				nodeset = ""
			}

			params := client.GETV1BmcUpgradeDellRepoParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := gc.GETV1BmcUpgradeDellRepo(context.Background(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Host", "Component", "Current Version", "Latest Version", "Reboot Required"})
			t.SetColumnConfigs([]table.ColumnConfig{
				{
					Name:      "Host",
					AutoMerge: true,
				},
			})

			for _, host := range res {
				if host.Status.Value != "success" {
					fmt.Printf("%s\t%s\n", host.Name.Value, host.Message.Value)
				}
				for _, fw := range host.UpdateList {
					t.AppendRow(table.Row{
						host.Name.Value,
						fw.DisplayName.Value,
						fw.InstalledVersion.Value,
						colorVersion(fw.InstalledVersion.Value, fw.PackageVersion.Value),
						fw.RebootType.Value,
					}, table.RowConfig{AutoMerge: true})
				}
				t.AppendSeparator()
			}

			t.AppendSeparator()

			t.SetStyle(table.StyleLight)
			t.Render()

			return nil
		},
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
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}
			nodeset := args[0]

			req := client.BmcDellInstallFromRepoRequest{
				ApplyUpdate:       client.NewOptBool(firmwareUpgradeApplyUpdate),
				CatalogFile:       client.NewOptString(firmwareUpgradeCatalogFile),
				IPAddress:         client.NewOptString(firmwareUpgradeIpAddress),
				IgnoreCertWarning: client.NewOptBool(firmwareUpgradeIgnoreCertWarning),
				RebootNeeded:      client.NewOptBool(firmwareUpgradeRebootNeeded),
				ShareName:         client.NewOptString(firmwareUpgradeShareName),
				ShareType:         client.NewOptString(firmwareUpgradeShareType),
			}
			params := client.POSTV1BmcUpgradeDellInstallfromrepoParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := gc.POSTV1BmcUpgradeDellInstallfromrepo(context.Background(), &req, params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			for _, jobMessage := range res {
				fmt.Printf("%s\t %s\t %s\n", jobMessage.Host.Value, jobMessage.Status.Value, jobMessage.Msg.Value)
			}

			return nil
		},
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
	firmwareCmd.AddCommand(firmwareCheckCmd)
	firmwareCmd.AddCommand(firmwareUpgradeCmd)

	firmwareUpgradeCmd.Flags().BoolVarP(&firmwareUpgradeApplyUpdate, "apply-update", "a", false, "By default only check for updates, do not queue them. Pass this flag to apply available updates.")
	firmwareUpgradeCmd.Flags().StringVar(&firmwareUpgradeCatalogFile, "catalog-file", "", "Update catalog name. Defaults to Catalog.xml")
	firmwareUpgradeCmd.Flags().StringVarP(&firmwareUpgradeIpAddress, "ip-address", "i", "downloads.dell.com", "IP or Domain name of share")
	firmwareUpgradeCmd.Flags().BoolVar(&firmwareUpgradeIgnoreCertWarning, "ignore-cert-warning", true, "Pass this flag to ignore invalid certs")
	firmwareUpgradeCmd.Flags().BoolVarP(&firmwareUpgradeRebootNeeded, "reboot", "r", false, "Reboot arg will immediately reboot the node when needed")
	firmwareUpgradeCmd.Flags().BoolVar(&firmwareUpgradeClearJobs, "clear-jobs", false, "Clear all jobs in the job queue before upgrading. apply-update must be true")
	firmwareUpgradeCmd.Flags().StringVar(&firmwareUpgradeShareName, "share-name", "", "")
	firmwareUpgradeCmd.Flags().StringVar(&firmwareUpgradeShareType, "share-type", "HTTPS", "Valid options: HTTPS, HTTP, NFS, or CIFS")
}
