package nodeset

import (
	"bytes"
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

type NodeSet struct {
	pat map[string]*RangeSetND
}

func NewNodeSet(nodestr string) (*NodeSet, error) {
	nodestr = strings.TrimSpace(nodestr)
	if nodestr == "" {
		return nil, fmt.Errorf("emtpy nodeset - %w", ErrParseNodeSet)
	}

	ranges := rangeSetRegexp.FindAllStringSubmatch(nodestr, -1)
	patterns := rangeSetRegexp.ReplaceAllString(nodestr, "%s")

	if strings.Index(patterns, "[") != -1 {
		return nil, fmt.Errorf("unbalanced '[' found while parsing %s - %w", nodestr, ErrParseNodeSet)
	}
	if strings.Index(patterns, "]") != -1 {
		return nil, fmt.Errorf("unbalanced ']' found while parsing %s - %w", nodestr, ErrParseNodeSet)
	}

	ns := &NodeSet{pat: make(map[string]*RangeSetND, 0)}

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
			ns.pat[pattern] = rsnd
			ridx += rangeSetCount
		} else {
			ns.pat[pattern] = nil
		}
	}

	return ns, nil
}

func (ns *NodeSet) Len() int {
	size := 0
	for _, rs := range ns.pat {
		if rs == nil {
			size++
			continue
		}

		size += rs.Len()
	}

	return size
}

func (ns *NodeSet) String() string {
	var buffer bytes.Buffer

	i := 0
	for pat, rs := range ns.pat {
		if rs == nil {
			buffer.WriteString(pat)
		} else {
			ranges := rs.Ranges()
			params := make([]interface{}, 0, len(ranges))
			for _, r := range ranges {
				params = append(params, fmt.Sprintf("[%s]", r.String()))
			}
			buffer.WriteString(fmt.Sprintf(pat, params...))
		}

		if i != len(ns.pat)-1 {
			buffer.WriteString(",")
		}

		i++
	}

	return buffer.String()
}

func (ns *NodeSet) Iterator() *NodeSetIterator {
	nodes := make([]string, 0, ns.Len())

	for pat, rsnd := range ns.pat {
		if rsnd == nil || rsnd.Len() == 0 {
			nodes = append(nodes, pat)
			continue
		}

		ranges := rsnd.Ranges()
		slices := make([][]interface{}, rsnd.Dim())
		for i := 0; i < rsnd.Dim(); i++ {
			strings := ranges[i].Strings()
			slices[i] = toIface(strings)
		}

		c := cartesian.Iter(slices...)
		for rec := range c {
			nodes = append(nodes, fmt.Sprintf(pat, rec...))
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
