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
)

var (
	ErrInvalidNodeSet = errors.New("invalid nodeset")
	ErrParseNodeSet   = errors.New("nodeset parse error")
	rangeSetRegexp    = regexp.MustCompile(`(\[[^\[\]]+\]|[0-9]+)`)
)

type Pattern struct {
	format   string
	rangeSet *RangeSetND
}

type NodeSet struct {
	patterns map[string]*RangeSetND
}

func NewNodeSet(nodestr string) (*NodeSet, error) {
	ns := &NodeSet{patterns: make(map[string]*RangeSetND, 0)}
	err := ns.Add(nodestr)
	if err != nil {
		return nil, err
	}

	return ns, nil
}

func (ns *NodeSet) Add(nodestr string) error {
	nodestr = strings.ReplaceAll(nodestr, " ", "")
	if nodestr == "" {
		return fmt.Errorf("empty nodeset - %w", ErrParseNodeSet)
	}

	ranges := rangeSetRegexp.FindAllStringSubmatch(nodestr, -1)
	patterns := rangeSetRegexp.ReplaceAllString(nodestr, "%s")

	if strings.Index(patterns, "[") != -1 {
		return fmt.Errorf("unbalanced '[' found while parsing %s - %w", nodestr, ErrParseNodeSet)
	}
	if strings.Index(patterns, "]") != -1 {
		return fmt.Errorf("unbalanced ']' found while parsing %s - %w", nodestr, ErrParseNodeSet)
	}

	ridx := 0
	for _, pattern := range strings.Split(patterns, ",") {
		rangeSetCount := strings.Count(pattern, "%s")
		if rangeSetCount == 0 {
			ns.patterns[pattern] = nil
			continue
		}

		rangeSets := make([]string, 0)
		for i := ridx; i < ridx+rangeSetCount; i++ {
			rangeSets = append(rangeSets, strings.Trim(ranges[i][1], "[]"))
		}

		rs, err := NewRangeSetND([][]string{rangeSets})
		if err != nil {
			return err
		}

		if _, ok := ns.patterns[pattern]; !ok {
			ns.patterns[pattern] = rs
		} else {
			err = ns.patterns[pattern].Update(rs)
			if err != nil {
				return err
			}
		}

		ridx += rangeSetCount
	}

	return nil
}

func (ns *NodeSet) Len() int {
	size := 0

	for _, rs := range ns.patterns {
		if rs == nil {
			size++
		} else {
			size += rs.Len()
		}
	}

	return size
}

func (ns *NodeSet) String() string {
	list := ns.toStringList()
	return strings.Join(list, ",")
}

func (ns *NodeSet) toStringList() []string {
	list := make([]string, 0)

	items := make([]*Pattern, 0, len(ns.patterns))
	for pattern, rs := range ns.patterns {
		items = append(items, &Pattern{format: pattern, rangeSet: rs})
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].rangeSet.Len() > items[j].rangeSet.Len()
	})

	for _, pattern := range items {
		if pattern.rangeSet == nil {
			list = append(list, pattern.format)
			continue
		}

		for _, params := range pattern.rangeSet.FormatList() {
			list = append(list, fmt.Sprintf(pattern.format, params...))
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

	items := make([]*Pattern, 0, len(ns.patterns))
	for pattern, rs := range ns.patterns {
		items = append(items, &Pattern{format: pattern, rangeSet: rs})
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].rangeSet.Len() > items[j].rangeSet.Len()
	})

	for _, pattern := range items {
		if pattern.rangeSet == nil {
			nodes = append(nodes, pattern.format)
			continue
		}

		it := pattern.rangeSet.Iterator()
		for it.Next() {
			params := it.FormatList()
			nodes = append(nodes, fmt.Sprintf(pattern.format, params...))
		}

	}

	return &NodeSetIterator{nodes: nodes, current: -1}
}
