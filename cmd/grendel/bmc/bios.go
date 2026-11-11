// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/go-faster/jx"
	"github.com/spf13/cobra"
	"github.com/ubccr/grendel/pkg/nodeset"
)

var (
	biosCmd = &cobra.Command{
		Use:   "bios",
		Short: "BMC BIOS settings commands",
	}
)

func init() {
	bmcCmd.AddCommand(biosCmd)
}

// rawValue shows an attribute value as it would be typed. A JSON string loses its quotes, anything else prints as JSON
func rawValue(raw jx.Raw) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

// biosValueGroup is the nodes that hold one value of an attribute
type biosValueGroup struct {
	Value string
	Nodes []string
}

// biosDiff finds the attributes that differ between nodes. Each node's value is its staged one when a change is pending, since that is what the node will run once it applies. A node missing the attribute counts as its own value, which is how a mix of models shows up
func biosDiff(nodes map[string]map[string]string, names []string) map[string][]biosValueGroup {
	all := map[string]bool{}
	for _, attrs := range nodes {
		for name := range attrs {
			all[name] = true
		}
	}
	if len(names) > 0 {
		all = map[string]bool{}
		for _, name := range names {
			all[name] = true
		}
	}

	out := map[string][]biosValueGroup{}
	for name := range all {
		byValue := map[string][]string{}
		for node, attrs := range nodes {
			value, ok := attrs[name]
			if !ok {
				value = "(absent)"
			}
			byValue[value] = append(byValue[value], node)
		}
		if len(byValue) < 2 {
			continue
		}

		groups := make([]biosValueGroup, 0, len(byValue))
		for value, holders := range byValue {
			sort.Strings(holders)
			groups = append(groups, biosValueGroup{Value: value, Nodes: holders})
		}
		// the most common value first, so the odd nodes out sit below it
		sort.Slice(groups, func(i, j int) bool {
			if len(groups[i].Nodes) != len(groups[j].Nodes) {
				return len(groups[i].Nodes) > len(groups[j].Nodes)
			}
			return groups[i].Value < groups[j].Value
		})
		out[name] = groups
	}
	return out
}

// fold joins node names into a nodeset string, or a plain list if they do not fold
func fold(names []string) string {
	joined := strings.Join(names, ",")
	ns, err := nodeset.NewNodeSet(joined)
	if err != nil {
		return joined
	}
	return ns.String()
}

// parseAssignments reads Name=Value arguments. Values go as strings, which the server converts to the type each attribute holds
func parseAssignments(args []string) (map[string]jx.Raw, error) {
	out := make(map[string]jx.Raw, len(args))
	for _, arg := range args {
		name, value, ok := strings.Cut(arg, "=")
		if !ok || name == "" {
			return nil, fmt.Errorf("%q is not Name=Value", arg)
		}
		if _, dup := out[name]; dup {
			return nil, fmt.Errorf("%s is set more than once", name)
		}
		raw, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		out[name] = raw
	}
	return out, nil
}
