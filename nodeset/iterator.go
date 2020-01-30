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
