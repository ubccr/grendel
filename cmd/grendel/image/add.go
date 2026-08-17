// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package image

import (
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	cmdline            string
	initrd             []string
	kernel             string
	provisionTemplates map[string]string
	verify             bool
	newCmd             = &cobra.Command{
		Use:   "add <name>",
		Short: "add image",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			provisionTemplatesMap := make(client.BootImageAddRequestBootImagesItemProvisionTemplates, 0)
			for k, v := range provisionTemplates {
				provisionTemplatesMap[k] = client.NewNilString(v)
			}

			newImage := []client.NilBootImageAddRequestBootImagesItem{
				client.NewNilBootImageAddRequestBootImagesItem(client.BootImageAddRequestBootImagesItem{
					Name:               client.NewOptString(args[0]),
					Cmdline:            client.NewOptString(cmdline),
					Initrd:             initrd,
					Kernel:             client.NewOptString(kernel),
					ProvisionTemplates: client.NewOptNilBootImageAddRequestBootImagesItemProvisionTemplates(provisionTemplatesMap),
					Verify:             client.NewOptBool(verify),
				}),
			}

			storeReq := &client.BootImageAddRequest{
				BootImages: newImage,
			}
			storeParams := client.POSTV1ImagesParams{}

			storeRes, err := apiclient.API.POSTV1Images(command.Context(), storeReq, storeParams)
			if err != nil {
				return apiclient.NewApiError(err)
			}

			return apiclient.NewApiResponse(storeRes)
		},
	}
)

func init() {
	newCmd.Flags().StringVar(&cmdline, "cmdline", "", "Kernel Command Line")
	newCmd.Flags().StringSliceVar(&initrd, "initrd", nil, "Path to Initrd image")
	newCmd.Flags().StringVar(&kernel, "kernel", "", "Path to Kernel")
	newCmd.Flags().StringToStringVar(&provisionTemplates, "provision-template", nil, "Provision template map. Example: `kickstart=/path.tmpl`")
	newCmd.Flags().BoolVar(&verify, "verify", false, "Verify the image through iPXE on boot. Requires a .sig file for the kernel & initrd in the same directory")

	imageCmd.AddCommand(newCmd)
}
