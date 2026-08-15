// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package cmd

import (
	"slices"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

var testColumns = []Column{
	{Name: "Name"},
	{Name: "Kernel"},
	{Name: "Command Line"},
	{Name: "Verify"},
}

func newTestTable() *Table {
	t := NewTable(testColumns, nil)
	t.AppendRow("rocky9", "/var/lib/grendel/images/rocky9/vmlinuz-5.14.0-503.el9.x86_64", "console=ttyS0,115200 root=live:http://10.0.0.1/rocky9.squashfs", "true")
	t.AppendRow("rocky10", "/var/lib/grendel/images/rocky10/vmlinuz-6.12.0-55.el10.x86_64", "console=ttyS0,115200 root=live:http://10.0.0.1/rocky10.squashfs", "true")

	return t
}

func TestTableRenderUnconstrained(t *testing.T) {
	out := newTestTable().render(0)

	if lipgloss.Width(out) <= 100 {
		t.Errorf("expected the natural table to be wider than 100 columns, got %d", lipgloss.Width(out))
	}

	for _, want := range []string{"Name", "Kernel", "Command Line", "Verify", "rocky9", "rocky10"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}
}

func TestTableRenderFitsWidth(t *testing.T) {
	for _, width := range []int{30, 40, 60, 80, 100, 120} {
		out := newTestTable().render(width)

		if got := lipgloss.Width(out); got > width {
			t.Errorf("render(%d) produced a table %d columns wide:\n%s", width, got, out)
		}

		for _, line := range strings.Split(out, "\n") {
			if got := lipgloss.Width(line); got > width {
				t.Errorf("render(%d) produced a line %d columns wide: %q", width, got, line)
			}
		}
	}
}

func TestTableKeepsNarrowHeaders(t *testing.T) {
	// Verify is no wider than its own header, so shrinking it can only truncate the header. Kernel and Command Line have room to give and should absorb the squeeze instead
	out := newTestTable().render(80)

	if !strings.Contains(out, "Verify") {
		t.Errorf("expected the Verify header to survive at 80 columns:\n%s", out)
	}

	if !strings.Contains(out, "Name") {
		t.Errorf("expected the Name header to survive at 80 columns:\n%s", out)
	}
}

func TestTableWrapsOnPathSeparators(t *testing.T) {
	// 20 wide leaves 18 for the content, so the path only fits by wrapping
	tbl := NewTable([]Column{{Name: "Kernel", Width: 20}}, nil)
	tbl.AppendRow("/var/lib/grendel/images/rocky9/vmlinuz")

	out := tbl.render(0)

	if !strings.Contains(out, "/var/lib/grendel/") {
		t.Errorf("expected the path to break after a separator:\n%s", out)
	}

	// what lipgloss does on its own, it only ever breaks on whitespace and hyphens
	if strings.Contains(out, "/var/lib/grendel/i") {
		t.Errorf("expected no break in the middle of a path element:\n%s", out)
	}
}

func TestTableRenderSkipsFittingWideTerminal(t *testing.T) {
	natural := newTestTable().render(0)

	// a terminal wider than the table must not stretch it to fill the width
	if out := newTestTable().render(500); out != natural {
		t.Errorf("expected the natural table for a 500 column terminal, got:\n%s", out)
	}
}

// cellWidths returns the rendered width of each column, padding included
func cellWidths(t *testing.T, out string) []int {
	t.Helper()

	border, _, _ := strings.Cut(out, "\n")
	border = strings.TrimPrefix(border, tableBorder.TopLeft)
	border = strings.TrimSuffix(border, tableBorder.TopRight)

	widths := []int{}
	for _, segment := range strings.Split(border, tableBorder.MiddleTop) {
		widths = append(widths, lipgloss.Width(segment))
	}

	return widths
}

func TestTableMinWidth(t *testing.T) {
	columns := []Column{{Name: "Name"}, {Name: "Kernel"}, {Name: "Command Line", MinWidth: 24}}

	tbl := NewTable(columns, nil)
	tbl.AppendRow("rocky9", "/var/lib/grendel/images/rocky9/vmlinuz-5.14.0-503.el9.x86_64", "console=ttyS0,115200 root=live:http://10.0.0.1/rocky9.squashfs")
	tbl.AppendRow("rocky10", "/var/lib/grendel/images/rocky10/vmlinuz-6.12.0-55.el10.x86_64", "console=ttyS0,115200")

	for _, width := range []int{60, 80, 100} {
		out := tbl.render(width)

		widths := cellWidths(t, out)
		if len(widths) != len(columns) {
			t.Fatalf("render(%d): measured %d columns, want %d:\n%s", width, len(widths), len(columns), out)
		}

		if widths[2] < 24 {
			t.Errorf("render(%d): Command Line is %d wide, want at least 24:\n%s", width, widths[2], out)
		}

		if lipgloss.Width(out) > width {
			t.Errorf("render(%d): table is %d wide:\n%s", width, lipgloss.Width(out), out)
		}
	}
}

func TestTableMinWidthIsAFloor(t *testing.T) {
	// 30 columns is not enough for any of this, so everything is pushed down to what it is allowed to give
	columns := []Column{{Name: "Name"}, {Name: "Kernel"}, {Name: "Command Line"}, {Name: "Verify", MinWidth: 8}}

	tbl := NewTable(columns, nil)
	tbl.AppendRow("rocky9", "/var/lib/grendel/images/rocky9/vmlinuz-5.14.0-503.el9.x86_64", "console=ttyS0,115200 root=live:http://10.0.0.1/rocky9.squashfs", "true")

	out := tbl.render(30)

	widths := cellWidths(t, out)
	if len(widths) != len(columns) {
		t.Fatalf("measured %d columns, want %d:\n%s", len(widths), len(columns), out)
	}

	if widths[3] != 8 {
		t.Errorf("Verify is %d wide, want its minimum of 8:\n%s", widths[3], out)
	}

	// the columns without a minimum went below it, so the minimum is what held Verify up
	if widths[1] >= widths[3] {
		t.Errorf("expected Kernel to give up more width than the column with a minimum, got %d vs %d:\n%s", widths[1], widths[3], out)
	}
}

func TestTableDistributesEvenly(t *testing.T) {
	// Initrd is empty in every row but one. lipgloss sizes on the median cell, which flattens a column like this while a uniformly wide one keeps its space
	columns := []Column{{Name: "Name"}, {Name: "Kernel"}, {Name: "Initrd"}}

	tbl := NewTable(columns, nil).Empty("-")
	tbl.AppendRow("rocky9", "/var/lib/grendel/images/rocky9/vmlinuz-5.14.0-503.el9.x86_64", "")
	tbl.AppendRow("rocky10", "/var/lib/grendel/images/rocky10/vmlinuz-6.12.0-55.el10.x86_64", "/var/lib/grendel/images/rocky10/initramfs.img, /var/lib/grendel/images/rocky10/microcode.img")
	tbl.AppendRow("alma9", "/var/lib/grendel/images/alma9/vmlinuz-5.14.0-503.el9.x86_64", "")

	out := tbl.render(100)

	widths := cellWidths(t, out)
	if len(widths) != len(columns) {
		t.Fatalf("measured %d columns, want %d:\n%s", len(widths), len(columns), out)
	}

	// both are squeezed, so they should land within a cell of each other rather than one keeping five times the other
	if diff := widths[1] - widths[2]; diff > 1 || diff < -1 {
		t.Errorf("Kernel is %d wide and Initrd %d, want them level:\n%s", widths[1], widths[2], out)
	}
}

func TestTableNoWrap(t *testing.T) {
	natural := newTestTable().render(0)

	tbl := newTestTable().Options(TableOptions{NoWrap: true})

	// Render is what reads the option, render(width) is the fitting path it bypasses
	if out := tbl.Render(); out != natural {
		t.Errorf("expected --no-wrap to keep the natural widths:\n%s", out)
	}
}

func TestTableOptionWidth(t *testing.T) {
	// Render, not render, so this covers the option plumbing. Stdout is not a terminal under go test, which is exactly the case an explicit width has to survive
	for _, width := range []int{40, 60, 80} {
		out := newTestTable().Options(TableOptions{Width: width}).Render()

		if got := lipgloss.Width(out); got > width {
			t.Errorf("--width %d produced a table %d wide:\n%s", width, got, out)
		}

		for _, line := range strings.Split(out, "\n") {
			if got := lipgloss.Width(line); got > width {
				t.Errorf("--width %d produced a line %d wide: %q", width, got, line)
			}
		}
	}
}

func TestTableOptionWidthBeatsNoWrap(t *testing.T) {
	out := newTestTable().Options(TableOptions{Width: 60, NoWrap: true}).Render()

	if got := lipgloss.Width(out); got > 60 {
		t.Errorf("expected --width to win over --no-wrap, got a table %d wide:\n%s", got, out)
	}
}

func TestTableOptionWidthWithNoBorder(t *testing.T) {
	out := newTestTable().Options(TableOptions{Width: 80, NoBorder: true}).Render()

	if got := lipgloss.Width(out); got > 80 {
		t.Errorf("borderless table is %d wide, want at most 80:\n%s", got, out)
	}

	if got := lipgloss.Width(out); got < 80-columnPadding {
		t.Errorf("borderless table is only %d wide, want it to use the full 80:\n%s", got, out)
	}
}

func TestTableNoBorder(t *testing.T) {
	out := newTestTable().Options(TableOptions{NoBorder: true}).render(80)

	if strings.Contains(out, tableBorder.Left) || strings.Contains(out, tableBorder.TopLeft) {
		t.Errorf("expected no border runes in the output:\n%s", out)
	}

	if got := lipgloss.Width(out); got > 80 {
		t.Errorf("borderless table is %d wide, want at most 80:\n%s", got, out)
	}

	// the cells the border would have taken are spent on content instead
	if got := lipgloss.Width(out); got < 80-columnPadding {
		t.Errorf("borderless table is only %d wide, want it to use the full 80:\n%s", got, out)
	}

	for _, want := range []string{"Name", "Kernel", "rocky9", "rocky10"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}
}

// rules counts the horizontal rules in a rendered table
func rules(out string) int {
	count := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, tableBorder.TopLeft) {
			count++
		}
	}

	return count
}

