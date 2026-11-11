// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"path"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/stmcginnis/gofish/schemas"
	"github.com/ubccr/grendel/pkg/model"
)

type BiosRequest struct {
	// Names keeps only these attributes, all of them when empty
	Names []string
}

type SetBiosRequest struct {
	Attributes map[string]any
	// ApplyTime is a schemas.SettingsApplyTime, empty for the bmc's default
	ApplyTime string
}

// Bios reads the BIOS attributes of every target, with any staged changes
func (s *Service) Bios(ctx context.Context, targets []Target, req BiosRequest) []NodeResult[model.BmcBios] {
	return run(ctx, s, targets, func(ctx context.Context, t Target) (*model.BmcBios, error) { return s.biosOne(ctx, t, req) })
}

// SetBios stages BIOS attribute changes on every target
func (s *Service) SetBios(ctx context.Context, targets []Target, req SetBiosRequest) []NodeResult[Ack] {
	return run(ctx, s, targets, func(ctx context.Context, t Target) (*Ack, error) { return s.setBiosOne(ctx, t, req) })
}

// DiscardPendingBios drops the BIOS changes staged on every target
func (s *Service) DiscardPendingBios(ctx context.Context, targets []Target) []NodeResult[Ack] {
	return run(ctx, s, targets, s.discardBiosOne)
}

// BiosRegistry reads the attribute registry describing every target's BIOS attributes
func (s *Service) BiosRegistry(ctx context.Context, targets []Target) []NodeResult[model.BmcBiosRegistry] {
	return run(ctx, s, targets, s.biosRegistryOne)
}

// biosConn is a connection with the system's Bios resource and where its changes are staged
type biosConn struct {
	*conn
	sys      *schemas.ComputerSystem
	bios     *schemas.Bios
	settings schemas.Settings
}

func (s *Service) connectBios(ctx context.Context, t Target) (*biosConn, error) {
	c, err := s.connect(ctx, t)
	if err != nil {
		return nil, err
	}

	b, err := biosOf(c)
	if err != nil {
		c.Close()
		return nil, err
	}
	return b, nil
}

func biosOf(c *conn) (*biosConn, error) {
	systems, err := c.service.Systems()
	if err != nil {
		return nil, err
	}
	if len(systems) == 0 {
		return nil, errors.New("bmc exposes no computer systems")
	}

	sys := systems[0]
	bios, err := sys.Bios()
	if err != nil {
		return nil, err
	}
	if bios == nil {
		return nil, errors.New("bmc exposes no bios resource")
	}

	settings, err := redfishSettings(c, bios.ODataID)
	if err != nil {
		return nil, err
	}

	return &biosConn{conn: c, sys: sys, bios: bios, settings: settings}, nil
}

// redfishSettings reads a resource's @Redfish.Settings, which gofish keeps private on Bios. A resource with no settings object takes its changes directly, and settings.SettingsObject is then empty
func redfishSettings(c *conn, uri string) (schemas.Settings, error) {
	res, err := c.client.Get(uri)
	defer schemas.DeferredCleanupHTTPResponse(res)
	if err != nil {
		return schemas.Settings{}, err
	}

	var raw struct {
		Settings schemas.Settings `json:"@Redfish.Settings"`
	}
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return schemas.Settings{}, fmt.Errorf("reading the bios resource: %w", err)
	}
	return raw.Settings, nil
}

// pending reads the values staged in the settings object that differ from the applied ones. Some BMCs list only the staged attributes there and others list every one, so the comparison is what tells them apart
func (b *biosConn) pending() (map[string]any, error) {
	out := map[string]any{}
	if b.settings.SettingsObject == "" || b.settings.SettingsObject == b.bios.ODataID {
		return out, nil
	}

	res, err := b.client.Get(b.settings.SettingsObject)
	defer schemas.DeferredCleanupHTTPResponse(res)
	if err != nil {
		return nil, err
	}

	var staged struct {
		Attributes map[string]any
	}
	if err := json.NewDecoder(res.Body).Decode(&staged); err != nil {
		return nil, fmt.Errorf("reading the bios settings object: %w", err)
	}

	for name, value := range staged.Attributes {
		if current, ok := b.bios.Attributes[name]; !ok || !reflect.DeepEqual(current, value) {
			out[name] = value
		}
	}
	return out, nil
}

