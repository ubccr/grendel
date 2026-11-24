// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Source string

const (
	SourceDefault  Source = "default"
	SourceDatabase Source = "database"
	SourceFile     Source = "file"
	SourceEnv      Source = "env"
	SourceFlag     Source = "flag"
	// SourceGenerated is a secret made up for this process because nothing sets it and there is no database to keep it in
	SourceGenerated Source = "generated"
	// SourceInherited is the value of the option named by InheritFrom, taken because nothing sets this one
	SourceInherited Source = "inherited"
	sourceOverride  Source = "override"
)

// External looks an option up in the layers above the database: a flag given on the command line, the environment, then the config file
type External func(o *Option) (raw any, src Source, ok bool)

type Row struct {
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	UpdatedAt time.Time       `json:"updated_at"`
	UpdatedBy string          `json:"updated_by"`
}

type Backend interface {
	// Settings returns the stored rows and a revision that changes whenever they do
	Settings() (rows []Row, revision int64, err error)
	SettingsRevision() (int64, error)
	// InsertSettings stores the rows whose keys are not there yet and leaves the rest as they are
	InsertSettings(set []Row, actor string) (revision int64, err error)
	// UpdateSettings stores set and removes unset in one transaction, and returns the new revision
	UpdateSettings(set []Row, unset []string, actor string) (revision int64, err error)
}

type Env struct {
	Lookup     External
	ConfigFile string
	// Persistent is false when the database lives in memory and saved settings are lost at exit
	Persistent bool
}

// Problem is a value that was skipped because it does not validate
type Problem struct {
	Key     string
	Source  Source
	Message string
}

func (p Problem) String() string {
	return fmt.Sprintf("%s from the %s is ignored: %s", p.Key, sourceName(p.Source), p.Message)
}

type value struct {
	v        any
	src      Source
	shadowed bool
	// fileKeys are the entries of a merged option that the config file sets
	fileKeys []string
}

type state struct {
	boot     map[string]value
	saved    map[string]value
	stored   map[string]any
	revision int64
}

var (
	// mu serialises everything that builds a state. Readers only load current.
	mu      sync.Mutex
	current atomic.Pointer[state]
	env     Env
	// ephemeral holds the generated secrets of a process with no database, so they stay the same for its lifetime
	ephemeral = map[string]any{}
	listeners []func()
)

func Configure(e Env) {
	mu.Lock()
	defer mu.Unlock()
	env = e
}

// OnChange registers fn to run after the settings change while the server is running
func OnChange(fn func()) {
	mu.Lock()
	defer mu.Unlock()
	listeners = append(listeners, fn)
}

func notify() {
	mu.Lock()
	fns := slices.Clone(listeners)
	mu.Unlock()
	for _, fn := range fns {
		fn()
	}
}

func snapshot() *state {
	if st := current.Load(); st != nil {
		return st
	}
	mu.Lock()
	defer mu.Unlock()
	if st := current.Load(); st != nil {
		return st
	}
	swap(map[string]any{}, 0, true, func(Problem) {})
	return current.Load()
}

func running(o *Option) any {
	st := snapshot()
	if o.Apply == ApplyRestart {
		return st.boot[o.Key].v
	}
	return st.saved[o.Key].v
}

// Value returns the value in effect for a key, for callers that only know the key at run time
func Value(key string) (any, bool) {
	o, ok := registry[key]
	if !ok {
		return nil, false
	}
	return running(o), true
}

func external(o *Option) (any, Source, bool) {
	if env.Lookup == nil {
		return nil, "", false
	}
	return env.Lookup(o)
}

func resolve(o *Option, stored map[string]any, problem func(Problem)) value {
	if o.Merged() {
		return resolveMerged(o, stored, problem)
	}

	dbValue, inDB := stored[o.Key]
	if raw, src, ok := external(o); ok {
		v, err := o.Coerce(raw)
		if err == nil {
			return value{v: v, src: src, shadowed: inDB && !o.Secret}
		}
		problem(Problem{Key: o.Key, Source: src, Message: err.Error()})
	}
	if inDB {
		return value{v: dbValue, src: SourceDatabase}
	}
	if o.Generated {
		// with no database to keep it in, the secret is made up once and lasts as long as the process
		if _, ok := ephemeral[o.Key]; !ok {
			ephemeral[o.Key] = randomHex(o.GeneratedBytes)
		}
		return value{v: ephemeral[o.Key], src: SourceGenerated}
	}
	return value{v: o.Default, src: SourceDefault}
}

