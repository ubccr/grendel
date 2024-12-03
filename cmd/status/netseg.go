// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package status

import (
	"context"
	"fmt"
	"net/netip"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/cmd"
	"github.com/ubccr/grendel/pkg/model"
	"go4.org/netipx"
)

var (
	ipmap      map[netip.Addr]*model.Host
	prefixes   []netip.Prefix
	netsegTags []string
	netsegLong bool
	netsegNext bool
	netsegCmd  = &cobra.Command{
		Use:   "netseg",
		Short: "Show IP segmentation",
		Long:  `Show IP segmentation`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(command *cobra.Command, args []string) error {
			gc, err := cmd.NewClient()
			if err != nil {
				return err
			}

			inputTags := strings.Join(netsegTags, ",")
			var hostList model.HostList

			if inputTags == "" {
				hostList, _, err = gc.HostApi.HostList(context.Background())
			} else {
				hostList, _, err = gc.HostApi.HostTags(context.Background(), inputTags)
			}

			if err != nil {
				return cmd.NewApiError("Failed to list hosts", err)
			}

			var builder netipx.IPSetBuilder
			ipmap = make(map[netip.Addr]*model.Host)

			if len(args) > 0 {
				for _, p := range args {
					ipp, err := netip.ParsePrefix(p)
					if err != nil {
						return err
					}
					builder.AddPrefix(ipp)
				}
			} else {
				for _, host := range hostList {
					for _, i := range host.Interfaces {
						ipp, err := i.IP.Addr().Prefix(i.IP.Bits())
						if err != nil {
							return err
						}
						builder.AddPrefix(ipp)
					}
				}
			}

			iset, _ := builder.IPSet()
			prefixes = iset.Prefixes()

			for _, host := range hostList {
				for _, i := range host.Interfaces {
					if iset.Contains(i.IP.Addr()) {
						builder.Remove(i.IP.Addr())
						ipmap[i.IP.Addr()] = host
					}
				}
			}

			iset, _ = builder.IPSet()

			if netsegNext {
				for _, p := range iset.Prefixes() {
					i := p.Addr()
					last := netipx.PrefixLastIP(p)
					for ; i.Compare(last) <= 0; i = i.Next() {
						if netsegSkipIP(i) {
							continue
						}

						fmt.Printf("%s\n", i)
						return nil
					}
				}
				return nil
			} else if netsegLong {
				for _, p := range prefixes {
					i := p.Addr()
					last := netipx.PrefixLastIP(p)
					for ; i.Compare(last) <= 0; i = i.Next() {
						if host, ok := ipmap[i]; ok {
							fmt.Printf("%-20s%-20s\n", i, host.Name)
							continue
						}

						if netsegSkipIP(i) {
							continue
						}

						fmt.Printf("%-20s%-20s\n", i, "")
					}
				}
				return nil

			}

			keys := make([]netip.Addr, 0, len(ipmap))
			for k := range ipmap {
				keys = append(keys, k)
			}

			sort.Slice(keys, func(i, j int) bool {
				return keys[i].Less(keys[j])
			})

			for _, k := range keys {
				host := ipmap[k]
				name := ""
				for _, i := range host.Interfaces {
					if i.Addr() == k && i.FQDN != "" {
						name = i.HostName()
					}
				}

				fmt.Printf("%-20s%-20s%-40s%-45s\n", k, host.Name, name, strings.Join(host.Tags, ","))
			}

			return nil
		},
	}
)

func netsegSkipIP(ip netip.Addr) bool {
	for _, p := range prefixes {
		pfirst := p.Addr()
		plast := netipx.PrefixLastIP(p)
		if ip == pfirst || ip == plast {
			return true
		}
	}

	return false
}

func init() {
	netsegCmd.Flags().StringSliceVarP(&netsegTags, "tags", "t", []string{}, "filter by tags")
	netsegCmd.Flags().BoolVar(&netsegLong, "long", false, "Display long format")
	netsegCmd.Flags().BoolVar(&netsegNext, "next", false, "Display next available IP")
	statusCmd.AddCommand(netsegCmd)
}
