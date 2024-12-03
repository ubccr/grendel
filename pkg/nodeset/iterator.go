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
	"fmt"
	"sort"

	"github.com/segmentio/fasthash/fnv1a"
)

type NodeSetIterator struct {
	nodes   []string
	current int
}

type RangeSetItem struct {
	value   int
	padding int
}

type RangeSetNDIterator struct {
	vects   [][]*RangeSetItem
	seen    map[uint64]struct{}
	current int
}

func (i *NodeSetIterator) Next() bool {
	i.current++

	if i.current < len(i.nodes) {
		return true
	}

	return false
}

func (i *NodeSetIterator) Len() int {
	return len(i.nodes)
}

func (i *NodeSetIterator) Value() string {
	return i.nodes[i.current]
}

func NewRangeSetNDIterator() *RangeSetNDIterator {
	return &RangeSetNDIterator{
		vects:   make([][]*RangeSetItem, 0),
		seen:    make(map[uint64]struct{}),
		current: -1,
	}
}

func (i *RangeSetNDIterator) Next() bool {
	i.current++

	if i.current < len(i.vects) {
		return true
	}

	return false
}

func (i *RangeSetNDIterator) Len() int {
	return len(i.vects)
}

func (i *RangeSetNDIterator) IntValue() []int {
	vals := make([]int, 0, len(i.vects[i.current]))
	for _, v := range i.vects[i.current] {
		vals = append(vals, v.value)
	}
	return vals
}

func (i *RangeSetNDIterator) FormatList() []interface{} {
	vals := make([]interface{}, 0, len(i.vects[i.current]))
	for _, v := range i.vects[i.current] {
		vals = append(vals, fmt.Sprintf("%0*d", v.padding, v.value))
	}
	return vals
}

func (it *RangeSetNDIterator) Sort() {
	sort.SliceStable(it.vects, func(i, j int) bool {
		for x := range it.vects[i] {
			if it.vects[i][x].value != it.vects[j][x].value {
				return it.vects[i][x].value < it.vects[j][x].value
			}
		}
		return false
	})
}

func (it *RangeSetNDIterator) product(result []*RangeSetItem, params ...[]*RangeSetItem) {
	if len(params) == 0 {
		hash := fnv1a.Init64
		for _, i := range result {
			hash = fnv1a.AddUint64(hash, uint64(i.value))
		}

		if _, ok := it.seen[hash]; !ok {
			it.seen[hash] = struct{}{}
			it.vects = append(it.vects, result)
		}

		return
	}

	p, params := params[0], params[1:]
	for i := 0; i < len(p); i++ {
		resultCopy := append([]*RangeSetItem{}, result...)
		it.product(append(resultCopy, p[i]), params...)
	}
}