// resolveMerged lays the default entries, then the database's, then the config file's over each other
func resolveMerged(o *Option, stored map[string]any, problem func(Problem)) value {
	out := value{src: SourceDefault}
	layers := []map[string]any{o.Default.(map[string]any)}
	db := storedEntries(o, stored)
	if len(db) > 0 {
		layers = append(layers, db)
		out.src = SourceDatabase
	}
	raw, src, ok := external(o)
	switch {
	case ok && src != SourceFile:
		problem(Problem{Key: o.Key, Source: src, Message: "its entries can only be set in the config file or the database"})
	case ok:
		v, err := o.Coerce(raw)
		if err != nil {
			problem(Problem{Key: o.Key, Source: src, Message: err.Error()})
			break
		}
		layers = append(layers, v.(map[string]any))
		out.fileKeys = entryKeys(v)
		if len(db) == 0 && len(out.fileKeys) > 0 {
			out.src = SourceFile
		}
	}

	merged := map[string]any{}
	for _, layer := range layers {
		maps.Copy(merged, layer)
	}
	out.v = merged
	return out
}

// storedEntries is the database's share of a merged option. A map has one row, and a group one row per field of each entry, which are put back together here.
func storedEntries(o *Option, stored map[string]any) map[string]any {
	if o.Kind != KindGroup {
		entries, _ := stored[o.Key].(map[string]any)
		return entries
	}
	out := map[string]any{}
	prefix := o.Key + "."
	for key, v := range stored {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		name, field, _ := strings.Cut(key[len(prefix):], ".")
		record, _ := out[name].(map[string]any)
		if record == nil {
			record = map[string]any{}
			out[name] = record
		}
		record[field] = v
	}
	return out
}

func entryKeys(v any) []string {
	keys := []string{}
	if m, ok := v.(map[string]any); ok {
		keys = slices.AppendSeq(keys, maps.Keys(m))
	}
	slices.Sort(keys)
	return keys
}

// swap publishes a state built from stored. fresh also makes it the state the options needing a restart run with.
func swap(stored map[string]any, revision int64, fresh bool, problem func(Problem)) {
	next := &state{stored: stored, revision: revision, saved: make(map[string]value, len(ordered))}
	for _, o := range ordered {
		next.saved[o.Key] = resolve(o, stored, problem)
	}
	inherit(next.saved)
	if prev := current.Load(); prev != nil && !fresh {
		next.boot = prev.boot
	} else {
		next.boot = next.saved
	}
	current.Store(next)
}

// inherit gives each option that nothing sets the value of the one it follows
func inherit(saved map[string]value) {
	for _, o := range ordered {
		if src := saved[o.Key].src; o.InheritFrom != "" && (src == SourceDefault || src == SourceInherited) {
			saved[o.Key] = value{v: saved[o.InheritFrom].v, src: SourceInherited}
		}
	}
}

// Load resolves every option as the process starts, from the flags, the environment and the config file alone when b is nil. A value that does not validate is skipped and returned, so a bad database row never stops a start.
func Load(b Backend) ([]Problem, error) {
	return load(b, true)
}

// Reload reads the database again while running. Options that need a restart keep the values the process started with.
func Reload(b Backend) ([]Problem, error) {
	problems, err := load(b, false)
	if err == nil {
		notify()
	}
	return problems, err
}

func load(b Backend, fresh bool) (problems []Problem, err error) {
	mu.Lock()
	defer mu.Unlock()

	problem := func(p Problem) { problems = append(problems, p) }
	stored, revision, err := readStored(b, problem)
	if err != nil {
		return nil, err
	}
	if secrets := newSecrets(stored); b != nil && len(secrets) > 0 {
		// inserted only where absent and then read back, so processes starting together on a new database all end up with the secret one of them made
		if _, err := b.InsertSettings(secrets, "system"); err != nil {
			return nil, err
		}
		problems = nil
		if stored, revision, err = readStored(b, problem); err != nil {
			return nil, err
		}
	}

	swap(stored, revision, fresh, problem)
	return problems, nil
}

func readStored(b Backend, problem func(Problem)) (map[string]any, int64, error) {
	stored := map[string]any{}
	if b == nil {
		return stored, 0, nil
	}
	rows, revision, err := b.Settings()
	if err != nil {
		return nil, 0, err
	}
	for _, row := range rows {
		var raw any
		err := json.Unmarshal(row.Value, &raw)
		if err == nil {
			raw, err = coerceStored(row.Key, raw)
		}
		if err != nil {
			problem(Problem{Key: row.Key, Source: SourceDatabase, Message: err.Error()})
			continue
		}
		stored[row.Key] = raw
	}
	return stored, revision, nil
}

