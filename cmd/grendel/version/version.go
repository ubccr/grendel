// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	build "github.com/ubccr/grendel/internal/version"
	"github.com/ubccr/grendel/pkg/client"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the client and server versions",
	Long:  "Print the client and server versions. The server version is read over the api, so a zero exit also means the server is up and answering. Use --version instead to print the client version without contacting anything.",
	Args:  cobra.NoArgs,
	RunE: func(command *cobra.Command, args []string) error {
		fmt.Printf("Client: %s\n", build.Version)

		res, err := apiclient.API.GETV1GrendelVersion(command.Context(), client.GETV1GrendelVersionParams{})
		if err != nil {
			return apiclient.NewApiError(err)
		}

		fmt.Printf("Server: %s\n", res.GetVersion().Value)

		return nil
	},
}

func init() {
	shared.Root.AddCommand(versionCmd)
}