func TestTableBorderRowsFollowWrapping(t *testing.T) {
	// nothing wraps at its natural width, so the rows are single lines and need no rules: top, under the header, bottom
	natural := newTestTable().render(0)
	if got := rules(natural); got != 3 {
		t.Errorf("expected 3 rules in an unwrapped table, got %d:\n%s", got, natural)
	}

	// squeezed to 80 the rows span several lines each, so a rule goes between them
	wrapped := newTestTable().render(80)
	if got := rules(wrapped); got != 4 {
		t.Errorf("expected 4 rules once the rows wrap, got %d:\n%s", got, wrapped)
	}
}

func TestTableBorderRowsAlways(t *testing.T) {
	ruled := newTestTable().Options(TableOptions{BorderRows: true}).render(0)

	if got := rules(ruled); got != 4 {
		t.Errorf("expected --border-rows to add a rule to an unwrapped table, got %d:\n%s", got, ruled)
	}
}

func TestTableBorderRowsWidth(t *testing.T) {
	ruled := newTestTable().Options(TableOptions{BorderRows: true}).render(80)

	if got := lipgloss.Width(ruled); got > 80 {
		t.Errorf("table is %d wide, want at most 80:\n%s", got, ruled)
	}
}

// rowCells returns the trimmed cells of the first rendered row containing marker
func rowCells(t *testing.T, out, marker string) []string {
	t.Helper()

	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, marker) {
			continue
		}

		parts := strings.Split(line, tableBorder.Left)
		if len(parts) < 3 {
			t.Fatalf("row %q has no cells:\n%s", marker, out)
		}

		cells := make([]string, 0, len(parts)-2)
		for _, part := range parts[1 : len(parts)-1] {
			cells = append(cells, strings.TrimSpace(part))
		}

		return cells
	}

	t.Fatalf("no row containing %q:\n%s", marker, out)

	return nil
}