func (s *Service) biosOne(ctx context.Context, t Target, req BiosRequest) (*model.BmcBios, error) {
	b, err := s.connectBios(ctx, t)
	if err != nil {
		return nil, err
	}
	defer b.Close()

	pending, err := b.pending()
	if err != nil {
		return nil, err
	}

	attrs := map[string]any(b.bios.Attributes)
	if len(req.Names) > 0 {
		attrs = pick(attrs, req.Names)
		pending = pick(pending, req.Names)
	}
	if attrs == nil {
		attrs = map[string]any{}
	}

	return &model.BmcBios{
		Attributes:        attrs,
		Pending:           pending,
		AttributeRegistry: b.bios.AttributeRegistry,
		ApplyTimes:        applyTimes(b.bios),
		ResetPending:      b.bios.ResetBiosToDefaultsPending,
		Model:             b.sys.Model,
		BIOSVersion:       b.sys.BiosVersion,
	}, nil
}

func pick(m map[string]any, names []string) map[string]any {
	out := map[string]any{}
	for _, name := range names {
		if v, ok := m[name]; ok {
			out[name] = v
		}
	}
	return out
}

func applyTimes(bios *schemas.Bios) []string {
	times := bios.AllowedAttributeUpdateApplyTimes()
	out := make([]string, 0, len(times))
	for _, t := range times {
		out = append(out, string(t))
	}
	return out
}

func (s *Service) setBiosOne(ctx context.Context, t Target, req SetBiosRequest) (*Ack, error) {
	if len(req.Attributes) == 0 {
		return nil, errors.New("no bios attributes given")
	}

	b, err := s.connectBios(ctx, t)
	if err != nil {
		return nil, err
	}
	defer b.Close()

	attrs := make(schemas.SettingsAttributes, len(req.Attributes))
	for name, value := range req.Attributes {
		current, ok := b.bios.Attributes[name]
		if !ok {
			return nil, fmt.Errorf("unknown bios attribute %s, nothing was changed", name)
		}
		v, err := coerce(current, value)
		if err != nil {
			return nil, fmt.Errorf("bios attribute %s: %w, nothing was changed", name, err)
		}
		attrs[name] = v
	}

	applyTime := schemas.SettingsApplyTime(req.ApplyTime)
	if allowed := applyTimes(b.bios); applyTime != "" && !slices.Contains(allowed, req.ApplyTime) {
		return nil, fmt.Errorf("bmc does not accept apply time %s, it accepts %s", applyTime, strings.Join(allowed, ", "))
	}

	// compared here rather than left to gofish, which also skips unchanged values but returns the same nil task whether it sent a change or not
	pending, err := b.pending()
	if err != nil {
		return nil, err
	}
	changed := false
	for name, value := range attrs {
		current, ok := pending[name]
		if !ok {
			current = b.bios.Attributes[name]
		}
		if !reflect.DeepEqual(current, value) {
			changed = true
			break
		}
	}
	if !changed {
		return doneWith("already set, nothing sent")
	}

	task, err := b.bios.UpdateBiosAttributesApplyAtWithTask(attrs, applyTime)
	if err != nil {
		return nil, err
	}

	verb := "staged"
	if b.settings.SettingsObject == "" || b.settings.SettingsObject == b.bios.ODataID {
		verb = "set"
	}
	when := ""
	if applyTime != "" {
		when = ", apply time " + string(applyTime)
	}
	if task != nil && task.TaskMonitor != "" {
		return doneWith("%s %d bios attribute(s)%s, task %s", verb, len(attrs), when, path.Base(task.TaskMonitor))
	}
	return doneWith("%s %d bios attribute(s)%s", verb, len(attrs), when)
}