// coerceStored turns a value into what the database keeps under key, which is a writable option, a generated secret or one field of a group's entry
func coerceStored(key string, raw any) (any, error) {
	if o, ok := registry[key]; ok {
		switch {
		case o.Kind == KindGroup:
			return nil, fmt.Errorf("set its entries one field at a time, as %s.<name>.<field>", key)
		case !o.Writable() && !o.Generated:
			return nil, errors.New("managed outside the database")
		}
		return o.Coerce(raw)
	}
	g, name, field, ok := groupPath(key)
	if !ok {
		return nil, errors.New("unknown setting")
	}
	if !groupName.MatchString(name) {
		return nil, fmt.Errorf("%q is not a name: use lowercase letters, digits, - and _", name)
	}
	f, ok := g.field(field)
	if !ok {
		return nil, fmt.Errorf("name one of its fields, as %s.%s.<field>", g.Key, name)
	}
	return f.coerce(raw)
}

// newSecrets makes up each generated secret that nothing sets
func newSecrets(stored map[string]any) []Row {
	var rows []Row
	for _, o := range ordered {
		if !o.Generated {
			continue
		}
		if _, ok := stored[o.Key]; ok {
			continue
		}
		if _, _, ok := external(o); ok {
			continue
		}
		encoded, _ := json.Marshal(randomHex(o.GeneratedBytes))
		rows = append(rows, Row{Key: o.Key, Value: encoded})
	}
	return rows
}

func randomHex(n int) string {
	buf := make([]byte, n)
	// crypto/rand.Read never fails on the platforms grendel runs on, and panics itself if it would
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

// Update stores every change or none, taking each value as text or already in the option's type. A non-empty first result maps a key to why it was refused.
func Update(b Backend, set map[string]any, unset []string, actor string) (map[string]string, error) {
	refused, err := update(b, set, unset, actor, false)
	if err == nil && len(refused) == 0 {
		notify()
	}
	return refused, err
}

// Restore stores the settings of a dump on top of what is there. The ones this version does not know or would not store are left out, and returned with the reason.
func Restore(b Backend, rows []Row, actor string) (skipped map[string]string, err error) {
	set := map[string]any{}
	skipped = map[string]string{}
	for _, row := range rows {
		var raw any
		if err := json.Unmarshal(row.Value, &raw); err != nil {
			skipped[row.Key] = err.Error()
			continue
		}
		set[row.Key] = raw
	}
	for len(set) > 0 {
		refused, err := update(b, set, nil, actor, true)
		if err != nil {
			return nil, err
		}
		if len(refused) == 0 {
			notify()
			break
		}
		for key, reason := range refused {
			skipped[key] = reason
			if _, ok := set[key]; ok {
				delete(set, key)
				continue
			}
			// an entry refused for a field the dump lacks goes whole
			entry := key[:strings.LastIndex(key, ".")+1]
			maps.DeleteFunc(set, func(k string, _ any) bool { return strings.HasPrefix(k, entry) })
		}
	}
	return skipped, nil
}

func update(b Backend, set map[string]any, unset []string, actor string, overridden bool) (map[string]string, error) {
	mu.Lock()
	defer mu.Unlock()

	refused := map[string]string{}
	prev := current.Load()
	stored := map[string]any{}
	var revision int64
	if prev != nil {
		stored = maps.Clone(prev.stored)
		revision = prev.revision
	}
	changed := map[string]bool{}
	removed := map[string]bool{}
	drop := func(key string) {
		delete(stored, key)
		delete(changed, key)
		removed[key] = true
	}

	for key, raw := range set {
		o, isOption := registry[key]
		g, name, field, isField := groupPath(key)
		switch {
		case isOption && !o.Writable():
			refused[key] = "managed outside the database"
			continue
		case isOption && !overridden && !o.Merged():
			if _, src, ok := external(o); ok {
				refused[key] = fmt.Sprintf("overridden by the %s, remove it there first", sourceName(src))
				continue
			}
		case isField && !overridden && prev != nil && slices.Contains(prev.saved[g.Key].fileKeys, name):
			refused[key] = fmt.Sprintf("%s is set in the config file, remove it there first", name)
			continue
		}
		// a field set to nothing is cleared, since an empty MTU is no MTU rather than a bad number
		if text, ok := raw.(string); ok && isField && field != "" && strings.TrimSpace(text) == "" {
			drop(key)
			continue
		}
		v, err := coerceStored(key, raw)
		switch {
		case err != nil:
			refused[key] = err.Error()
		case isField && empty(v):
			drop(key)
		default:
			stored[key] = v
			changed[key] = true
		}
	}
	for _, key := range unset {
		o, isOption := registry[key]
		_, _, field, isField := groupPath(key)
		switch {
		case isOption && !o.Writable():
			refused[key] = "managed outside the database"
		case (isOption && o.Kind == KindGroup) || (isField && field == ""):
			// a group or one of its entries goes with every field stored under it
			under := slices.DeleteFunc(slices.Collect(maps.Keys(stored)), func(k string) bool { return !strings.HasPrefix(k, key+".") })
			if isField && len(under) == 0 {
				refused[key] = "the database has no such entry"
			}
			for _, k := range under {
				drop(k)
			}
		case isOption || isField:
			drop(key)
		default:
			refused[key] = "unknown setting"
		}
	}
	for _, g := range ordered {
		if g.Kind != KindGroup {
			continue
		}
		for name, record := range storedEntries(g, stored) {
			if field, ok := missing(g, record.(map[string]any)); ok {
				refused[g.Key+"."+name+"."+field] = "must be set"
			}
		}
	}
	if len(refused) > 0 {
		return refused, nil
	}

	var rows []Row
	for key := range changed {
		encoded, err := json.Marshal(storedForm(stored[key]))
		if err != nil {
			return nil, err
		}
		rows = append(rows, Row{Key: key, Value: encoded})
	}
	if b != nil {
		var err error
		if revision, err = b.UpdateSettings(rows, slices.Collect(maps.Keys(removed)), actor); err != nil {
			return nil, err
		}
	}
	swap(stored, revision, false, func(Problem) {})
	return nil, nil
}

// groupPath splits a key such as dhcp.subnets.rack-f.mtu into its group option, the entry's name and, when there is one, the field
func groupPath(key string) (o *Option, name, field string, ok bool) {
	for _, g := range ordered {
		if g.Kind != KindGroup || !strings.HasPrefix(key, g.Key+".") {
			continue
		}
		name, field, _ = strings.Cut(strings.TrimPrefix(key, g.Key+"."), ".")
		if name == "" || strings.Contains(field, ".") {
			return nil, "", "", false
		}
		return g, name, field, true
	}
	return nil, "", "", false
}

// Watch reloads the settings when another process changes them in the database
func Watch(ctx context.Context, b Backend, every time.Duration, report func(error)) {
	tick := time.NewTicker(every)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}

		revision, err := b.SettingsRevision()
		if err != nil {
			report(err)
			continue
		}
		if st := current.Load(); st != nil && st.revision == revision {
			continue
		}
		if _, err := Reload(b); err != nil {
			report(err)
		}
	}
}