func TestTableEmpty(t *testing.T) {
	columns := []Column{{Name: "Name"}, {Name: "Kernel"}, {Name: "Live Image"}}

	tbl := NewTable(columns, nil).Empty("-")
	tbl.AppendRow("rocky9", "/var/lib/grendel/images/rocky9/vmlinuz-5.14.0-503.el9.x86_64", "")
	tbl.AppendRow("rocky10", "/var/lib/grendel/images/rocky10/vmlinuz-6.12.0-55.el10.x86_64", "")

	out := tbl.render(0)

	if got := rowCells(t, out, "rocky9")[2]; got != "-" {
		t.Errorf("empty cell rendered as %q, want the placeholder:\n%s", got, out)
	}

	// substituting before the measuring keeps a column that is empty all the way down from collapsing
	widths := cellWidths(t, out)
	if len(widths) != len(columns) {
		t.Fatalf("measured %d columns, want %d:\n%s", len(widths), len(columns), out)
	}

	if want := lipgloss.Width("Live Image") + columnPadding; widths[2] != want {
		t.Errorf("empty column is %d wide, want %d:\n%s", widths[2], want, out)
	}

	if !strings.Contains(out, "Live Image") {
		t.Errorf("expected the empty column to keep its header:\n%s", out)
	}
}

