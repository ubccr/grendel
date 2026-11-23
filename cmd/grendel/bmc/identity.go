// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/ubccr/grendel/pkg/client"
)

// location is where a node sits, as its bmc records it. The generated client has one location type per response it appears in, all with these fields, so they convert to it
type location struct {
	Datacenter string
	Rack       string
	RackUnit   client.OptNilInt
	Room       string
	Row        string
}

// formatLocation shows a location as rack and unit first, since that is what finds the machine, with the wider place after it
func formatLocation(loc location) string {
	var head []string
	if loc.Rack != "" {
		head = append(head, loc.Rack)
	}
	if unit, ok := loc.RackUnit.Get(); ok {
		head = append(head, fmt.Sprintf("U%d", unit))
	}
	if loc.Row != "" {
		head = append(head, "row "+loc.Row)
	}

	var place []string
	for _, p := range []string{loc.Datacenter, loc.Room} {
		if p != "" {
			place = append(place, p)
		}
	}

	out := strings.Join(head, " ")
	switch {
	case len(place) == 0:
		return out
	case out == "":
		return strings.Join(place, "/")
	}
	return out + " (" + strings.Join(place, "/") + ")"
}

// changedCell marks a new value that differs from the current one, so the rows a run touches stand out
func changedCell(value string, changed bool) string {
	if changed {
		return text.FgHiYellow.Sprint(value)
	}
	return value
}

// serialColumns names the headers recognised as the serial and asset tag columns, compared after normalizeHeader
var (
	serialHeaders = []string{"serial", "serialnumber", "servicetag", "sn"}
	tagHeaders    = []string{"assettag", "tag", "asset"}
)

func normalizeHeader(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	return strings.NewReplacer(" ", "", "_", "", "-", "", "\ufeff", "").Replace(h)
}

// parseAssetCSV reads serial to asset tag rows. The columns are found by header, by the names given, or are the first two when the file has no header row. A serial listed twice is an error even with the same tag, since it usually means rows were pasted twice from different sources
func parseAssetCSV(r io.Reader, serialColumn, tagColumn string) (map[string]string, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true

	out := map[string]string{}
	serialIdx, tagIdx := -1, -1
	for first := true; ; first = false {
		row, err := cr.Read()
		if errors.Is(err, io.EOF) {
			if first {
				return nil, errors.New("csv is empty")
			}
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading csv: %w", err)
		}

		if first {
			var header bool
			serialIdx, tagIdx, header = findColumns(row, serialColumn, tagColumn)
			switch {
			case header && (serialIdx < 0 || tagIdx < 0):
				return nil, fmt.Errorf("csv header names only one of the serial and asset tag columns, pick them with --serial-column and --tag-column. The header is: %s", strings.Join(row, ", "))
			case !header && (serialColumn != "" || tagColumn != ""):
				return nil, fmt.Errorf("csv header has no %q or %q column, it is: %s", serialColumn, tagColumn, strings.Join(row, ", "))
			case !header:
				serialIdx, tagIdx = 0, 1
			}
			if header {
				continue
			}
		}

		// the reader skips blank lines and a quoted field can span several, so a row's number is not its line in the file
		line, _ := cr.FieldPos(0)
		if len(row) <= max(serialIdx, tagIdx) {
			if len(row) == 1 && strings.TrimSpace(row[0]) == "" {
				continue
			}
			return nil, fmt.Errorf("csv line %d has %d column(s), expected at least %d", line, len(row), max(serialIdx, tagIdx)+1)
		}
		serial := strings.TrimSpace(row[serialIdx])
		tag := strings.TrimSpace(row[tagIdx])
		if serial == "" {
			continue
		}
		if prev, ok := out[serial]; ok {
			if prev != tag {
				return nil, fmt.Errorf("csv line %d: serial %s is listed twice, with asset tags %q and %q", line, serial, prev, tag)
			}
			return nil, fmt.Errorf("csv line %d: serial %s is listed twice", line, serial)
		}
		out[serial] = tag
	}

	if len(out) == 0 {
		return nil, errors.New("csv has no rows with a serial number")
	}
	return out, nil
}

// findColumns locates the serial and tag columns in a header row. header is false when the row matched neither, and is then data
func findColumns(first []string, serialColumn, tagColumn string) (serialIdx, tagIdx int, header bool) {
	serialNames, tagNames := serialHeaders, tagHeaders
	if serialColumn != "" {
		serialNames = []string{normalizeHeader(serialColumn)}
	}
	if tagColumn != "" {
		tagNames = []string{normalizeHeader(tagColumn)}
	}

	serialIdx, tagIdx = -1, -1
	for i, h := range first {
		n := normalizeHeader(h)
		for _, s := range serialNames {
			if n == s && serialIdx < 0 {
				serialIdx = i
			}
		}
		for _, s := range tagNames {
			if n == s && tagIdx < 0 {
				tagIdx = i
			}
		}
	}
	return serialIdx, tagIdx, serialIdx >= 0 || tagIdx >= 0
}
