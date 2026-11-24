// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

import (
	"fmt"
	"slices"
	"strings"
)

type Kind string

const (
	KindString     Kind = "string"
	KindBool       Kind = "bool"
	KindInt        Kind = "int"
	KindFloat      Kind = "float"
	KindDuration   Kind = "duration"
	KindPath       Kind = "path"
	KindEnum       Kind = "enum"
	KindStringList Kind = "string_list"
	KindIntList    Kind = "int_list"
	// KindMap holds map[string]any with any keys. The entries in the config file and the ones in the database are merged, and the file wins per entry.
	KindMap Kind = "map"
	// KindGroup holds map[string]any of named records, each a map[string]any with the columns in Fields. A name in the config file wins whole over the database, which keeps one row per field under a key such as dhcp.subnets.rack-f.mtu.
	KindGroup Kind = "group"
	// KindTable holds []map[string]any, a list of records with the columns in Fields. It can only be set in the config file.
	KindTable Kind = "table"
)

type Apply string

const (
	ApplyLive    Apply = "live"
	ApplyRestart Apply = "restart"
)

type Scope string

const (
	ScopeServer Scope = "server"
	// ScopeClient options configure the grendel command line client and mean nothing to the server
	ScopeClient Scope = "client"
)

type Section struct {
	ID          string
	Title       string
	Description string
}

var sections = []Section{
	{ID: "general", Title: "General", Description: "The database, logging and which services this process runs."},
	{ID: "api", Title: "API", Description: "The HTTP API the command line client and this web UI talk to."},
	{ID: "provision", Title: "Provisioning", Description: "The HTTP server hosts boot and install from, and what their templates are given."},
	{ID: "dhcp", Title: "DHCP", Description: "Leases and the network options handed to hosts."},
	{ID: "dns", Title: "DNS", Description: "The built-in DNS server that answers for the hosts grendel knows."},
	{ID: "boot", Title: "TFTP and PXE", Description: "The servers that hand a booting host its first loader."},
	{ID: "events", Title: "Events", Description: "The history of provisioning and BMC actions."},
	{ID: "bmc", Title: "BMC", Description: "How grendel signs in to and talks with baseboard management controllers."},
	{ID: "switch", Title: "Switches", Description: "How grendel reads network switches. Arista and SONiC are read over gNMI, Dell OS10 over SNMP."},
}

// Field is one column of a KindTable or KindGroup option
type Field struct {
	Name     string            `json:"name"`
	Summary  string            `json:"summary"`
	Help     string            `json:"help"`
	Kind     Kind              `json:"kind"`
	Required bool              `json:"required"`
	Validate func(v any) error `json:"-"`
}

func (f Field) coerce(raw any) (any, error) {
	v, err := coerce(f.Kind, raw, &Option{})
	if err != nil {
		return nil, err
	}
	if f.Validate != nil {
		if err := f.Validate(v); err != nil {
			return nil, err
		}
	}
	return v, nil
}

type Option struct {
	Key     string
	Summary string
	Help    string
	Section string
	// Group names a subheading the option is listed under within its section
	Group string
	Kind  Kind
	// Default holds the canonical Go value for Kind
	Default  any
	Unit     string
	Min      *int
	Max      *int
	Required bool
	Enum     []string
	Fields   []Field
	Apply    Apply
	Scope    Scope
	// Flag is the flag that overrides the option, without the leading dashes
	Flag   string
	Secret bool
	// Generated marks a secret that is made up and kept in the database when nothing sets it
	Generated bool
	// GeneratedBytes is how many random bytes a generated secret is made from, written as twice that many hex characters
	GeneratedBytes int
	// External marks an option that only the config file, the environment or a flag may set. Unlike a secret its value is shown.
	External bool
	// EntryLabel is what one entry of a KindGroup option is called, such as Subnet
	EntryLabel string
	// InheritFrom names the option whose value this one takes while nothing sets it, as switch.gnmi.sonic.port follows switch.gnmi.port
	InheritFrom string
	// Deprecated says, for an option only still read for old config files, what to do about it. It completes the sentence "deprecated, ...".
	Deprecated string
	// ReplacedBy is the key of the option a deprecated one's value belongs under, which is where an import puts it
	ReplacedBy string
	// FileEntriesSecret hides the values of the entries a merged option takes from the config file, since that is where its secrets are kept
	FileEntriesSecret bool
	// Validate checks a value Coerce already turned into its canonical type
	Validate func(v any) error
}

// Writable reports whether the option may be stored in the database
func (o *Option) Writable() bool {
	return !o.Secret && !o.External && o.Scope == ScopeServer
}

// Merged reports whether the config file and the database each contribute entries to the option, rather than one replacing the other
func (o *Option) Merged() bool {
	return o.Kind == KindMap || o.Kind == KindGroup
}

func (o *Option) field(name string) (Field, bool) {
	i := slices.IndexFunc(o.Fields, func(f Field) bool { return f.Name == name })
	if i < 0 {
		return Field{}, false
	}
	return o.Fields[i], true
}

func (o *Option) Env() string {
	return "GRENDEL_" + strings.ToUpper(strings.ReplaceAll(o.Key, ".", "_"))
}

var (
	registry = map[string]*Option{}
	ordered  []*Option
)

// Opt is a registered option read as a T
type Opt[T any] struct{ *Option }

func register[T any](o Option) Opt[T] {
	if _, ok := o.Default.(T); !ok {
		panic(fmt.Sprintf("settings: default of %s is %T", o.Key, o.Default))
	}
	if _, dup := registry[o.Key]; dup {
		panic("settings: " + o.Key + " registered twice")
	}
	if o.Apply == "" {
		o.Apply = ApplyRestart
	}
	if o.Scope == "" {
		o.Scope = ScopeServer
	}
	if o.Scope == ScopeServer && !slices.ContainsFunc(sections, func(s Section) bool { return s.ID == o.Section }) {
		panic("settings: " + o.Key + " is in an unknown section")
	}
	registry[o.Key] = &o
	ordered = append(ordered, &o)
	return Opt[T]{&o}
}

// Dumped reports whether a stored row belongs in a dump: a setting someone chose, rather than a secret generated for this install
func Dumped(key string) bool {
	o, ok := registry[key]
	return !ok || !o.Generated
}

func Lookup(key string) (*Option, bool) {
	o, ok := registry[key]
	return o, ok
}

// All returns every registered option in the order it is listed
func All() []*Option {
	return slices.Clone(ordered)
}

// Get returns the value in effect: the saved one for a live option, and the one read at startup for an option that needs a restart
func (o Opt[T]) Get() T {
	v, _ := running(o.Option).(T)
	return v
}

func (o Opt[T]) DefaultValue() T {
	return o.Default.(T)
}

func ptr(n int) *int { return &n }