func TestTableEmptyUnset(t *testing.T) {
	tbl := NewTable([]Column{{Name: "Name"}, {Name: "Live Image"}}, nil)
	tbl.AppendRow("rocky9", "")

	// without a placeholder the cell stays empty
	out := tbl.render(0)
	if got := rowCells(t, out, "rocky9")[1]; got != "" {
		t.Errorf("cell rendered as %q, want it left empty:\n%s", got, out)
	}
}

func TestTableRenderNoRows(t *testing.T) {
	out := NewTable(testColumns, nil).render(80)

	for _, want := range []string{"Name", "Kernel", "Command Line", "Verify"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in the header only output:\n%s", want, out)
		}
	}
}

func TestTableFilterHidesColumns(t *testing.T) {
	tbl := NewTable(testColumns, []string{"Kernel", "Verify"})
	tbl.AppendRow("rocky9", "/var/lib/grendel/images/rocky9/vmlinuz", "console=ttyS0", "true")

	out := tbl.render(0)

	for _, want := range []string{"Name", "Command Line", "rocky9", "console=ttyS0"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in output:\n%s", want, out)
		}
	}

	for _, notWant := range []string{"Kernel", "vmlinuz", "Verify", "true"} {
		if strings.Contains(out, notWant) {
			t.Errorf("found filtered %q in output:\n%s", notWant, out)
		}
	}
}

func TestTableFixedWidthColumn(t *testing.T) {
	columns := []Column{{Name: "Name"}, {Name: "Kernel", Width: 20}, {Name: "Verify"}}

	tbl := NewTable(columns, nil)
	tbl.AppendRow("rocky9", "/var/lib/grendel/images/rocky9/vmlinuz-5.14.0-503.el9.x86_64", "true")

	// the pinned column keeps its width whether the table is squeezed or not
	for _, width := range []int{0, 60} {
		out := tbl.render(width)

		found := false
		for _, line := range strings.Split(out, "\n") {
			cells := strings.Split(line, tableBorder.Left)
			if len(cells) < 4 {
				continue
			}
			if lipgloss.Width(cells[2]) != 20 {
				t.Errorf("render(%d): pinned column is %d wide, want 20:\n%s", width, lipgloss.Width(cells[2]), out)
			}
			found = true
		}

		if !found {
			t.Errorf("render(%d): no data rows found in output:\n%s", width, out)
		}
	}
}

func TestTableIgnoresSurplusCells(t *testing.T) {
	table := NewTable(testColumns, nil)
	table.AppendRow("rocky9", "vmlinuz", "console=ttyS0", "true", "surplus")

	out := table.render(80)

	if got := lipgloss.Width(out); got > 80 {
		t.Errorf("a row with more cells than columns widened the table to %d:\n%s", got, out)
	}

	if cells := rowCells(t, out, "rocky9"); len(cells) != len(testColumns) {
		t.Errorf("expected %d cells, got %d %q:\n%s", len(testColumns), len(cells), cells, out)
	}

	if strings.Contains(out, "surplus") {
		t.Errorf("cell without a column was rendered:\n%s", out)
	}
}

func TestColumnCompletion(t *testing.T) {
	complete := ColumnCompletion(testColumns)

	for _, tc := range []struct {
		toComplete string
		want       []string
	}{
		{"", []string{"Name", "Kernel", "Command Line", "Verify"}},
		// the shell matches against the whole word, so the values already typed have to come back with every suggestion
		{"Name,", []string{"Name,Kernel", "Name,Command Line", "Name,Verify"}},
		{"Name,Ker", []string{"Name,Kernel", "Name,Command Line", "Name,Verify"}},
		{"Name,Kernel,", []string{"Name,Kernel,Command Line", "Name,Kernel,Verify"}},
	} {
		got, directive := complete(nil, nil, tc.toComplete)

		if directive != cobra.ShellCompDirectiveNoFileComp {
			t.Errorf("complete(%q) directive = %v, want NoFileComp", tc.toComplete, directive)
		}

		if !slices.Equal(got, tc.want) {
			t.Errorf("complete(%q) = %q, want %q", tc.toComplete, got, tc.want)
		}
	}
}
