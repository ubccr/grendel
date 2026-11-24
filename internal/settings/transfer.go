// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

type ImportAction string

const (
	ImportReady ImportAction = "import"
	// ImportEntries is a map whose entries are imported only when named, since some may be secrets
	ImportEntries    ImportAction = "entries"
	ImportDefault    ImportAction = "default"
	ImportSecret     ImportAction = "secret"
	ImportExternal   ImportAction = "external"
	ImportClient     ImportAction = "client"
	ImportDeprecated ImportAction = "deprecated"
	ImportUnknown    ImportAction = "unknown"
	ImportInvalid    ImportAction = "invalid"
)

// File is a config file read on its own, which need not be the one this process runs with
type File struct {
	Path string
	// Keys lists every key the file sets, with the entries of a table flattened to dotted keys
	Keys   []string
	Lookup func(key string) (any, bool)
}

// ImportItem is one thing a config file sets, and what an import does with it
type ImportItem struct {
	Key    string
	Action ImportAction
	// Value is what the file sets, left out for a secret
	Value string
	Note  string
}

// ImportPlan is what importing a config file would store, and what the file has to keep
type ImportPlan struct {
	Items []ImportItem
	set   map[string]any
	keep  map[string]any
}

// PlanImport works out what importing f would do. entries names the entries to import from a map that may hold secrets, such as provision.extra_vars.
func PlanImport(f File, entries map[string][]string) (*ImportPlan, error) {
	mu.Lock()
	defer mu.Unlock()

	for key := range entries {
		if o, ok := registry[key]; !ok || !o.Merged() || !o.FileEntriesSecret {
			return nil, fmt.Errorf("%s is not a setting whose entries are imported by name", key)
		}
	}

	plan := &ImportPlan{set: map[string]any{}, keep: map[string]any{}}
	st := current.Load()
	known := map[string]bool{}
	for _, o := range ordered {
		raw, ok := f.Lookup(o.Key)
		if !ok {
			continue
		}
		known[o.Key] = true

		item := ImportItem{Key: o.Key}
		v, err := o.Coerce(raw)
		switch {
		case err != nil && (o.Secret || o.FileEntriesSecret):
			item.Action, item.Note = ImportInvalid, "invalid value"
		case err != nil:
			item.Action, item.Note = ImportInvalid, "invalid value: "+err.Error()
		case Text(v) == Text(o.Default):
			item.Action, item.Note = ImportDefault, "default value"
			if !o.Secret {
				item.Value = Text(v)
			}
		case o.Scope != ScopeServer:
			item.Action, item.Value, item.Note = ImportClient, Text(v), "config file only, read by the grendel command"
			plan.keep[o.Key] = v
		case o.Secret:
			item.Action, item.Note = ImportSecret, "config file only"
			plan.keep[o.Key] = v
		case o.External:
			item.Action, item.Value, item.Note = ImportExternal, Text(v), "config file only"
			plan.keep[o.Key] = v
		case o.Deprecated != "" && replaceable(o, f):
			// a value under an old key is imported under the key that replaced it
			item.Action, item.Value, item.Note = ImportReady, Text(v), "deprecated, imported as "+o.ReplacedBy
			plan.set[o.ReplacedBy] = v
		case o.Deprecated != "":
			item.Action, item.Value, item.Note = ImportDeprecated, Text(v), "deprecated, "+o.Deprecated
		case o.Merged() && o.FileEntriesSecret:
			file := anyMap(v)
			item.Value = strings.Join(entryKeys(v), ", ")
			merged := map[string]any{}
			if st != nil {
				maps.Copy(merged, anyMap(st.stored[o.Key]))
			}
			left := maps.Clone(file)
			for _, name := range entries[o.Key] {
				if _, ok := file[name]; !ok {
					return nil, fmt.Errorf("%s has no entry %s in %s", o.Key, name, f.Path)
				}
				merged[name] = file[name]
				delete(left, name)
			}
			if len(entries[o.Key]) > 0 {
				item.Action, item.Note = ImportReady, "only "+strings.Join(entries[o.Key], ", ")
				plan.set[o.Key] = merged
			} else {
				item.Action, item.Note = ImportEntries, "please manually import non-secrets with --entry "+o.Key+"=<entry>"
			}
			if len(left) > 0 {
				plan.keep[o.Key] = left
			}
		case o.Merged():
			// entries that cannot be secrets, such as subnets, are imported together and shown by name
			item.Action, item.Value = ImportReady, strings.Join(entryNames(v), ", ")
			for name, record := range anyMap(v) {
				for field, cell := range anyMap(record) {
					plan.set[o.Key+"."+name+"."+field] = cell
				}
			}
		default:
			item.Action, item.Value = ImportReady, Text(v)
			plan.set[o.Key] = v
		}
		plan.Items = append(plan.Items, item)
	}

	for _, key := range f.Keys {
		if known[key] || underKnown(key, known) {
			continue
		}
		plan.Items = append(plan.Items, ImportItem{Key: key, Action: ImportUnknown, Note: "invalid setting"})
	}
	return plan, nil
}

