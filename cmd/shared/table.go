// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package shared

import (
	"os"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	// columnPadding is the horizontal padding every cell gets, one column on each side
	columnPadding = 2
	// breakpoints are the extra characters a cell may wrap on. lipgloss only ever breaks on whitespace and hyphens, which chops a path like /var/lib/grendel/images/rocky9/vmlinuz mid-name
	breakpoints = "/"
)

// tableBorder is the border the tables are drawn with, borderCells assumes its runes are a single cell wide
var tableBorder = lipgloss.ASCIIBorder()

// Column describes a single column of a list command's table output
type Column struct {
	// Name is the header text and the value matched against the --filter flag
	Name string
	// Width pins the column to a fixed width, padding included, instead of letting the table size it. Zero lets the table pick the width
	Width int
	// MinWidth keeps the column from being squeezed below this width, padding included, while leaving it free to grow. Without one a column can be squeezed down to a single character, which is only reached once every other column is that narrow too
	MinWidth int
	// Hidden drops the column from the output. Columns named in the --filter flag are hidden on top of this
	Hidden bool
}

// TableOptions are the display flags the list commands share, see RegisterTableFlags
type TableOptions struct {
	// Width fits the table to this many columns instead of to the terminal. It takes precedence over NoWrap and applies to piped output too, where there is no terminal width to go by
	Width int
	// NoWrap leaves the columns at their natural width instead of fitting them to the terminal, long rows run past the edge
	NoWrap bool
	// NoBorder drops the box around the table and the rules between the columns
	NoBorder bool
	// BorderRows draws a rule between every row even when none of them wrap. By default the rules appear only once a row spans more than one line, where they are what tells the rows apart
	BorderRows bool
}

// RegisterTableFlags adds the shared table flags to a list command and returns the options they write into
func RegisterTableFlags(command *cobra.Command) *TableOptions {
	options := &TableOptions{}

	command.Flags().IntVar(&options.Width, "width", 0, "fit the table to this many columns instead of the terminal width")
	command.Flags().BoolVar(&options.NoWrap, "no-wrap", false, "don't fit the table to the terminal width")
	command.Flags().BoolVar(&options.NoBorder, "no-border", false, "don't draw the table border")
	command.Flags().BoolVar(&options.BorderRows, "border-rows", false, "always draw a rule between rows, not just when a row wraps")

	return options
}

// Table renders the output of a list command, sizing the columns to the terminal when stdout is a tty
type Table struct {
	columns []Column
	rows    [][]string
	options TableOptions
	empty   string
}

// NewTable builds a table from the full column list, hiding any column named in filter
func NewTable(columns []Column, filter []string) *Table {
	t := &Table{columns: make([]Column, 0, len(columns))}

	for _, column := range columns {
		if slices.Contains(filter, column.Name) {
			column.Hidden = true
		}

		t.columns = append(t.columns, column)
	}

	return t
}

// Options applies the display flags of a list command
func (t *Table) Options(options TableOptions) *Table {
	t.options = options

	return t
}

// Empty sets the string a cell with no value falls back to. It is substituted before the columns are measured, so a column that is empty all the way down is still sized and kept readable rather than collapsing
func (t *Table) Empty(value string) *Table {
	t.empty = value

	return t
}

// fillEmpty substitutes the placeholder into the cells that have no value
func (t *Table) fillEmpty() {
	if t.empty == "" {
		return
	}

	for _, row := range t.rows {
		for i, cell := range row {
			if cell == "" {
				row[i] = t.empty
			}
		}
	}
}

// AppendRow adds a row to the table. Cells line up with the full column list passed to NewTable, the hidden ones are dropped here. A cell without a column is dropped too, since the table can only size the columns it was told about and lipgloss would draw the surplus as an unnamed column past the right edge
func (t *Table) AppendRow(cells ...string) {
	row := make([]string, 0, len(cells))

	for i, cell := range cells {
		if i >= len(t.columns) {
			break
		}

		if t.columns[i].Hidden {
			continue
		}

		row = append(row, cell)
	}

	t.rows = append(t.rows, row)
}

// visible returns the columns that are actually rendered, the index into it is what the table hands to StyleFunc
func (t *Table) visible() []Column {
	columns := make([]Column, 0, len(t.columns))

	for _, column := range t.columns {
		if !column.Hidden {
			columns = append(columns, column)
		}
	}

	return columns
}

// ColumnCompletion completes the --filter flag of a list command with the names of its columns, leaving out the ones already given.
//
// --filter is a string slice, so the shell hands over the whole comma separated word and matches what comes back against all of it. Every suggestion has to carry the values already typed, otherwise nothing matches from the first comma on
func ColumnCompletion(columns []Column) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var prefix string
		if i := strings.LastIndex(toComplete, ","); i >= 0 {
			prefix = toComplete[:i+1]
		}

		chosen := strings.Split(prefix, ",")

		names := make([]string, 0, len(columns))
		for _, column := range columns {
			if slices.Contains(chosen, column.Name) {
				continue
			}

			names = append(names, prefix+column.Name)
		}

		return names, cobra.ShellCompDirectiveNoFileComp
	}
}

