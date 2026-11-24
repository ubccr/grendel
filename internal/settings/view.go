// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

import (
	"maps"
	"slices"
)

type View struct {
	Key     string   `json:"key"`
	Summary string   `json:"summary"`
	Help    string   `json:"help"`
	Group   string   `json:"group"`
	Kind    Kind     `json:"kind"`
	Unit    string   `json:"unit"`
	Apply   Apply    `json:"apply"`
	Flag    string   `json:"flag"`
	Env     string   `json:"env"`
	Min     *int     `json:"min,omitempty"`
	Max     *int     `json:"max,omitempty"`
	Enum    []string `json:"enum"`
	Fields  []Field  `json:"fields"`

	// Values have the option's own JSON type: text, a number, true or false, a list or an object. A duration is text such as 30s.
	Default any `json:"default"`
	// Value is what a restart would run with
	Value   any `json:"value"`
	Running any `json:"running"`
	// Stored is what the database holds, and is left out when it holds nothing for the option
	Stored any `json:"stored,omitempty"`
	// Source is where Value comes from. A flag, the environment or the config file wins over the database, so the option cannot be changed while one of them sets it.
	Source Source `json:"source"`
	// FileKeys names the entries of a map that the config file sets, which win over the database's
	FileKeys []string `json:"file_keys"`
	// FileEntriesSecret means the values of those entries are left out of Value
	FileEntriesSecret bool `json:"file_entries_secret"`

	Shadowed bool `json:"shadowed"`
	Pending  bool `json:"pending"`
	Secret   bool `json:"secret"`
	// External means only the config file, the environment or a flag may set the option
	External bool `json:"external"`
	// InheritFrom names the setting whose value this one takes while nothing sets it
	InheritFrom string `json:"inherit_from"`
	// EntryOf and Entry are set on a field of one named entry, such as dhcp.subnets.rack-f.mtu: the setting the entry belongs to and the entry's name
	EntryOf string `json:"entry_of"`
	Entry   string `json:"entry"`
	// EntryLabel is what an entry of this setting is called, such as Subnet
	EntryLabel string `json:"entry_label"`
	Deprecated string `json:"deprecated"`
}

type SectionView struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Options     []View `json:"options"`
}

type Listing struct {
	Persistent bool          `json:"persistent"`
	ConfigFile string        `json:"config_file"`
	Pending    int           `json:"pending"`
	Sections   []SectionView `json:"sections"`
}

// List describes every server option and its state. A secret never carries a value, only whether it is set.
func List() Listing {
	st := snapshot()

	mu.Lock()
	out := Listing{Persistent: env.Persistent, ConfigFile: env.ConfigFile}
	mu.Unlock()

	for _, sec := range sections {
		sv := SectionView{ID: sec.ID, Title: sec.Title, Description: sec.Description, Options: []View{}}
		for _, o := range ordered {
			if o.Section != sec.ID || o.Scope != ScopeServer {
				continue
			}
			v := view(o, st)
			// an option kept only for old config files is not offered unless one still sets it
			if o.Deprecated != "" && (v.Source == SourceDefault || v.Source == SourceInherited) {
				continue
			}
			if v.Pending {
				out.Pending++
			}
			sv.Options = append(sv.Options, v)
			if o.Kind == KindGroup {
				sv.Options = append(sv.Options, entryViews(o, st)...)
			}
		}
		out.Sections = append(out.Sections, sv)
	}
	return out
}

func view(o *Option, st *state) View {
	saved, boot := st.saved[o.Key], st.boot[o.Key]
	v := View{
		Key:               o.Key,
		Summary:           o.Summary,
		Help:              o.Help,
		Group:             o.Group,
		Kind:              o.Kind,
		Unit:              o.Unit,
		Apply:             o.Apply,
		Flag:              o.Flag,
		Env:               o.Env(),
		Min:               o.Min,
		Max:               o.Max,
		Enum:              o.Enum,
		Fields:            o.Fields,
		Source:            saved.src,
		FileKeys:          saved.fileKeys,
		FileEntriesSecret: o.FileEntriesSecret,
		Shadowed:          saved.shadowed,
		Secret:            o.Secret,
		InheritFrom:       o.InheritFrom,
		External:          o.External,
		Deprecated:        o.Deprecated,
	}
	if v.Enum == nil {
		v.Enum = []string{}
	}
	if v.Fields == nil {
		v.Fields = []Field{}
	}
	if v.FileKeys == nil {
		v.FileKeys = []string{}
	}
	if o.Merged() {
		v.Env = ""
	}
	if o.Secret {
		return v
	}
	if o.Kind == KindGroup {
		// the entries are listed field by field after this, so here they are only named
		v.EntryLabel = o.EntryLabel
		v.Value = entryNames(saved.v)
		v.Running = v.Value
		return v
	}

	v.Default = storedForm(o.Default)
	if o.InheritFrom != "" {
		v.Default = storedForm(st.saved[o.InheritFrom].v)
	}
	v.Value = storedForm(visible(o, saved))
	v.Running = v.Value
	if o.Apply == ApplyRestart {
		v.Running = storedForm(visible(o, boot))
	}
	v.Pending = Text(v.Running) != Text(v.Value)
	if dbValue, ok := st.stored[o.Key]; ok {
		v.Stored = storedForm(dbValue)
	}
	return v
}

// visible is a value with the entries that may be secret taken out
func visible(o *Option, val value) any {
	m, ok := val.v.(map[string]any)
	if !ok || !o.FileEntriesSecret || len(val.fileKeys) == 0 {
		return val.v
	}
	shown := maps.Clone(m)
	for _, key := range val.fileKeys {
		delete(shown, key)
	}
	return shown
}

func entryNames(group any) []string {
	names := entryKeys(group)
	slices.SortFunc(names, NaturalCompare)
	return names
}

// entryViews lists each field of each entry of a group as a setting of its own, such as dhcp.subnets.rack-f.mtu, in the order the entries are used
func entryViews(o *Option, st *state) []View {
	saved := st.saved[o.Key]
	group, _ := saved.v.(map[string]any)
	out := []View{}
	for _, name := range entryNames(group) {
		record, _ := group[name].(map[string]any)
		fromFile := slices.Contains(saved.fileKeys, name)
		for _, f := range o.Fields {
			v := View{
				Key:      o.Key + "." + name + "." + f.Name,
				Summary:  f.Summary,
				Help:     f.Help,
				Group:    o.EntryLabel + " " + name,
				Kind:     f.Kind,
				Apply:    o.Apply,
				Enum:     []string{},
				Fields:   []Field{},
				FileKeys: []string{},
				Source:   SourceDefault,
				EntryOf:  o.Key,
				Entry:    name,
			}
			if raw, ok := record[f.Name]; ok {
				v.Value, v.Source = storedForm(raw), SourceDatabase
				v.Stored = v.Value
			}
			// every field of an entry the config file defines is the file's, set or not
			if fromFile {
				v.Source, v.Stored = SourceFile, nil
			}
			v.Running = v.Value
			out = append(out, v)
		}
	}
	return out
}
