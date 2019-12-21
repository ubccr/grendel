package nodeset

import (
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

func NewRangeSet(pattern string) (rs *RangeSet, err error) {
	rs = &RangeSet{}
	for _, subrange := range strings.Split(pattern, ",") {
		if subrange == "" {
			return nil, fmt.Errorf("emtpy range - %w", ErrParseRangeSet)
		}

		step := 1
		parts := strings.SplitN(subrange, "/", 2)
		baserange := parts[0]
		if len(parts) > 1 && parts[1] != "" {
			step, err = strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("cannont convert step to integer %s - %w", subrange, ErrParseRangeSet)
			}
		}

		var start, stop, pad int

		parts = strings.SplitN(baserange, "-", 2)
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

		if len(parts) == 2 && parts[1] != "" {
			stop, err = strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("cannont convert ending range to integer %s - %w", parts[1], ErrParseRangeSet)
			}
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

	// inherit padding info only if currently not defined
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
	return rs.bits.String()
}

func (rs *RangeSet) update(start, stop, step int) {
	for i := start; i < stop; i += step {
		rs.bits.Set(uint(i))
	}
}
