// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// tomlDoc writes options as a config file. value gives what to write for an option, or false to leave it out, and sample describes each option and comments out all but the ones a packaged install sets.
func tomlDoc(opts []*Option, value func(o *Option) (any, bool), sample bool) string {
	describe := sample
	type entry struct {
		o *Option
		v any
	}
	tables := map[string][]entry{}
	var order []string
	for _, o := range opts {
		v, ok := value(o)
		if !ok {
			continue
		}
		table := ""
		if o.Merged() {
			table = o.Key
		} else if i := strings.LastIndex(o.Key, "."); i >= 0 {
			table = o.Key[:i]
		}
		if _, seen := tables[table]; !seen {
			order = append(order, table)
		}
		tables[table] = append(tables[table], entry{o, v})
	}
	// keys outside any table have to come before the first table header
	slices.SortStableFunc(order, func(a, b string) int {
		if (a == "") != (b == "") {
			if a == "" {
				return -1
			}
			return 1
		}
		return 0
	})

	mark := func(o *Option) string {
		if sample && o.Packaged == nil {
			return "#"
		}
		return ""
	}
	var b strings.Builder
	lines := func(o *Option, m map[string]any) {
		for _, name := range slices.Sorted(maps.Keys(m)) {
			b.WriteString(mark(o) + tomlLine(name, m[name]) + "\n")
		}
	}
	for _, table := range order {
		entries := tables[table]
		first := entries[0]
		if first.o.Kind == KindGroup {
			if describe {
				comment(&b, first.o)
			}
			group := first.v.(map[string]any)
			for _, name := range slices.Sorted(maps.Keys(group)) {
				fmt.Fprintf(&b, "%s[%s.%s]\n", mark(first.o), table, name)
				lines(first.o, group[name].(map[string]any))
				b.WriteString("\n")
			}
			continue
		}

		if table != "" {
			if first.o.Merged() && describe {
				comment(&b, first.o)
			}
			fmt.Fprintf(&b, "[%s]\n", table)
		}
		for _, e := range entries {
			if e.o.Merged() {
				lines(e.o, anyMap(e.v))
				continue
			}
			if describe {
				comment(&b, e.o)
			}
			b.WriteString(mark(e.o) + tomlLine(e.o.Key[strings.LastIndex(e.o.Key, ".")+1:], e.v) + "\n")
			if describe {
				b.WriteString("\n")
			}
		}
		if !describe || first.o.Merged() {
			b.WriteString("\n")
		}
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func comment(b *strings.Builder, o *Option) {
	fmt.Fprintf(b, "# %s. %s\n", o.Summary, o.Help)
	if len(o.Enum) > 0 {
		fmt.Fprintf(b, "# One of: %s\n", strings.Join(o.Enum, ", "))
	}
	if o.Unit != "" {
		fmt.Fprintf(b, "# In %s.\n", o.Unit)
	}
	if o.InheritFrom != "" {
		fmt.Fprintf(b, "# Follows %s unless it is set.\n", o.InheritFrom)
	}
	if o.Deprecated != "" {
		fmt.Fprintf(b, "# Deprecated: %s\n", o.Deprecated)
	}
}

func anyMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

// tomlLine writes one key and its value on a single line, with a list of records or a nested map written inline
func tomlLine(name string, v any) string {
	var b strings.Builder
	enc := toml.NewEncoder(&b)
	enc.SetTablesInline(true)
	if err := enc.Encode(map[string]any{name: storedForm(v)}); err != nil {
		return "# " + name + ": " + err.Error()
	}
	return strings.TrimRight(b.String(), "\n")
}