// coerce converts a requested value to the JSON type of the attribute's current value, so a value typed on the command line, which is always a string, reaches the bmc as the number or boolean it expects. An attribute with no current value, such as a password, takes the value as given
func coerce(current, requested any) (any, error) {
	switch current.(type) {
	case nil:
		return requested, nil
	case float64:
		switch v := requested.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case string:
			// a float64 like the current value, which is what decoding the bmc's JSON produced, so comparing them works
			n, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err != nil {
				return nil, fmt.Errorf("%q is not a number", v)
			}
			return n, nil
		}
		return nil, fmt.Errorf("expects a number, got %v", requested)
	case bool:
		switch v := requested.(type) {
		case bool:
			return v, nil
		case string:
			b, err := strconv.ParseBool(strings.TrimSpace(v))
			if err != nil {
				return nil, fmt.Errorf("%q is not true or false", v)
			}
			return b, nil
		}
		return nil, fmt.Errorf("expects true or false, got %v", requested)
	case string:
		if v, ok := requested.(string); ok {
			return v, nil
		}
		return nil, fmt.Errorf("expects a string, got %v", requested)
	}
	return requested, nil
}

// discardBiosOne empties the settings object, which drops every staged change. On an iDRAC a change staged OnReset has already become a BIOS configuration job, and removing that job is what discards it
func (s *Service) discardBiosOne(ctx context.Context, t Target) (*Ack, error) {
	b, err := s.connectBios(ctx, t)
	if err != nil {
		return nil, err
	}
	defer b.Close()

	if b.settings.SettingsObject == "" || b.settings.SettingsObject == b.bios.ODataID {
		return nil, errors.New("bmc applies bios changes directly, it has nothing staged to discard")
	}

	pending, err := b.pending()
	if err != nil {
		return nil, err
	}
	if len(pending) == 0 {
		return doneWith("no staged bios changes")
	}

	res, err := b.client.Delete(b.settings.SettingsObject)
	schemas.DeferredCleanupHTTPResponse(res)
	if err != nil {
		return nil, fmt.Errorf("bmc refused to discard the staged bios changes, clearing its bios configuration job may: %s", errorMessage(err))
	}

	return doneWith("discarded %d staged bios change(s)", len(pending))
}

func (s *Service) biosRegistryOne(ctx context.Context, t Target) (*model.BmcBiosRegistry, error) {
	b, err := s.connectBios(ctx, t)
	if err != nil {
		return nil, err
	}
	defer b.Close()

	name := b.bios.AttributeRegistry
	if name == "" {
		return &model.BmcBiosRegistry{Attributes: []model.BmcBiosAttribute{}}, nil
	}

	key := strings.Join([]string{b.service.Vendor, b.sys.Model, b.sys.BiosVersion, name}, "\x00")
	if reg, ok := s.registries.get(key); ok {
		return reg, nil
	}

	reg, err := fetchRegistry(b.conn, name)
	if err != nil {
		return nil, err
	}

	// only a registry that was found is kept, so one the bmc failed to serve is tried again next time
	if reg.Available {
		s.registries.put(key, reg)
	}
	return reg, nil
}

// fetchRegistry finds the registry the Bios resource names among the bmc's registry files, preferring the English copy. A registry the bmc only points to elsewhere, by a publication or archive uri, is reported unavailable rather than fetched
func fetchRegistry(c *conn, name string) (*model.BmcBiosRegistry, error) {
	unavailable := &model.BmcBiosRegistry{Registry: name, Attributes: []model.BmcBiosAttribute{}}

	files, err := c.service.Registries()
	if err != nil {
		return nil, err
	}

	file := findRegistryFile(files, name)
	if file == nil {
		return unavailable, nil
	}

	uri, language := "", ""
	for _, loc := range file.Location {
		if loc.URI == "" {
			continue
		}
		if uri == "" || strings.HasPrefix(strings.ToLower(loc.Language), "en") {
			uri, language = loc.URI, loc.Language
		}
		if strings.HasPrefix(strings.ToLower(loc.Language), "en") {
			break
		}
	}
	if uri == "" {
		return unavailable, nil
	}

	reg, err := schemas.GetAttributeRegistry(c.client, uri)
	if err != nil {
		return nil, err
	}

	out := narrowRegistry(reg)
	if out.Registry == "" {
		out.Registry = name
	}
	if out.Language == "" {
		out.Language = language
	}
	return out, nil
}

