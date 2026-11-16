// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package node

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd/grendel/apiclient"
	"github.com/ubccr/grendel/cmd/shared"
	"github.com/ubccr/grendel/pkg/client"
)

var (
	ipCount        int
	ipAfterHighest bool
	ipNoPing       bool
	ipExclude      []string
	ipBMC          bool

	ipRangesFilter  []string
	ipRangesOptions *shared.TableOptions
	ipRangesColumns = []shared.Column{
		{Name: "Range"},
		{Name: "Prefix"},
		{Name: "Nodes"},
	}

	ipCmd = &cobra.Command{
		Use:   "ip",
		Short: "Find addresses for nodes",
	}

	ipNextCmd = &cobra.Command{
		Use:   "next <range>",
		Short: "Suggest free addresses in a range",
		Long: `Suggest free addresses in a range: the lowest ones no node holds, or with
--after-highest the ones after the highest held. Each is pinged, and one that
answers is skipped. An address is printed with the prefix length the nodes
around it have, ready to give a port. It exits non-zero when nothing is free,
or when the addresses couldn't be pinged.

The range is a CIDR like 10.17.25.0/24, an address standing for its /24, or
the first three octets of one, like 10.17.25.`,
		Example: `  grendel node ip next 10.17.25
  grendel node ip next 10.17.25.0/24 -n 3 --after-highest
  grendel node ip next 10.17.0.0/18 -n 40 --no-ping`,
		Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			from := "lowest"
			if ipAfterHighest {
				from = "after_highest"
			}
			res, err := apiclient.API.GETV1NodesIpsNext(command.Context(), client.GETV1NodesIpsNextParams{
				Cidr:    args[0],
				From:    client.NewOptString(from),
				Count:   client.NewOptInt(ipCount),
				Ping:    client.NewOptBool(!ipNoPing),
				Exclude: client.NewOptString(strings.Join(ipExclude, ",")),
			})
			if err != nil {
				return apiclient.NewApiError(err)
			}

			// only the list asked for comes back, and only addresses nothing answered at are printed, so a script never takes one something is already on
			found := 0
			show := func(ip string, answered bool) {
				if !answered {
					fmt.Printf("%s/%d\n", ip, res.PrefixLen.Value)
					found++
				}
			}
			for _, s := range res.Lowest {
				show(s.IP.Value, s.Answered.Value)
			}
			for _, s := range res.AfterHighest {
				show(s.IP.Value, s.Answered.Value)
			}

			if found == 0 {
				return fmt.Errorf("no free address in %s", res.Range.Value)
			}
			// the addresses are still printed, but a script checking the exit code won't take them as checked
			if !ipNoPing && !res.Pinged.Value {
				return fmt.Errorf("the addresses weren't pinged: %s", res.PingError.Value)
			}
			return nil
		},
	}

	ipRangesCmd = &cobra.Command{
		Use:   "ranges [nodeset]",
		Short: "List the ranges nodes' addresses sit in",
		Long: `List the ranges the ports of some nodes sit in, each once however many ports
share it: the /24 of a larger subnet, or a /24 to /30 subnet whole. Nodes are
picked by nodeset, which may also be part of a name, by --tags, or both, and
every node is taken with neither. The prefix is the one most of a range's ports
carry.`,
		Example: `  grendel node ip ranges cpn-u25-[01-40]
  grendel node ip ranges -t rack=u25 --bmc`,
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: nodesetCompletion,
		RunE: func(command *cobra.Command, args []string) error {
			p := client.GETV1NodesIpsRangesParams{
				Tags: client.NewOptString(strings.Join(tags, ",")),
			}
			if len(args) == 1 {
				p.Nodeset = client.NewOptString(args[0])
			}
			if command.Flags().Changed("bmc") {
				p.Bmc = client.NewOptBool(ipBMC)
			}
			res, err := apiclient.API.GETV1NodesIpsRanges(command.Context(), p)
			if err != nil {
				return apiclient.NewApiError(err)
			}
			if len(res) == 0 {
				return errors.New("no addresses found on those nodes")
			}

			t := shared.NewTable(ipRangesColumns, ipRangesFilter).Options(*ipRangesOptions).Empty("-")
			for _, r := range res {
				t.AppendRow(r.Range.Value, fmt.Sprintf("/%d", r.PrefixLen.Value), strconv.Itoa(r.Nodes.Value))
			}
			t.Print()

			return nil
		},
	}
)

func init() {
	ipNextCmd.Flags().IntVarP(&ipCount, "count", "n", 1, "how many addresses to suggest, up to 8 when pinging or 64 with --no-ping")
	ipNextCmd.Flags().BoolVar(&ipAfterHighest, "after-highest", false, "suggest the addresses after the highest one in use rather than the lowest free")
	ipNextCmd.Flags().BoolVar(&ipNoPing, "no-ping", false, "don't ping the suggestions")
	ipNextCmd.Flags().StringSliceVarP(&ipExclude, "exclude", "x", nil, "addresses to treat as taken, like ones already handed out in a script")
	ipRangesCmd.Flags().BoolVar(&ipBMC, "bmc", false, "list the ranges with BMC ports first, or with --bmc=false the other ports first")
	ipRangesOptions = shared.RegisterTableFlags(ipRangesCmd)
	ipRangesCmd.Flags().StringSliceVar(&ipRangesFilter, "filter", nil, "filter table columns by name")
	ipRangesCmd.RegisterFlagCompletionFunc("filter", shared.ColumnCompletion(ipRangesColumns))

	ipCmd.AddCommand(ipNextCmd, ipRangesCmd)
	nodeCmd.AddCommand(ipCmd)
}
