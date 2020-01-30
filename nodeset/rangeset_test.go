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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testRangeSet struct {
	test   string
	result string
	length int
}

func TestRangeSetSimple(t *testing.T) {
	tests := []testRangeSet{
		testRangeSet{"0", "0", 1},
		testRangeSet{"1", "1", 1},
		testRangeSet{"0-2", "0-2", 3},
		testRangeSet{"1-3", "1-3", 3},
		testRangeSet{"1-3,4-6", "1-6", 6},
		testRangeSet{"1-3,4-6,7-10", "1-10", 10},
		testRangeSet{"0001-0010", "0001-0010", 10},
	}

	for _, rstest := range tests {
		r1, err := NewRangeSet(rstest.test)
		assert.Nil(t, err)
		assert.Equal(t, rstest.result, r1.String())
		assert.Equal(t, rstest.length, r1.Len())
	}
}

func TestRangeSetSimpleStep(t *testing.T) {

	tests := []testRangeSet{
		testRangeSet{"0-4/2", "0,2,4", 3},
		testRangeSet{"1-4/2", "1,3", 2},
		testRangeSet{"1-4/3", "1,4", 2},
		testRangeSet{"1-4/4", "1", 1},
	}

	for _, rstest := range tests {
		r1, err := NewRangeSet(rstest.test)
		assert.Nil(t, err)
		assert.Equal(t, rstest.result, r1.String())
		assert.Equal(t, rstest.length, r1.Len())
	}
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
			assert.Equal(t, ErrParseRangeSet, errors.Unwrap(err))
		}
	}
}

func TestRangeSetSuperset(t *testing.T) {
	r1, err := NewRangeSet("1-100,102,105-242,800")
	assert.Nil(t, err)
	assert.Equal(t, 240, r1.Len())

	r2, err := NewRangeSet("3-98,140-199,800")
	assert.Nil(t, err)
	assert.Equal(t, 157, r2.Len())
	assert.True(t, r1.Superset(r1))
	assert.True(t, r1.Superset(r2))
	assert.True(t, r2.Subset(r1))

	r3, err := NewRangeSet("3-98,140-199,243,800")
	assert.Nil(t, err)
	assert.Equal(t, 158, r3.Len())
	assert.False(t, r1.Superset(r3))
}

type testRangeSetND struct {
	test   []string
	result string
	length int
}

func TestRangeSetND(t *testing.T) {

	tests := []testRangeSetND{
		testRangeSetND{[]string{"0-10"}, "0-10", 11},
		testRangeSetND{[]string{"0-10/2", "01-02"}, "0,2,4,6,8,10; 01-02", 12},
		testRangeSetND{[]string{"008-009", "0-10/2", "01-02"}, "008-009; 0,2,4,6,8,10; 01-02", 24},
	}

	for _, rstest := range tests {
		r1, err := NewRangeSetND(rstest.test)
		assert.Nil(t, err)
		assert.Equal(t, rstest.result, r1.String())
		assert.Equal(t, rstest.length, r1.Len())
	}
}

func TestRangeSetNDSuperset(t *testing.T) {
	r1, err := NewRangeSetND([]string{"0-10", "40-60"})
	assert.Nil(t, err)
	assert.True(t, r1.Superset(r1))
	assert.True(t, r1.Subset(r1))

	r2, err := NewRangeSetND([]string{"0-10", "40-60"})
	assert.Nil(t, err)
	assert.True(t, r2.Subset(r1))
	assert.True(t, r1.Subset(r2))
	assert.True(t, r2.Superset(r1))
	assert.True(t, r1.Superset(r2))

	r1, err = NewRangeSetND([]string{"0-10", "40-60"})
	assert.Nil(t, err)

	r2, err = NewRangeSetND([]string{"4", "40-41"})
	assert.False(t, r1.Subset(r2))
	assert.True(t, r2.Subset(r1))
	assert.True(t, r1.Superset(r2))
}