// findRegistryFile picks the registry file for the registry a Bios resource names. An exact id or registry match wins. Failing that the names are compared without their versions, because BMCs do not version them consistently: an iDRAC names BiosAttributeRegistry.v1_0_3 on the Bios resource, files it as BiosAttributeRegistry.v1_0_0, and calls it BiosAttributeRegistry1.0 inside the file
func findRegistryFile(files []*schemas.MessageRegistryFile, name string) *schemas.MessageRegistryFile {
	for _, f := range files {
		if f != nil && (f.ID == name || f.Registry == name) {
			return f
		}
	}

	base := registryBase(name)
	if base == "" {
		return nil
	}
	for _, f := range files {
		if f != nil && (registryBase(f.ID) == base || registryBase(f.Registry) == base) {
			return f
		}
	}
	return nil
}

// registryBase strips the version from a registry name, so BiosAttributeRegistry.v1_0_3, BiosAttributeRegistry.1.0 and BiosAttributeRegistry1.0 all become BiosAttributeRegistry
func registryBase(name string) string {
	name, _, _ = strings.Cut(name, ".")
	return strings.TrimRight(name, "0123456789")
}

func narrowRegistry(reg *schemas.AttributeRegistry) *model.BmcBiosRegistry {
	out := &model.BmcBiosRegistry{
		Available:  true,
		Registry:   reg.ID,
		Version:    reg.RegistryVersion,
		Language:   reg.Language,
		Attributes: make([]model.BmcBiosAttribute, 0, len(reg.RegistryEntries.Attributes)),
	}

	for _, a := range reg.RegistryEntries.Attributes {
		attr := model.BmcBiosAttribute{
			Name:          a.AttributeName,
			DisplayName:   a.DisplayName,
			HelpText:      a.HelpText,
			WarningText:   a.WarningText,
			MenuPath:      a.MenuPath,
			Type:          string(a.Type),
			Values:        make([]model.BmcBiosValue, 0, len(a.Value)),
			DefaultValue:  a.DefaultValue,
			ReadOnly:      a.ReadOnly,
			WriteOnly:     a.WriteOnly,
			Hidden:        a.Hidden,
			Immutable:     a.Immutable,
			GrayOut:       a.GrayOut,
			ResetRequired: a.ResetRequired,
			SystemUnique:  a.IsSystemUniqueProperty,
			LowerBound:    bound(a.LowerBound),
			UpperBound:    bound(a.UpperBound),
			MinLength:     a.MinLength,
			MaxLength:     a.MaxLength,
		}
		if a.DisplayOrder != nil {
			attr.DisplayOrder = *a.DisplayOrder
		}
		for _, v := range a.Value {
			attr.Values = append(attr.Values, model.BmcBiosValue{Name: v.ValueName, DisplayName: v.ValueDisplayName})
		}
		out.Attributes = append(out.Attributes, attr)
	}

	sort.SliceStable(out.Attributes, func(i, j int) bool {
		a, b := out.Attributes[i], out.Attributes[j]
		if a.MenuPath != b.MenuPath {
			return a.MenuPath < b.MenuPath
		}
		return a.DisplayOrder < b.DisplayOrder
	})

	return out
}

func bound(v *uint64) *int64 {
	if v == nil {
		return nil
	}
	n := int64(min(*v, math.MaxInt64))
	return &n
}

// registryCache keeps the attribute registries already fetched. A registry is the same for every node of one model and bios version, and is large enough that a rack of them should cost one fetch
type registryCache struct {
	mu    sync.Mutex
	max   int
	order []string
	items map[string]*model.BmcBiosRegistry
}

func newRegistryCache(max int) *registryCache {
	return &registryCache{max: max, items: map[string]*model.BmcBiosRegistry{}}
}

func (r *registryCache) get(key string) (*model.BmcBiosRegistry, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	reg, ok := r.items[key]
	if ok {
		r.touch(key)
	}
	return reg, ok
}

func (r *registryCache) put(key string, reg *model.BmcBiosRegistry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[key]; !ok && len(r.order) >= r.max {
		oldest := r.order[0]
		r.order = r.order[1:]
		delete(r.items, oldest)
	}
	r.items[key] = reg
	r.touch(key)
}

// touch moves key to the most recently used end
func (r *registryCache) touch(key string) {
	if i := slices.Index(r.order, key); i >= 0 {
		r.order = slices.Delete(r.order, i, i+1)
	}
	r.order = append(r.order, key)
}
