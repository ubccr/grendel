// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Many of the tests written here were adopted from ClusterShell
// https://github.com/cea-hpc/clustershell

package nodeset

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		require.NoError(t, err)
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
		require.NoError(t, err)
		assert.Equal(t, rstest.result, r1.String())
		assert.Equal(t, rstest.length, r1.Len())
	}
}

func TestRangeSetBadSyntax(t *testing.T) {
	badSyntax := []string{
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

func TestRangeSetEquality(t *testing.T) {
	r1, err := NewRangeSet("")
	require.NoError(t, err)
	r2, err := NewRangeSet("")
	require.NoError(t, err)
	assert.True(t, r1.Equal(r2))

	r1, err = NewRangeSet("2-5")
	require.NoError(t, err)
	r2, err = NewRangeSet("1,2,3,4")
	require.NoError(t, err)
	assert.False(t, r1.Equal(r2))

	r1, err = NewRangeSet("1-5")
	require.NoError(t, err)
	r2, err = NewRangeSet("1,2,3,4,5")
	require.NoError(t, err)
	assert.True(t, r1.Equal(r2))
}

func TestRangeSetIntersectSimple(t *testing.T) {
	r1, err := NewRangeSet("4-34")
	require.NoError(t, err)
	r2, err := NewRangeSet("27-42")
	require.NoError(t, err)
	r1.InPlaceIntersection(r2)
	assert.Equal(t, "27-34", r1.String())
	assert.Equal(t, 8, r1.Len())

	r1, err = NewRangeSet("2-450,654-700,800")
	require.NoError(t, err)
	r2, err = NewRangeSet("500-502,690-820,830-840,900")
	require.NoError(t, err)
	r1.InPlaceIntersection(r2)
	assert.Equal(t, "690-700,800", r1.String())
	assert.Equal(t, 12, r1.Len())

	r1, err = NewRangeSet("2-450,654-700,800")
	require.NoError(t, err)
	r3 := r1.Intersection(r2)
	assert.Equal(t, "690-700,800", r3.String())
	assert.Equal(t, 12, r3.Len())

	r1, err = NewRangeSet("")
	require.NoError(t, err)
	r3 = r1.Intersection(r2)
	assert.Equal(t, "", r3.String())
	assert.Equal(t, 0, r3.Len())
}

func TestRangeSetSymmetricDifference(t *testing.T) {
	r1, err := NewRangeSet("4,7-33")
	require.NoError(t, err)
	r2, err := NewRangeSet("8-34")
	require.NoError(t, err)
	r1.InPlaceSymmetricDifference(r2)
	assert.Equal(t, "4,7,34", r1.String())
	assert.Equal(t, 3, r1.Len())

	r1, err = NewRangeSet("4,7-33")
	require.NoError(t, err)
	r3 := r1.SymmetricDifference(r2)
	assert.Equal(t, "4,7,34", r3.String())
	assert.Equal(t, 3, r3.Len())

	r1, err = NewRangeSet("5,7,10-12,33-50")
	require.NoError(t, err)
	r2, err = NewRangeSet("8-34")
	require.NoError(t, err)
	r1.InPlaceSymmetricDifference(r2)
	assert.Equal(t, "5,7-9,13-32,35-50", r1.String())
	assert.Equal(t, 40, r1.Len())

	r1, err = NewRangeSet("8-30")
	require.NoError(t, err)
	r2, err = NewRangeSet("31-40")
	require.NoError(t, err)
	r1.InPlaceSymmetricDifference(r2)
	assert.Equal(t, "8-40", r1.String())
	assert.Equal(t, 33, r1.Len())

	r1, err = NewRangeSet("8-30")
	require.NoError(t, err)
	r2, err = NewRangeSet("8-30")
	require.NoError(t, err)
	r1.InPlaceSymmetricDifference(r2)
	assert.Equal(t, "", r1.String())
	assert.Equal(t, 0, r1.Len())
}

func TestRangeSetSuperset(t *testing.T) {
	r1, err := NewRangeSet("1-100,102,105-242,800")
	require.NoError(t, err)
	assert.Equal(t, 240, r1.Len())

	r2, err := NewRangeSet("3-98,140-199,800")
	require.NoError(t, err)
	assert.Equal(t, 157, r2.Len())
	assert.True(t, r1.Superset(r1))
	assert.True(t, r1.Superset(r2))
	assert.True(t, r2.Subset(r1))

	r3, err := NewRangeSet("3-98,140-199,243,800")
	require.NoError(t, err)
	assert.Equal(t, 158, r3.Len())
	assert.False(t, r1.Superset(r3))
}

func TestRangeSetIterator(t *testing.T) {
	rgs, err := NewRangeSet("011,003,005-008,001,004")
	require.NoError(t, err)

	smatches := []string{"001", "003", "004", "005", "006", "007", "008", "011"}
	slist := []string{}
	for _, rg := range rgs.Strings() {
		slist = append(slist, rg)
	}
	assert.Equal(t, smatches, slist)

	imatches := []int{1, 3, 4, 5, 6, 7, 8, 11}
	ilist := []int{}
	for _, rg := range rgs.Ints() {
		ilist = append(ilist, rg)
	}
	assert.Equal(t, imatches, ilist)
}

type testRangeSetND struct {
	test   [][]string
	result string
	length int
}

func TestRangeSetNDSimpleAndFold(t *testing.T) {

	tests := []testRangeSetND{
		testRangeSetND{[][]string{[]string{"0-10"}}, "0-10\n", 11},
		testRangeSetND{[][]string{[]string{"0-10/2", "01-02"}}, "0,2,4,6,8,10; 01-02\n", 12},
		testRangeSetND{[][]string{[]string{"008-009", "0-10/2", "01-02"}}, "008-009; 0,2,4,6,8,10; 01-02\n", 24},
		testRangeSetND{[][]string{[]string{"0-10"}, []string{"40-60"}}, "0-10,40-60\n", 32},
		testRangeSetND{[][]string{[]string{"0-2", "1-2"}, []string{"10", "3-5"}}, "0-2; 1-2\n10; 3-5\n", 9},
		testRangeSetND{[][]string{[]string{"0-10", "1-2"}, []string{"5-15,40-60", "1-3"}, []string{"0-4", "3"}}, "0-15,40-60; 1-3\n", 111},
		testRangeSetND{[][]string{[]string{"0-10"}, []string{"11-60"}}, "0-60\n", 61},
		testRangeSetND{[][]string{[]string{"0-2", "1-2"}, []string{"3", "1-2"}}, "0-3; 1-2\n", 8},
		testRangeSetND{[][]string{[]string{"3", "1-3"}, []string{"0-2", "1-2"}}, "0-2; 1-2\n3; 1-3\n", 9},
		testRangeSetND{[][]string{[]string{"0-2", "1-2"}, []string{"3", "1-3"}}, "0-2; 1-2\n3; 1-3\n", 9},
		testRangeSetND{[][]string{[]string{"0-2", "1-2"}, []string{"1-3", "1-3"}}, "1-2; 1-3\n0,3; 1-2\n3; 3\n", 11},
		testRangeSetND{[][]string{[]string{"0-2", "1-2", "0-4"}, []string{"3", "1-2", "0-5"}}, "0-2; 1-2; 0-4\n3; 1-2; 0-5\n", 42},
		testRangeSetND{[][]string{[]string{"0-2", "1-2", "0-4"}, []string{"1-3", "1-3", "0-4"}}, "1-2; 1-3; 0-4\n0,3; 1-2; 0-4\n3; 3; 0-4\n", 55},
		testRangeSetND{[][]string{[]string{"0-100", "50-200"}, []string{"2-101", "49"}}, "0-100; 50-200\n2-101; 49\n", 15351},
	}

	for _, rstest := range tests {
		r1, err := NewRangeSetND(rstest.test)
		require.NoError(t, err)
		assert.Equal(t, rstest.result, r1.String())
		assert.Equal(t, rstest.length, r1.Len())
	}
}