// replaceable reports whether a deprecated option's value can be imported under the key that replaced it, which it cannot when the file sets that key too
func replaceable(o *Option, f File) bool {
	to, ok := registry[o.ReplacedBy]
	if !ok || !to.Writable() {
		return false
	}
	_, set := f.Lookup(to.Key)
	return !set
}

func underKnown(key string, known map[string]bool) bool {
	for k := range known {
		if o := registry[k]; o.Merged() && strings.HasPrefix(key, k+".") {
			return true
		}
	}
	return false
}

// Count is how many of the file's settings the plan imports
func (p *ImportPlan) Count() int {
	n := 0
	for _, item := range p.Items {
		if item.Action == ImportReady {
			n++
		}
	}
	return n
}

// Invalid lists the values the file sets that do not validate
func (p *ImportPlan) Invalid() []ImportItem {
	return slices.DeleteFunc(slices.Clone(p.Items), func(i ImportItem) bool { return i.Action != ImportInvalid })
}

// Remaining is a config file holding only what the database cannot: secrets, settings read before the database opens, client settings and map entries that were not imported
func (p *ImportPlan) Remaining() string {
	if len(p.keep) == 0 {
		return ""
	}
	return tomlDoc(All(), func(o *Option) (any, bool) {
		v, ok := p.keep[o.Key]
		return v, ok
	}, false)
}

// Apply stores the plan in the database. A config file that still sets these keeps winning until the lines are removed from it and the server restarts.
func (p *ImportPlan) Apply(b Backend, actor string) (map[string]string, error) {
	if len(p.set) == 0 {
		return nil, nil
	}
	refused, err := update(b, p.set, nil, actor, true)
	if err == nil && len(refused) == 0 {
		notify()
	}
	return refused, err
}

// Export writes what the database stores as a config file, or with effective every value in use. Secrets are always left out, as are the map entries the config file keeps.
func Export(effective bool) string {
	st := snapshot()
	server := slices.DeleteFunc(All(), func(o *Option) bool { return o.Scope != ScopeServer || o.Secret })
	return tomlDoc(server, func(o *Option) (any, bool) {
		if !effective && o.Merged() {
			entries := storedEntries(o, st.stored)
			return entries, len(entries) > 0
		}
		if !effective {
			v, ok := st.stored[o.Key]
			return v, ok
		}
		saved := st.saved[o.Key]
		if saved.src == SourceInherited || (saved.src == SourceDefault && o.Deprecated != "") {
			return nil, false
		}
		return visible(o, saved), true
	}, false)
}

// Sample is a config file describing the options that are only ever set in one, or that hold entries which are. The rest live in the database.
func Sample() string {
	shown := slices.DeleteFunc(All(), func(o *Option) bool {
		return o.Deprecated != "" || (o.Writable() && !o.FileEntriesSecret)
	})
	return "# The settings grendel reads from this file only: secrets, paths on this host and what is needed to open the database. The few a packaged install needs are set, and the rest are commented out at their default.\n" +
		"# Every other setting is stored in the database and changed with `grendel config set` or in the web UI. Any of them can still be set here, and a value in this file wins over the database: `grendel config list` shows them all.\n\n" +
		tomlDoc(shown, func(o *Option) (any, bool) {
			if o.Kind == KindTable && len(o.Default.([]map[string]any)) == 0 {
				return sampleRows(o), true
			}
			if o.Kind == KindGroup {
				return map[string]any{"example": sampleRows(o)[0]}, true
			}
			if o.Packaged != nil {
				return o.Packaged, true
			}
			return o.Default, true
		}, true)
}

// sampleRows is one made-up row for a table with no default, so the sample shows its shape
func sampleRows(o *Option) []map[string]any {
	row := map[string]any{}
	for _, f := range o.Fields {
		switch f.Kind {
		case KindInt:
			row[f.Name] = 0
		default:
			row[f.Name] = ""
		}
	}
	return []map[string]any{row}
}
