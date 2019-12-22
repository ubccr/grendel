package nodeset

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testRS(t *testing.T, test, res string, length int) {
	r1, err := NewRangeSet(test)
	assert.Nil(t, err)
	assert.Equal(t, r1.String(), res)
	assert.Equal(t, r1.Len(), length)
}
func TestRangeSetSimple(t *testing.T) {
	testRS(t, "0", "0", 1)
	testRS(t, "1", "1", 1)
	testRS(t, "0-2", "0-2", 3)
	testRS(t, "1-3", "1-3", 3)
	testRS(t, "1-3,4-6", "1-6", 6)
	testRS(t, "1-3,4-6,7-10", "1-10", 10)
	testRS(t, "0001-0010", "0001-0010", 10)
}

func TestRangeSetSimpleStep(t *testing.T) {
	testRS(t, "0-4/2", "0,2,4", 3)
	testRS(t, "1-4/2", "1,3", 2)
	testRS(t, "1-4/3", "1,4", 2)
	testRS(t, "1-4/4", "1", 1)
}

func TestRangeSetBadSyntax(t *testing.T) {
	badSyntax := []string{
		"",
		"-",
		"A",
		"2-5/a",
		"3/2",
		"3-/2",
		"-3/2",
		"-/2",
		"4-a/2",
		"4-3/2",
		"4-5/-2",
		"4-2/-2",
		"004-002",
		"3-59/2,102a",
	}

	for _, syn := range badSyntax {
		_, err := NewRangeSet(syn)
		if assert.Errorf(t, err, "Parsed %s ok", syn) {
			assert.Equal(t, errors.Unwrap(err), ErrParseRangeSet)
		}
	}
}
