package nodeset

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/willf/bitset"
)

type NodeSet struct {
	bits    bitset.BitSet
	size    int
	current int
	prefix  string
}

func NewNodeSet(prefix, pattern string) (*NodeSet, error) {
	n := &NodeSet{prefix: prefix, current: -1}

	pattern = strings.ReplaceAll(pattern, " ", "")

	intervals := strings.Split(pattern, ",")

	for _, i := range intervals {
		numbers := strings.Split(i, "-")

		if len(numbers) == 1 {
			bit, err := strconv.Atoi(numbers[0])
			if err != nil {
				return nil, err
			}
			n.bits.Set(uint(bit))
			n.size++
		} else if len(numbers) == 2 {
			from, err := strconv.Atoi(numbers[0])
			if err != nil {
				return nil, err
			}
			to, err := strconv.Atoi(numbers[1])
			if err != nil {
				return nil, err
			}

			for j := from; j <= to; j++ {
				n.bits.Set(uint(j))
				n.size++
			}
		}
	}

	return n, nil
}

func (n *NodeSet) Len() int {
	return n.size
}

func (n *NodeSet) Next() bool {
	if !n.bits.Any() {
		return false
	}

	bit, ok := n.bits.NextSet(uint(n.current) + 1)
	if !ok {
		return false
	}

	n.current = int(bit)

	return true
}

func (n *NodeSet) Value() string {
	return fmt.Sprintf("%s%02d", n.prefix, n.current)
}

func (n *NodeSet) IntValue() int {
	return n.current
}
