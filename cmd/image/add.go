// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package image

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	cmdline            string
	initrd             []string
	kernel             string
	liveimg            string
	provisionTemplates []string
	verify             bool
	newCmd             = &cobra.Command{
		Use:   "add <name>",
		Short: "add image",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewOgenClient()
			if err != nil {
				return err
			}

			provisionTemplatesMap := make(client.BootImageAddRequestBootImagesItemProvisionTemplates, 0)
			for _, str := range provisionTemplates {
				kv := strings.Split(str, "=")
				if len(kv) != 2 {
					return fmt.Errorf("failed to parse provision template key value pair %s", str)
				}

				provisionTemplatesMap[kv[0]] = client.NewNilString(kv[1])
			}

			newImage := []client.NilBootImageAddRequestBootImagesItem{
				client.NewNilBootImageAddRequestBootImagesItem(client.BootImageAddRequestBootImagesItem{
					Name:               client.NewOptString(args[0]),
					Cmdline:            client.NewOptString(cmdline),
					Initrd:             initrd,
					Kernel:             client.NewOptString(kernel),
					Liveimg:            client.NewOptString(liveimg),
					ProvisionTemplates: client.NewOptNilBootImageAddRequestBootImagesItemProvisionTemplates(provisionTemplatesMap),
					Verify:             client.NewOptBool(verify),
				}),
			}

			storeReq := &client.BootImageAddRequest{
				BootImages: newImage,
			}
			storeParams := client.POSTV1ImagesParams{}

			storeRes, err := gc.POSTV1Images(context.Background(), storeReq, storeParams)
			if err != nil {
				return cmd.NewApiError(err)
			}

			return cmd.NewApiResponse(storeRes)
		},
	}
)

func init() {
	newCmd.PersistentFlags().StringVar(&cmdline, "cmdline", "", "Kernel Command Line")
	newCmd.PersistentFlags().StringArrayVar(&initrd, "initrd", []string{}, "Path to Initrd images. Can be passed multiple times")
	newCmd.PersistentFlags().StringVar(&kernel, "kernel", "", "Path to Kernel")
	newCmd.PersistentFlags().StringVar(&liveimg, "liveimg", "", "Path to Live Image")
	newCmd.PersistentFlags().StringArrayVar(&provisionTemplates, "provision-template", []string{}, "Provision template map. Example: kickstart=/var/lib/grendel/templates/ubuntu-kickstart.tmpl  Can be passed multiple times")
	newCmd.PersistentFlags().BoolVar(&verify, "verify", false, "Verify the image through iPXE on boot. Requires a .sig file for the kernel & initrd in the same directory")

	imageCmd.AddCommand(newCmd)
}
