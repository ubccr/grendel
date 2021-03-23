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
	"sort"

	"github.com/segmentio/fasthash/fnv1a"
)

type NodeSetIterator struct {
	nodes   []string
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

type RangeSetNDIterator struct {
	vects   [][]int
	seen    map[uint64]struct{}
	current int
}

func NewRangeSetNDIterator() *RangeSetNDIterator {
	return &RangeSetNDIterator{
		vects:   make([][]int, 0),
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

func (i *RangeSetNDIterator) Value() []int {
	return i.vects[i.current]
}

func (it *RangeSetNDIterator) Sort() {
	sort.SliceStable(it.vects, func(i, j int) bool {
		for x := range it.vects[i] {
			if it.vects[i][x] != it.vects[j][x] {
				return it.vects[i][x] < it.vects[j][x]
			}
		}
		return false
	})
}

func (it *RangeSetNDIterator) product(result []int, params ...[]int) {
	if len(params) == 0 {
		hash := fnv1a.Init64
		for _, i := range result {
			hash = fnv1a.AddUint64(hash, uint64(i))
		}

		if _, ok := it.seen[hash]; !ok {
			it.seen[hash] = struct{}{}
			it.vects = append(it.vects, result)
		}

		return
	}

	p, params := params[0], params[1:]
	for i := 0; i < len(p); i++ {
		resultCopy := append([]int{}, result...)
		it.product(append(resultCopy, p[i]), params...)
	}
}
