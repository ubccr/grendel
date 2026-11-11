// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/stmcginnis/gofish/schemas"
	"github.com/ubccr/grendel/pkg/model"
)

// AssetTagRequest sets each node's asset tag, either by node name or by the serial its bmc reports. Invalid holds the nodes whose value could not be worked out, which fail without being contacted
type AssetTagRequest struct {
	ByNode   map[string]string
	BySerial map[string]string
	Invalid  map[string]error
	DryRun   bool
}

// LocationUpdate is the location fields to write. A nil field is left as it is. RackUnit is only written when SetRackUnit is true, and a nil RackUnit then clears it
type LocationUpdate struct {
	Datacenter  *string
	Room        *string
	Row         *string
	Rack        *string
	RackUnit    *int
	SetRackUnit bool
}

type LocationRequest struct {
	ByNode  map[string]LocationUpdate
	Invalid map[string]error
	DryRun  bool
}

type AssetTagResult struct {
	model.BmcAssetTagResult
	Note string
}

func (r AssetTagResult) note() string { return r.Note }

type LocationResult struct {
	model.BmcLocationResult
	Note string
}

func (r LocationResult) note() string { return r.Note }

// SetAssetTag writes the asset tag of every target
func (s *Service) SetAssetTag(ctx context.Context, targets []Target, req AssetTagRequest) []NodeResult[AssetTagResult] {
	return run(ctx, s, targets, func(ctx context.Context, t Target) (*AssetTagResult, error) { return s.assetTagOne(ctx, t, req) })
}

// SetLocation writes the rack location of every target
func (s *Service) SetLocation(ctx context.Context, targets []Target, req LocationRequest) []NodeResult[LocationResult] {
	return run(ctx, s, targets, func(ctx context.Context, t Target) (*LocationResult, error) { return s.locationOne(ctx, t, req) })
}

func (s *Service) assetTagOne(ctx context.Context, t Target, req AssetTagRequest) (*AssetTagResult, error) {
	if err := req.Invalid[t.Name]; err != nil {
		return nil, err
	}

	c, err := s.connect(ctx, t)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	sys, err := firstSystem(c)
	if err != nil {
		return nil, err
	}

	out := &AssetTagResult{BmcAssetTagResult: model.BmcAssetTagResult{Serial: c.serialNumber(sys), Current: sys.AssetTag}}

	var ok bool
	if req.BySerial != nil {
		if out.Serial == "" {
			return nil, errors.New("bmc reports no serial number to match")
		}
		out.Desired, ok = req.BySerial[out.Serial]
		if !ok {
			return nil, fmt.Errorf("no row for serial %s", out.Serial)
		}
	} else if out.Desired, ok = req.ByNode[t.Name]; !ok {
		return nil, errors.New("no asset tag given for this node")
	}

	out.Changed = out.Desired != out.Current
	switch {
	case !out.Changed:
		out.Note = "already set"
		return out, nil
	case req.DryRun:
		out.Note = "would change, dry run"
		return out, nil
	}

	res, err := c.client.Patch(sys.ODataID, map[string]any{"AssetTag": out.Desired})
	schemas.DeferredCleanupHTTPResponse(res)
	if err != nil {
		return nil, err
	}
	out.Note = "asset tag set"
	return out, nil
}

func (s *Service) locationOne(ctx context.Context, t Target, req LocationRequest) (*LocationResult, error) {
	if err := req.Invalid[t.Name]; err != nil {
		return nil, err
	}
	update, ok := req.ByNode[t.Name]
	if !ok {
		return nil, errors.New("no location given for this node")
	}

	c, err := s.connect(ctx, t)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	sys, err := firstSystem(c)
	if err != nil {
		return nil, err
	}

	src, err := locationSourceOf(c, sys)
	if err != nil {
		return nil, err
	}

	out := &LocationResult{BmcLocationResult: model.BmcLocationResult{Current: src.current, Desired: applyLocation(src.current, update)}}
	out.Changed = !sameLocation(out.Current, out.Desired)
	switch {
	case !out.Changed:
		out.Note = "already set"
		return out, nil
	case req.DryRun:
		out.Note = "would change, dry run"
		return out, nil
	}

	if err := src.write(c, update); err != nil {
		return nil, err
	}
	out.Note = "location set"
	return out, nil
}

func firstSystem(c *conn) (*schemas.ComputerSystem, error) {
	systems, err := c.service.Systems()
	if err != nil {
		return nil, err
	}
	if len(systems) == 0 {
		return nil, errors.New("bmc exposes no computer systems")
	}
	return systems[0], nil
}

// the iDRAC keeps its location in these manager attributes, and reports them in the chassis Location as well
const (
	dellDatacenter = "ServerTopology.1.DataCenterName"
	dellRoom       = "ServerTopology.1.RoomName"
	dellAisle      = "ServerTopology.1.AisleName"
	dellRack       = "ServerTopology.1.RackName"
	dellRackSlot   = "ServerTopology.1.RackSlot"
)

// locationSource is where a bmc keeps the location: the iDRAC's ServerTopology attributes when it has them, the chassis Location otherwise
type locationSource struct {
	attributesURI string
	chassisURI    string
	current       model.BmcLocation
}

func locationSourceOf(c *conn, sys *schemas.ComputerSystem) (*locationSource, error) {
	if c.isDell() {
		if src := dellLocationSource(c); src != nil {
			return src, nil
		}
	}

	chassis, err := sys.Chassis()
	if err != nil {
		return nil, err
	}
	if len(chassis) == 0 {
		return nil, errors.New("bmc exposes no chassis to hold a location")
	}
	ch := chassis[0]
	return &locationSource{chassisURI: ch.ODataID, current: chassisLocation(ch.Location)}, nil
}

