// Copyright 2019 Grendel Authors. All rights reserved.
//
// This file is part of Grendel.
//
// Grendel is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Grendel is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Grendel. If not, see <https://www.gnu.org/licenses/>.

package nodeset

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/schwarmco/go-cartesian-product"
)

var (
	ErrInvalidNodeSet = errors.New("invalid nodeset")
	ErrParseNodeSet   = errors.New("nodeset parse error")
	rangeSetRegexp    = regexp.MustCompile(`(\[[^\[\]]+\])`)
)

type Pattern struct {
	format string
	set    *RangeSetND
}

type NodeSet struct {
	pat []*Pattern
}

func NewNodeSet(nodestr string) (*NodeSet, error) {
	nodestr = strings.TrimSpace(nodestr)
	if nodestr == "" {
		return nil, fmt.Errorf("empty nodeset - %w", ErrParseNodeSet)
	}

	ranges := rangeSetRegexp.FindAllStringSubmatch(nodestr, -1)
	patterns := rangeSetRegexp.ReplaceAllString(nodestr, "%s")

	if strings.Index(patterns, "[") != -1 {
		return nil, fmt.Errorf("unbalanced '[' found while parsing %s - %w", nodestr, ErrParseNodeSet)
	}
	if strings.Index(patterns, "]") != -1 {
		return nil, fmt.Errorf("unbalanced ']' found while parsing %s - %w", nodestr, ErrParseNodeSet)
	}

	ns := &NodeSet{pat: make([]*Pattern, 0)}

	ridx := 0
	for _, pattern := range strings.Split(patterns, ",") {
		rangeSetCount := strings.Count(pattern, "%s")
		if rangeSetCount > 0 {
			rangeSets := make([]string, 0)
			for i := ridx; i < ridx+rangeSetCount; i++ {
				rangeSets = append(rangeSets, strings.Trim(ranges[i][1], "[]"))
			}

			rsnd, err := NewRangeSetND(rangeSets)
			if err != nil {
				return nil, fmt.Errorf("%w", err)
			}
			ns.pat = append(ns.pat, &Pattern{format: pattern, set: rsnd})
			ridx += rangeSetCount
		} else {
			ns.pat = append(ns.pat, &Pattern{format: pattern, set: nil})
		}
	}

	return ns, nil
}

func (ns *NodeSet) Len() int {
	size := 0
	for _, p := range ns.pat {
		if p.set == nil {
			size++
			continue
		}

		size += p.set.Len()
	}

	return size
}

func (ns *NodeSet) String() string {
	list := ns.toStringList()
	return strings.Join(list, ",")
}

func (ns *NodeSet) toStringList() []string {
	list := make([]string, 0)

	for _, p := range ns.pat {
		if p.set == nil {
			list = append(list, p.format)
		} else {
			ranges := p.set.Ranges()
			params := make([]interface{}, 0, len(ranges))
			for _, r := range ranges {
				params = append(params, fmt.Sprintf("[%s]", r.String()))
			}
			list = append(list, fmt.Sprintf(p.format, params...))
		}
	}

	return list
}

func (ns *NodeSet) MarshalJSON() ([]byte, error) {
	list := ns.toStringList()
	return json.Marshal(&list)
}

func (ns *NodeSet) UnmarshalJSON(data []byte) error {
	var list []string
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	n, err := NewNodeSet(strings.Join(list, ","))
	if err != nil {
		return err
	}

	*ns = *n

	return nil
}

func (ns *NodeSet) Iterator() *NodeSetIterator {
	nodes := make([]string, 0, ns.Len())

	for _, p := range ns.pat {
		if p.set == nil || p.set.Len() == 0 {
			nodes = append(nodes, p.format)
			continue
		}

		ranges := p.set.Ranges()
		slices := make([][]interface{}, p.set.Dim())
		for i := 0; i < p.set.Dim(); i++ {
			strings := ranges[i].Strings()
			slices[i] = toIface(strings)
		}

		c := cartesian.Iter(slices...)
		for rec := range c {
			nodes = append(nodes, fmt.Sprintf(p.format, rec...))
		}
	}

	sort.Strings(nodes)

	return &NodeSetIterator{nodes: nodes, current: -1}
}

func toIface(list []string) []interface{} {
	vals := make([]interface{}, len(list))
	for i, v := range list {
		vals[i] = v
	}
	return vals
}
