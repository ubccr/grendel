// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ubccr/grendel/pkg/client"
)

func TestParseAssetCSV(t *testing.T) {
	cases := []struct {
		name, csv, serialCol, tagCol string
		want                         map[string]string
		err                          string
	}{
		{name: "header", csv: "Service Tag,Asset_Tag\nABC123,UB-1\nDEF456,UB-2\n",
			want: map[string]string{"ABC123": "UB-1", "DEF456": "UB-2"}},
		{name: "header in another order with extra columns", csv: "rack,asset tag,notes,serial number\nf01,UB-1,x,ABC123\n",
			want: map[string]string{"ABC123": "UB-1"}},
		{name: "excel bom", csv: "\ufeffSerial,Tag\nABC123,UB-1\n",
			want: map[string]string{"ABC123": "UB-1"}},
		{name: "no header", csv: "ABC123,UB-1\nDEF456,UB-2\n",
			want: map[string]string{"ABC123": "UB-1", "DEF456": "UB-2"}},
		{name: "named columns", csv: "st,inventory\nABC123,UB-1\n", serialCol: "ST", tagCol: "Inventory",
			want: map[string]string{"ABC123": "UB-1"}},
		{name: "blank lines and serials are skipped", csv: "serial,tag\n\n,UB-0\nABC123,UB-1\n",
			want: map[string]string{"ABC123": "UB-1"}},
		{name: "conflicting duplicate", csv: "serial,tag\nABC123,UB-1\nABC123,UB-2\n", err: "listed twice, with asset tags"},
		{name: "duplicate", csv: "ABC123,UB-1\nABC123,UB-1\n", err: "line 2: serial ABC123 is listed twice"},
		{name: "header names one column", csv: "serial,inventory\nABC123,UB-1\n", err: "--tag-column"},
		{name: "named column missing", csv: "a,b\nABC123,UB-1\n", serialCol: "st", tagCol: "inv", err: "has no"},
		{name: "short row", csv: "serial,tag\nABC123\n", err: "line 2 has 1 column"},
		{name: "empty", csv: "", err: "empty"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseAssetCSV(strings.NewReader(tc.csv), tc.serialCol, tc.tagCol)
			if tc.err != "" {
				assert.ErrorContains(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFormatLocation(t *testing.T) {
	unit := client.NewOptNilInt(4)
	assert.Equal(t, "f01 U4 (VTC/Suite 3)", formatLocation("VTC", "Suite 3", "", "f01", unit))
	assert.Equal(t, "f01 U4 row A", formatLocation("", "", "A", "f01", unit))
	assert.Equal(t, "VTC", formatLocation("VTC", "", "", "", client.OptNilInt{}))
	assert.Equal(t, "", formatLocation("", "", "", "", client.OptNilInt{}))
}
