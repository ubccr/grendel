package nodeset

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/willf/bitset"
)

var (
	ErrInvalidRangeSet = errors.New("invalid range set")
	ErrParseRangeSet   = errors.New("rangeset parse error")
	ErrInvalidPadding  = errors.New("invalid padding")
)

type RangeSet struct {
	bits    bitset.BitSet
	padding int
}

type RangeSetND struct {
	ranges []*RangeSet
}

type Slice struct {
	start int
	stop  int
	step  int
	pad   int
}

func NewRangeSet(pattern string) (rs *RangeSet, err error) {
	rs = &RangeSet{}
	for _, subrange := range strings.Split(pattern, ",") {
		err := rs.AddString(subrange)
		if err != nil {
			return nil, err
		}
	}

	return rs, nil
}

func (rs *RangeSet) AddString(subrange string) (err error) {
	if subrange == "" {
		return fmt.Errorf("empty range - %w", ErrParseRangeSet)
	}

	baserange := subrange
	step := 1
	if strings.Index(subrange, "/") >= 0 {
		parts := strings.SplitN(subrange, "/", 2)
		baserange = parts[0]
		if len(parts) != 2 || parts[1] == "" {
			return fmt.Errorf("cannont parse step %s - %w", subrange, ErrParseRangeSet)
		}

		step, err = strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("cannont convert step to integer %s - %w", subrange, ErrParseRangeSet)
		}
	}

	var start, stop, pad int
	parts := []string{baserange}

	if strings.Index(baserange, "-") < 0 {
		if step != 1 {
			return fmt.Errorf("invalid step usage %s - %w", subrange, ErrParseRangeSet)
		}
	} else {
		parts = strings.SplitN(baserange, "-", 2)
		if len(parts) != 2 || parts[1] == "" {
			return fmt.Errorf("cannpt parse end value %s - %w", subrange, ErrParseRangeSet)
		}
	}

	start, err = strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("cannont convert starting range to integer %s - %w", parts[0], ErrParseRangeSet)
	}

	if start != 0 {
		begins := strings.TrimLeft(parts[0], "0")
		if len(parts[0])-len(begins) > 0 {
			pad = len(parts[0])
		}
	} else {
		if len(parts[0]) > 1 {
			pad = len(parts[0])
		}
	}

	if len(parts) == 2 {
		stop, err = strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("cannont convert ending range to integer %s - %w", parts[1], ErrParseRangeSet)
		}
	} else {
		stop = start
	}

	if stop > math.MaxInt64 || start > stop || step < 1 {
		return fmt.Errorf("invalid value in range %s - %w", subrange, ErrParseRangeSet)
	}

	return rs.AddSlice(&Slice{start, stop + 1, step, pad})
}

func (rs *RangeSet) AddSlice(slice *Slice) error {
	if slice.start > slice.stop {
		return fmt.Errorf("invalid range start > stop - %w", ErrInvalidRangeSet)
	}
	if slice.step <= 0 {
		return fmt.Errorf("invalid range step <= 0 - %w", ErrInvalidRangeSet)
	}
	if slice.pad < 0 {
		return fmt.Errorf("invalid range padding < 0 - %w", ErrInvalidRangeSet)
	}
	if slice.stop-slice.start > math.MaxInt64 {
		return fmt.Errorf("range too large - %w", ErrInvalidRangeSet)
	}

	if slice.pad > 0 && rs.padding == 0 {
		rs.padding = slice.pad
	}

	rs.update(slice)

	return nil
}

func (rs *RangeSet) Superset(other *RangeSet) bool {
	return rs.bits.IsSuperSet(&other.bits)
}

func (rs *RangeSet) Subset(other *RangeSet) bool {
	return other.bits.IsSuperSet(&rs.bits)
}

func (rs *RangeSet) Len() int {
	return int(rs.bits.Count())
}

func (rs *RangeSet) String() string {
	var buffer bytes.Buffer
	slices := rs.Slices()
	for i, sli := range slices {
		if sli.start+1 == sli.stop {
			buffer.WriteString(fmt.Sprintf("%0*d", rs.padding, sli.start))
		} else {
			buffer.WriteString(fmt.Sprintf("%0*d-%0*d", rs.padding, sli.start, rs.padding, sli.stop-1))
		}
		if i != len(slices)-1 {
			buffer.WriteString(",")
		}
	}
	return buffer.String()
}

func (rs *RangeSet) Strings() []string {
	strings := make([]string, 0)
	for _, sli := range rs.Slices() {
		for i := sli.start; i < sli.stop; i += sli.step {
			strings = append(strings, fmt.Sprintf("%0*d", rs.padding, i))
		}
	}

	return strings
}

func (rs *RangeSet) update(slice *Slice) {
	for i := slice.start; i < slice.stop; i += slice.step {
		rs.bits.Set(uint(i))
	}
}

func (s *Slice) String() string {
	return fmt.Sprintf("%d-%d", s.start, s.stop)
}

func (rs *RangeSet) Slices() []*Slice {
	result := make([]*Slice, 0)
	i, e := rs.bits.NextSet(0)
	k := i
	j := i
	for e {
		if i-j > 1 {
			result = append(result, &Slice{int(k), int(j + 1), 1, rs.padding})
			k = i
		}
		j = i
		i, e = rs.bits.NextSet(i + 1)
	}
	result = append(result, &Slice{int(k), int(j) + 1, 1, rs.padding})

	return result
}

func NewRangeSetND(patterns []string) (nd *RangeSetND, err error) {
	nd = &RangeSetND{ranges: make([]*RangeSet, 0)}

	for _, pat := range patterns {
		rs, err := NewRangeSet(pat)
		if err != nil {
			return nil, err
		}

		nd.ranges = append(nd.ranges, rs)
	}

	return nd, nil
}

func (nd *RangeSetND) Update(patterns []string) error {
	if nd.Dim() != len(patterns) {
		return fmt.Errorf("mismatched dimensions - %w", ErrParseRangeSet)
	}

	for i := range nd.ranges {
		err := nd.ranges[i].AddString(patterns[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (nd *RangeSetND) Dim() int {
	return len(nd.ranges)
}

func (nd *RangeSetND) Len() int {
	if len(nd.ranges) == 0 {
		return 0
	}
	size := nd.ranges[0].Len()
	for _, rs := range nd.ranges[1:] {
		size *= rs.Len()
	}

	return size
}

func (nd *RangeSetND) String() string {
	var buffer bytes.Buffer
	for i, rs := range nd.ranges {
		buffer.WriteString(rs.String())
		if i != len(nd.ranges)-1 {
			buffer.WriteString("; ")
		}
	}

	return buffer.String()
}

func (nd *RangeSetND) Ranges() []*RangeSet {
	return nd.ranges
}

func (nd *RangeSetND) Superset(other *RangeSetND) bool {
	if nd.Dim() != other.Dim() {
		return false
	}

	count := 0
	for i, rs := range nd.ranges {
		if rs.Superset(other.ranges[i]) {
			count++
		}
	}

	return count == len(nd.ranges)
}

func (nd *RangeSetND) Subset(other *RangeSetND) bool {
	return other.Superset(nd)
}