// naturalWidths returns the width of the widest cell of each visible column, the header included
func (t *Table) naturalWidths(columns []Column) []int {
	widths := make([]int, len(columns))

	for i, column := range columns {
		widths[i] = lipgloss.Width(column.Name)
	}

	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(widths) {
				widths[i] = max(widths[i], lipgloss.Width(cell))
			}
		}
	}

	return widths
}

// floor is the width a column is never squeezed below, padding included
func floor(column Column) int {
	return max(column.MinWidth, columnPadding+1)
}

// distribute returns the width of each column, padding included. Every column starts at the width its content wants and the widest one gives a cell back until the table fits, which levels the squeezed columns off at a common width and leaves the ones already narrower than that untouched.
//
// lipgloss does this by shrinking whichever column sits furthest above its own median cell instead, so a column that is empty in most rows and long in one gets flattened while a uniformly wide column keeps its space.
func (t *Table) distribute(columns []Column, width int) []int {
	natural := t.naturalWidths(columns)

	widths := make([]int, len(columns))
	total := 0

	for i, column := range columns {
		widths[i] = natural[i] + columnPadding
		if column.Width > 0 {
			widths[i] = column.Width
		}

		total += widths[i]
	}

	if width <= 0 {
		return widths
	}

	available := width
	if !t.options.NoBorder {
		available -= borderCells(len(columns))
	}

	for total > available {
		widest, index := 0, -1

		for i, column := range columns {
			// a column given an explicit width keeps it, the rest of the table absorbs the squeeze
			if column.Width > 0 || widths[i] <= floor(column) {
				continue
			}

			if widths[i] > widest {
				widest, index = widths[i], i
			}
		}

		// everything left is at its floor, the table cannot get any narrower
		if index < 0 {
			break
		}

		widths[index]--
		total--
	}

	return widths
}

func (t *Table) build(columns []Column) *table.Table {
	headers := make([]string, 0, len(columns))
	for _, column := range columns {
		headers = append(headers, column.Name)
	}

	return table.New().
		Headers(headers...).
		Rows(t.rows...).
		Border(tableBorder).
		BorderRow(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			style := lipgloss.NewStyle().Padding(0, 1)

			if col < len(columns) && columns[col].Width > 0 {
				style = style.Width(columns[col].Width)
			}

			if row == table.HeaderRow {
				return style.Bold(true)
			}

			return style
		})
}

// multiline reports whether any cell spans more than one line, which is when a rule between the rows earns the line it costs
func (t *Table) multiline() bool {
	for _, row := range t.rows {
		for _, cell := range row {
			if strings.Contains(cell, "\n") {
				return true
			}
		}
	}

	return false
}

// borderCells is what the border costs a row: the two outer edges plus a separator between each pair of columns
func borderCells(columns int) int {
	return columns + 1
}

// finish wraps the cells to the widths distribute settled on, so that long paths break on their separators instead of mid-name, then draws the table with the borders the command asked for. lipgloss hardcodes an empty breakpoint set, so wrapping the cells ourselves is the only way in
func (t *Table) finish(columns []Column, widths []int) string {
	// t.rows is already free of the hidden cells, so this skips AppendRow
	wrapped := &Table{columns: t.columns, options: t.options, rows: make([][]string, 0, len(t.rows))}

	for _, row := range t.rows {
		cells := make([]string, len(row))

		for i, cell := range row {
			if i < len(widths) && widths[i] > columnPadding {
				cell = ansi.Wrap(cell, widths[i]-columnPadding, breakpoints)
			}

			cells[i] = cell
		}

		wrapped.rows = append(wrapped.rows, cells)
	}

	sized := make([]Column, len(columns))
	copy(sized, columns)

	for i := range sized {
		sized[i].Width = widths[i]
	}

	final := wrapped.build(sized)

	switch {
	case t.options.NoBorder:
		final = final.BorderTop(false).BorderBottom(false).BorderLeft(false).BorderRight(false).BorderColumn(false).BorderHeader(false).BorderRow(false)
	case t.options.BorderRows || wrapped.multiline():
		final = final.BorderRow(true)
	}

	return final.Render()
}

// Render returns the table, wrapped to the terminal width only when it would otherwise overflow. Piped output is never constrained since term.GetSize fails on anything that isn't a tty
func (t *Table) Render() string {
	// an explicit width is honoured whether or not stdout is a terminal
	if t.options.Width > 0 {
		return t.render(t.options.Width)
	}

	// --no-wrap asks for the natural widths, which is what a width of zero means here
	if t.options.NoWrap {
		return t.render(0)
	}

	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 0
	}

	return t.render(width)
}

// render lays out the table within width, a width of zero or less leaves the columns at their natural size
func (t *Table) render(width int) string {
	// ahead of every measurement, an empty cell and its placeholder are not the same width
	t.fillEmpty()

	columns := t.visible()

	return t.finish(columns, t.distribute(columns, width))
}

// Print writes the table to stdout, downsampling the colors to what the terminal supports
func (t *Table) Print() {
	lipgloss.Println(t.Render())
}
