// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/pkg/client"
	"github.com/ubccr/grendel/pkg/model"
)

var (
	renderBMC    bool
	renderArch   string
	renderServer string
	renderCmd    = &cobra.Command{
		Use:   "render <nodeset> <template>",
		Short: "Render a provision template for nodes",
		Long: `Print a template as the provision server would send it to each node, without booting the node or recording a provision stage.

The template is ipxe, kickstart, user_data, meta_data, butane, ignition (the butane template translated), proxmox, a key of the image's provision_templates, or a template file name.

Templates are read from disk each time, so this shows an edit at once. The output holds a working boot token and whatever secrets the template renders.`,
		Example: `  grendel node render cpn-01 ipxe
  grendel node render cpn-[01-04] kickstart --arch aarch64
  grendel node render cpn-01 idrac-config.json.tmpl --bmc
  grendel node render cpn-01 ipxe --arch all`,
		Args: cobra.ExactArgs(2),
		ValidArgsFunction: func(command *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 0:
				return nodesetCompletion(command, args, toComplete)
			case 1:
				return templateCompletion(command, args[0]), cobra.ShellCompDirectiveNoFileComp
			default:
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
		},
		RunE: func(command *cobra.Command, args []string) error {
			iface := "boot"
			if renderBMC {
				iface = "bmc"
			}
			res, err := apiclient.API.GETV1NodesRenderTemplate(command.Context(), client.GETV1NodesRenderTemplateParams{
				Template:  args[1],
				Nodeset:   client.NewOptString(args[0]),
				Tags:      client.NewOptString(strings.Join(tags, ",")),
				Interface: client.NewOptString(iface),
				Arch:      client.NewOptString(renderArch),
				Server:    client.NewOptString(renderServer),
			})
			if err != nil {
				return apiclient.NewApiError(err)
			}

			failed := 0
			for _, node := range res.Nodes {
				if node.Error.Value != "" {
					failed++
					fmt.Fprintf(os.Stderr, "%s: %s\n", node.Name.Value, node.Error.Value)
					continue
				}

				// one node prints bare, so the output can be piped to a file or a validator
				if len(res.Nodes) > 1 {
					header := node.Template.Value
					if node.Arch.Value != "" {
						header += ", " + node.Arch.Value
					}
					fmt.Printf("==> %s (%s) <==\n", node.Name.Value, header)
				}
				fmt.Print(node.Content.Value)
				if !strings.HasSuffix(node.Content.Value, "\n") {
					fmt.Println()
				}
			}

			if failed > 0 {
				return fmt.Errorf("%d of %d nodes did not render", failed, len(res.Nodes))
			}
			return nil
		},
	}
)

// templateCompletion offers the kinds every image has and the provision_templates keys of the images the nodes boot. A lookup that fails only costs the keys, since the kinds still complete
func templateCompletion(command *cobra.Command, nodeset string) []string {
	kinds := []string{"ipxe", "kickstart", "user_data", "meta_data", "butane", "ignition", "proxmox"}

	tags, _ := command.Flags().GetStringSlice("tags")
	nodes, err := apiclient.API.GETV1NodesFind(command.Context(), client.GETV1NodesFindParams{
		Nodeset: client.NewOptString(nodeset),
		Tags:    client.NewOptString(strings.Join(tags, ",")),
	})
	if err == nil {
		var names []string
		for _, node := range nodes {
			if name := node.BootImage.Value; name != "" && !slices.Contains(names, name) {
				names = append(names, name)
			}
		}

		// with no names the route would return every image, none of which these nodes boot
		if len(names) > 0 {
			images, err := apiclient.API.GETV1ImagesFind(command.Context(), client.GETV1ImagesFindParams{Names: client.NewOptString(strings.Join(names, ","))})
			if err == nil {
				for _, image := range images {
					for key := range image.ProvisionTemplates.Value {
						if !slices.Contains(kinds, key) {
							kinds = append(kinds, key)
						}
					}
				}
			}
		}
	}

	return cobra.AppendActiveHelp(kinds, "A template file name also works, such as idrac-config.json.tmpl")
}

// archFlags adds the flags render and check share, which stand in for what only a real boot knows
func archFlags(command *cobra.Command) {
	command.Flags().StringVar(&renderArch, "arch", "", "Arch for nodes with no arch set, standing in for the one detected during a PXE boot: x86_64, aarch64, or all for each in turn")
	command.RegisterFlagCompletionFunc("arch", cobra.FixedCompletions(append(slices.Clone(model.Arches), "all"), cobra.ShellCompDirectiveNoFileComp))
	command.Flags().StringVar(&renderServer, "server", "", "Address of the provision server to put in URLs, when the server's own choice is not the one the nodes use")
}

func init() {
	renderCmd.Flags().BoolVar(&renderBMC, "bmc", false, "Render for the node's BMC interface instead of its boot interface")
	archFlags(renderCmd)
	nodeCmd.AddCommand(renderCmd)
}
