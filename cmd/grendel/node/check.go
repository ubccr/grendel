// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	checkFailed  bool
	checkTable   *shared.TableFlags
	checkColumns = []shared.Column{
		{Name: "Node"},
		{Name: "Arch"},
		{Name: "Image"},
		{Name: "Result"},
		{Name: "Files"},
		{Name: "Templates"},
		{Name: "Problems"},
	}
	checkCmd = &cobra.Command{
		Use:   "check {nodeset | all}",
		Short: "Check that nodes can be served their boot files and templates",
		Long: `Check, for each node, that the kernel and initrds of its boot image are readable on the server at the paths rendered for the node, and that the iPXE script and every template in the image's provision_templates render.

The Files and Templates columns count how many were readable and rendered out of those the image has. Files are read on the machine the API runs on. Exits non-zero when any node fails.`,
		Example: `  grendel node check cpn-[01-40]
  grendel node check all --tags compute --failed
  grendel node check cpn-[01-40] --arch all`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: nodesetCompletion,
		RunE: func(command *cobra.Command, args []string) error {
			nodeset := args[0]
			if nodeset == "all" {
				nodeset = ""
			}
			res, err := apiclient.API.GETV1NodesCheck(command.Context(), client.GETV1NodesCheckParams{
				Nodeset: client.NewOptString(nodeset),
				Tags:    client.NewOptString(strings.Join(tags, ",")),
				Arch:    client.NewOptString(renderArch),
				Server:  client.NewOptString(renderServer),
			})
			if err != nil {
				return apiclient.NewApiError(err)
			}

			t := checkTable.Table()
			failed := 0
			for _, node := range res.Nodes {
				if node.Ok.Value && checkFailed {
					continue
				}

				result := "ok"
				problems := []string{}
				if !node.Ok.Value {
					failed++
					result = "failed"
					if node.Error.Value != "" {
						problems = append(problems, node.Error.Value)
					}
				}

				readable := 0
				for _, f := range node.Files {
					if f.Error.Value == "" {
						readable++
						continue
					}
					problems = append(problems, fmt.Sprintf("%s: %s", f.Name.Value, f.Error.Value))
				}

				rendered := 0
				for _, tmpl := range node.Templates {
					if tmpl.Error.Value == "" {
						rendered++
						continue
					}
					problems = append(problems, fmt.Sprintf("%s (%s): %s", tmpl.Kind.Value, tmpl.Template.Value, tmpl.Error.Value))
				}

				t.AppendRow(
					node.Name.Value,
					node.Arch.Value,
					node.Image.Value,
					result,
					outOf(readable, len(node.Files)),
					outOf(rendered, len(node.Templates)),
					strings.Join(problems, "\n"),
				)
			}

			if failed == 0 && checkFailed {
				fmt.Printf("all %d nodes ok\n", len(res.Nodes))
				return nil
			}

			t.Print()
			if failed > 0 {
				return fmt.Errorf("%d of %d nodes failed", failed, len(res.Nodes))
			}
			return nil
		},
	}
)

// outOf is how many of total passed, or nothing when there were none to try
func outOf(passed, total int) string {
	if total == 0 {
		return ""
	}
	return fmt.Sprintf("%d/%d", passed, total)
}

func init() {
	archFlags(checkCmd)
	checkCmd.Flags().BoolVar(&checkFailed, "failed", false, "Only list the nodes that failed")
	checkTable = shared.RegisterTableFlags(checkCmd, checkColumns)
	nodeCmd.AddCommand(checkCmd)
}
