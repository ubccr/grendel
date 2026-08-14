// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	configureCmd = &cobra.Command{
		Use:   "configure",
		Short: "Configure BMC",
	}
	configureAutoCmd = &cobra.Command{
		Use:   "auto {nodeset | all}",
		Short: "Set iDRAC to Auto configure",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			params := client.POSTV1BmcConfigureAutoParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := cmd.API.POSTV1BmcConfigureAuto(command.Context(), params)
			if err != nil {
				return cmd.NewApiError(err)
			}

			for _, jobMessage := range res {
				fmt.Printf("%s\t %s\t %s\n", jobMessage.Host.Value, jobMessage.Status.Value, jobMessage.Msg.Value)
			}

			return nil
		},
	}
	configureImportCmd = &cobra.Command{
		Use:   "import {nodeset | all} {NoReboot | Graceful | Forced} <filename>",
		Short: "Import BMC configuration",
		Long: `Args:
	shutdownType:	action bmc should take, NoReboot will wait for the node to be rebooted manually before applying. Any other type WILL REBOOT THE NODE.
	filename:	idrac config file relative to grendel template folder.
		`,
		Args: cobra.ExactArgs(3),
		RunE: func(command *cobra.Command, args []string) error {
			nodeset := args[0]
			if args[0] == "all" {
				nodeset = ""
			}
			req := &client.BmcImportConfigurationRequest{
				ShutdownType: client.NewOptString(args[1]),
				File:         client.NewOptString(args[2]),
			}
			params := client.POSTV1BmcConfigureImportParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
			}
			res, err := cmd.API.POSTV1BmcConfigureImport(command.Context(), req, params)
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

func init() {
	bmcCmd.AddCommand(configureCmd)
	configureCmd.AddCommand(configureAutoCmd)
	configureCmd.AddCommand(configureImportCmd)
}
