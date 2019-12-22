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

type slice struct {
	start uint
	stop  uint
}

func NewRangeSet(pattern string) (rs *RangeSet, err error) {
	rs = &RangeSet{}
	for _, subrange := range strings.Split(pattern, ",") {
		if subrange == "" {
			return nil, fmt.Errorf("emtpy range - %w", ErrParseRangeSet)
		}

		baserange := subrange
		step := 1
		if strings.Index(subrange, "/") >= 0 {
			parts := strings.SplitN(subrange, "/", 2)
			baserange = parts[0]
			if len(parts) != 2 || parts[1] == "" {
				return nil, fmt.Errorf("cannont parse step %s - %w", subrange, ErrParseRangeSet)
			}

			var err error
			step, err = strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("cannont convert step to integer %s - %w", subrange, ErrParseRangeSet)
			}
		}

		var start, stop, pad int
		parts := []string{baserange}

		if strings.Index(baserange, "-") < 0 {
			if step != 1 {
				return nil, fmt.Errorf("invalid step usage %s - %w", subrange, ErrParseRangeSet)
			}
		} else {
			parts = strings.SplitN(baserange, "-", 2)
			if len(parts) != 2 || parts[1] == "" {
				return nil, fmt.Errorf("cannpt parse end value %s - %w", subrange, ErrParseRangeSet)
			}
		}

		start, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("cannont convert starting range to integer %s - %w", parts[0], ErrParseRangeSet)
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
				return nil, fmt.Errorf("cannont convert ending range to integer %s - %w", parts[1], ErrParseRangeSet)
			}
		} else {
			stop = start
		}

		if stop > math.MaxInt64 || start > stop || step < 1 {
			return nil, fmt.Errorf("invalid value in range %s - %w", subrange, ErrParseRangeSet)
		}

		rs.AddRange(start, stop+1, step, pad)
	}

	return rs, nil
}

func (rs *RangeSet) AddRange(start, stop, step, pad int) error {
	if start > stop {
		return fmt.Errorf("invalid range start > stop - %w", ErrInvalidRangeSet)
	}
	if step <= 0 {
		return fmt.Errorf("invalid range step <= 0 - %w", ErrInvalidRangeSet)
	}
	if pad < 0 {
		return fmt.Errorf("invalid range padding < 0 - %w", ErrInvalidRangeSet)
	}
	if stop-start > math.MaxInt64 {
		return fmt.Errorf("range too large - %w", ErrInvalidRangeSet)
	}

	if pad > 0 && rs.padding == 0 {
		rs.padding = pad
	}

	rs.update(start, stop, step)

	return nil
}

func (rs *RangeSet) Len() int {
	return int(rs.bits.Count())
}

func (rs *RangeSet) String() string {
	var buffer bytes.Buffer
	slices := rs.slices()
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

func (rs *RangeSet) update(start, stop, step int) {
	for i := start; i < stop; i += step {
		rs.bits.Set(uint(i))
	}
}

func (s *slice) String() string {
	return fmt.Sprintf("%d-%d", s.start, s.stop)
}

func (rs *RangeSet) slices() []*slice {
	result := make([]*slice, 0)
	i, e := rs.bits.NextSet(0)
	k := i
	j := i
	for e {
		if i-j > 1 {
			result = append(result, &slice{k, j + 1})
			k = i
		}
		j = i
		i, e = rs.bits.NextSet(i + 1)
	}
	result = append(result, &slice{k, j + 1})

	return result
}
