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