func sourceName(src Source) string {
	switch src {
	case SourceFile:
		return "config file"
	case SourceEnv:
		return "environment"
	case SourceFlag:
		return "command line"
	}
	return string(src)
}

// SetOverride pins an option until the returned func is called, or until the settings are next loaded or updated. It exists for tests.
func SetOverride(key string, raw any) (restore func(), err error) {
	mu.Lock()
	defer mu.Unlock()

	o, ok := registry[key]
	if !ok {
		return nil, fmt.Errorf("unknown setting %s", key)
	}
	v, err := o.Coerce(raw)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}

	if current.Load() == nil {
		swap(map[string]any{}, 0, true, func(Problem) {})
	}
	prev := current.Load()
	next := &state{stored: prev.stored, revision: prev.revision, saved: maps.Clone(prev.saved)}
	next.saved[key] = value{v: v, src: sourceOverride}
	inherit(next.saved)
	// boot too, so a test override takes effect at once even for an option that otherwise needs a restart
	next.boot = next.saved
	current.Store(next)
	return func() { current.Store(prev) }, nil
}

// Reset forgets everything loaded, so the next Load behaves like a fresh start. It exists for tests.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	current.Store(nil)
	env = Env{}
	listeners = nil
	clear(ephemeral)
}

// Derived is a value computed from settings and kept until they change
type Derived[T any] struct {
	fn    func() T
	cache atomic.Pointer[derived[T]]
}

type derived[T any] struct {
	st *state
	v  T
}

func Derive[T any](fn func() T) *Derived[T] {
	return &Derived[T]{fn: fn}
}

func (d *Derived[T]) Get() T {
	st := snapshot()
	if c := d.cache.Load(); c != nil && c.st == st {
		return c.v
	}
	v := d.fn()
	d.cache.Store(&derived[T]{st: st, v: v})
	return v
}