// dellLocationSource reads the ServerTopology attributes, or returns nil when the iDRAC does not serve them, which leaves the chassis to try
func dellLocationSource(c *conn) *locationSource {
	m, err := dellManager(c)
	if err != nil {
		return nil
	}

	uri := m.ODataID + "/Attributes"
	res, err := c.client.Get(uri)
	defer schemas.DeferredCleanupHTTPResponse(res)
	if err != nil {
		log.Debugf("node %s: no idrac attributes, using the chassis location: %s", c.target.Name, err)
		return nil
	}

	var raw struct {
		Attributes map[string]any
	}
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil
	}
	if _, ok := raw.Attributes[dellRack]; !ok {
		return nil
	}

	str := func(k string) string { v, _ := raw.Attributes[k].(string); return v }
	loc := model.BmcLocation{
		Datacenter: str(dellDatacenter),
		Room:       str(dellRoom),
		Row:        str(dellAisle),
		Rack:       str(dellRack),
	}
	if slot, ok := raw.Attributes[dellRackSlot].(float64); ok && slot > 0 {
		n := int(slot)
		loc.RackUnit = &n
	}
	return &locationSource{attributesURI: uri, current: loc}
}

// chassisLocation reads Placement and PostalAddress, falling back to the deprecated Info string on firmware that only fills that
func chassisLocation(l schemas.Location) model.BmcLocation {
	loc := model.BmcLocation{
		Datacenter: l.PostalAddress.Building,
		Room:       l.PostalAddress.Room,
		Row:        l.Placement.Row,
		Rack:       l.Placement.Rack,
		RackUnit:   l.Placement.RackOffset,
	}
	if loc.Room == "" {
		loc.Room = l.Placement.Room
	}
	if loc == (model.BmcLocation{}) && l.Info != "" {
		return infoLocation(l.Info, l.InfoFormat)
	}
	return loc
}

// infoLocation parses the semicolon separated Info string by the field names InfoFormat gives, such as DataCenter;RoomName;Aisle;RackName;RackSlot
func infoLocation(info, format string) model.BmcLocation {
	var loc model.BmcLocation
	values := strings.Split(info, ";")
	for i, name := range strings.Split(format, ";") {
		if i >= len(values) {
			break
		}
		v := strings.TrimSpace(values[i])
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "datacenter", "datacentername":
			loc.Datacenter = v
		case "roomname", "room":
			loc.Room = v
		case "aisle", "aislename", "row":
			loc.Row = v
		case "rackname", "rack":
			loc.Rack = v
		case "rackslot", "rackoffset":
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				loc.RackUnit = &n
			}
		}
	}
	return loc
}

func applyLocation(cur model.BmcLocation, u LocationUpdate) model.BmcLocation {
	out := cur
	set := func(dst *string, v *string) {
		if v != nil {
			*dst = *v
		}
	}
	set(&out.Datacenter, u.Datacenter)
	set(&out.Room, u.Room)
	set(&out.Row, u.Row)
	set(&out.Rack, u.Rack)
	if u.SetRackUnit {
		out.RackUnit = u.RackUnit
	}
	return out
}

func sameLocation(a, b model.BmcLocation) bool {
	if (a.RackUnit == nil) != (b.RackUnit == nil) || (a.RackUnit != nil && *a.RackUnit != *b.RackUnit) {
		return false
	}
	a.RackUnit, b.RackUnit = nil, nil
	return a == b
}

func (src *locationSource) write(c *conn, u LocationUpdate) error {
	uri, payload := src.payload(u)
	res, err := c.client.Patch(uri, payload)
	schemas.DeferredCleanupHTTPResponse(res)
	return err
}

// payload builds the PATCH carrying only the fields being changed
func (src *locationSource) payload(u LocationUpdate) (string, map[string]any) {
	if src.attributesURI != "" {
		attrs := map[string]any{}
		put := func(k string, v *string) {
			if v != nil {
				attrs[k] = *v
			}
		}
		put(dellDatacenter, u.Datacenter)
		put(dellRoom, u.Room)
		put(dellAisle, u.Row)
		put(dellRack, u.Rack)
		if u.SetRackUnit {
			// the iDRAC has no empty rack slot, 0 is what it reports when none is set
			slot := 0
			if u.RackUnit != nil {
				slot = *u.RackUnit
			}
			attrs[dellRackSlot] = slot
		}
		return src.attributesURI, map[string]any{"Attributes": attrs}
	}

	placement := map[string]any{}
	postal := map[string]any{}
	put := func(m map[string]any, k string, v *string) {
		if v != nil {
			m[k] = *v
		}
	}
	put(postal, "Building", u.Datacenter)
	put(postal, "Room", u.Room)
	put(placement, "Row", u.Row)
	put(placement, "Rack", u.Rack)
	if u.SetRackUnit {
		if u.RackUnit != nil {
			placement["RackOffset"] = *u.RackUnit
			placement["RackOffsetUnits"] = "EIA_310"
		} else {
			placement["RackOffset"] = nil
		}
	}

	location := map[string]any{}
	if len(placement) > 0 {
		location["Placement"] = placement
	}
	if len(postal) > 0 {
		location["PostalAddress"] = postal
	}
	return src.chassisURI, map[string]any{"Location": location}
}
