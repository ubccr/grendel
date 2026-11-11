// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/facette/natsort"
	"github.com/go-fuego/fuego"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/pkg/model"
)

// BmcAssetTagBody sets asset tags from a template, or from a table of serial numbers. With neither, the bmc.asset_tag template from the config is used
type BmcAssetTagBody struct {
	Template *string           `json:"template" oai3:"nullable" description:"Go template rendered per node, such as UB-{{.Name}}. A plain string is a valid template"`
	BySerial map[string]string `json:"by_serial" oai3:"nullable" description:"Asset tag for each serial number. A selected node whose serial is not listed fails"`
	DryRun   bool              `json:"dry_run" description:"Read the bmcs and report what would change, without writing"`
}

// BmcLocationBody holds a Go template per location field. A field left out uses its bmc.location default from the config, an empty one is left unchanged, and one whose template renders empty is cleared
type BmcLocationBody struct {
	Datacenter *string `json:"datacenter" oai3:"nullable"`
	Room       *string `json:"room" oai3:"nullable"`
	Row        *string `json:"row" oai3:"nullable"`
	Rack       *string `json:"rack" oai3:"nullable" example:"{{ index .Parts 1 }}"`
	RackUnit   *string `json:"rack_unit" oai3:"nullable" description:"Must render to a whole number" example:"{{ index .Parts 2 }}"`
	DryRun     bool    `json:"dry_run" description:"Read the bmcs and report what would change, without writing"`
}

// BmcIdentityDefaults are the templates configured under bmc.location and bmc.asset_tag
type BmcIdentityDefaults struct {
	Datacenter string `json:"datacenter"`
	Room       string `json:"room"`
	Row        string `json:"row"`
	Rack       string `json:"rack"`
	RackUnit   string `json:"rack_unit"`
	AssetTag   string `json:"asset_tag"`
}

func identityDefaults() BmcIdentityDefaults {
	return BmcIdentityDefaults{
		Datacenter: viper.GetString("bmc.location.datacenter"),
		Room:       viper.GetString("bmc.location.room"),
		Row:        viper.GetString("bmc.location.row"),
		Rack:       viper.GetString("bmc.location.rack"),
		RackUnit:   viper.GetString("bmc.location.rack_unit"),
		AssetTag:   viper.GetString("bmc.asset_tag"),
	}
}

// BmcIdentityDefaults returns the configured location and asset tag templates
func (h *Handler) BmcIdentityDefaults(c fuego.ContextNoBody) (BmcIdentityDefaults, error) {
	return identityDefaults(), nil
}

// BmcAssetTag sets the asset tag of node(s)
func (h *Handler) BmcAssetTag(c fuego.ContextWithBody[BmcAssetTagBody]) (model.BmcResponse[model.BmcAssetTagResult], error) {
	body, err := c.Body()
	if err != nil {
		return model.BmcResponse[model.BmcAssetTagResult]{}, badRequest(err, "failed to parse the asset tag request")
	}

	req := redfish.AssetTagRequest{DryRun: body.DryRun}
	var operation string

	switch {
	case body.Template != nil && body.BySerial != nil:
		return model.BmcResponse[model.BmcAssetTagResult]{}, badRequest(errors.New("template and by_serial both given"), "give either a template or by_serial, not both")
	case body.BySerial != nil:
		if len(body.BySerial) == 0 {
			return model.BmcResponse[model.BmcAssetTagResult]{}, badRequest(errors.New("by_serial is empty"), "by_serial must list at least one serial")
		}
		req.BySerial = body.BySerial
		operation = fmt.Sprintf("bmc asset tag set from a table of %d serial(s)", len(body.BySerial))
	default:
		text := identityDefaults().AssetTag
		if body.Template != nil {
			text = *body.Template
		}
		if text == "" {
			return model.BmcResponse[model.BmcAssetTagResult]{}, badRequest(errors.New("no asset tag given"), "give a template or by_serial, or set bmc.asset_tag in the config")
		}
		tmpl, err := parseIdentityTemplate("asset tag", text)
		if err != nil {
			return model.BmcResponse[model.BmcAssetTagResult]{}, err
		}
		hosts, err := h.bmcSelect(c)
		if err != nil {
			return model.BmcResponse[model.BmcAssetTagResult]{}, err
		}
		req.ByNode, req.Invalid = map[string]string{}, map[string]error{}
		hosts = identityHosts(hosts)
		for i, host := range hosts {
			v, err := renderIdentity(tmpl, host, position{i, len(hosts)})
			if err != nil {
				req.Invalid[host.Name] = err
				continue
			}
			req.ByNode[host.Name] = v
		}
		// the template, not the tags it produced, keeps the event short
		operation = "bmc asset tag set to " + text
	}

	return runBmc(h, c, operation, !body.DryRun,
		func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.AssetTagResult] {
			return h.Redfish.SetAssetTag(ctx, t, req)
		},
		assetTagData,
	)
}

// BmcLocation sets the rack location of node(s), rendered per node from templates
func (h *Handler) BmcLocation(c fuego.ContextWithBody[BmcLocationBody]) (model.BmcResponse[model.BmcLocationResult], error) {
	body, err := c.Body()
	if err != nil {
		return model.BmcResponse[model.BmcLocationResult]{}, badRequest(err, "failed to parse the location request")
	}

	fields, err := locationTemplates(body, identityDefaults())
	if err != nil {
		return model.BmcResponse[model.BmcLocationResult]{}, err
	}

	hosts, err := h.bmcSelect(c)
	if err != nil {
		return model.BmcResponse[model.BmcLocationResult]{}, err
	}

	req := redfish.LocationRequest{ByNode: map[string]redfish.LocationUpdate{}, Invalid: map[string]error{}, DryRun: body.DryRun}
	hosts = identityHosts(hosts)
	for i, host := range hosts {
		u, err := renderLocation(fields, host, position{i, len(hosts)})
		if err != nil {
			req.Invalid[host.Name] = err
			continue
		}
		req.ByNode[host.Name] = u
	}

	names := make([]string, 0, len(fields))
	for _, f := range fields {
		names = append(names, f.name)
	}
	return runBmc(h, c, "bmc location set "+strings.Join(names, ","), !body.DryRun,
		func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.LocationResult] {
			return h.Redfish.SetLocation(ctx, t, req)
		},
		locationData,
	)
}

// a failed node has no data, and runBmc maps every result
func assetTagData(data *redfish.AssetTagResult) *model.BmcAssetTagResult {
	if data == nil {
		return nil
	}
	return &data.BmcAssetTagResult
}

func locationData(data *redfish.LocationResult) *model.BmcLocationResult {
	if data == nil {
		return nil
	}
	return &data.BmcLocationResult
}

// position is a node's place among the nodes a request selected, so a template can number them
type position struct{ index, count int }

// identityHosts keeps the hosts that will be contacted, in natural name order, which is the order .Index counts in. A host without a bmc interface gets no index, so the numbering has no gaps
func identityHosts(hosts model.HostList) model.HostList {
	out := make(model.HostList, 0, len(hosts))
	for _, h := range hosts {
		if h != nil && h.InterfaceBMC() != nil {
			out = append(out, h)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return natsort.Compare(out[i].Name, out[j].Name) })
	return out
}

type locationField struct {
	name string
	tmpl *template.Template
}

// locationTemplates picks each field's template, the request's or else the config's, and parses it. A request that sets nothing is rejected
func locationTemplates(body BmcLocationBody, defaults BmcIdentityDefaults) ([]locationField, error) {
	pick := func(v *string, def string) string {
		if v != nil {
			return *v
		}
		return def
	}
	texts := []struct{ name, text string }{
		{"datacenter", pick(body.Datacenter, defaults.Datacenter)},
		{"room", pick(body.Room, defaults.Room)},
		{"row", pick(body.Row, defaults.Row)},
		{"rack", pick(body.Rack, defaults.Rack)},
		{"rack_unit", pick(body.RackUnit, defaults.RackUnit)},
	}

	var fields []locationField
	for _, t := range texts {
		if t.text == "" {
			continue
		}
		tmpl, err := parseIdentityTemplate(t.name, t.text)
		if err != nil {
			return nil, err
		}
		fields = append(fields, locationField{name: t.name, tmpl: tmpl})
	}
	if len(fields) == 0 {
		return nil, badRequest(errors.New("no location fields given"), "give at least one location field, or set defaults under bmc.location in the config")
	}
	return fields, nil
}

func renderLocation(fields []locationField, host *model.Host, pos position) (redfish.LocationUpdate, error) {
	var u redfish.LocationUpdate
	for _, f := range fields {
		v, err := renderIdentity(f.tmpl, host, pos)
		if err != nil {
			return u, err
		}
		switch f.name {
		case "datacenter":
			u.Datacenter = &v
		case "room":
			u.Room = &v
		case "row":
			u.Row = &v
		case "rack":
			u.Rack = &v
		case "rack_unit":
			u.SetRackUnit = true
			if v == "" {
				continue
			}
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				return u, fmt.Errorf("template rack_unit: %q is not a whole number", v)
			}
			u.RackUnit = &n
		}
	}
	return u, nil
}

// identityData is what a location or asset tag template sees. Parts is the node name split on -, so cpn-f01-04 gives [cpn f01 04]. Tags holds the key=value tags by key
type identityData struct {
	Index   int
	Count   int
	Name    string
	Parts   []string
	Tags    map[string]string
	TagList []string
}

var identityFuncs = template.FuncMap{
	"split":      strings.Split,
	"join":       func(sep string, s []string) string { return strings.Join(s, sep) },
	"trimPrefix": func(prefix, s string) string { return strings.TrimPrefix(s, prefix) },
	"trimSuffix": func(suffix, s string) string { return strings.TrimSuffix(s, suffix) },
	"trimLeft":   func(cutset, s string) string { return strings.TrimLeft(s, cutset) },
	"replace":    func(old, new, s string) string { return strings.ReplaceAll(s, old, new) },
	"upper":      strings.ToUpper,
	"lower":      strings.ToLower,
	"default": func(def, s string) string {
		if s == "" {
			return def
		}
		return s
	},
	"atoi": strconv.Atoi,
	"add":  func(a, b int) int { return a + b },
}

func parseIdentityTemplate(name, text string) (*template.Template, error) {
	funcs := template.FuncMap{"tag": func(string) (string, error) { return "", nil }}
	for k, v := range identityFuncs {
		funcs[k] = v
	}
	tmpl, err := template.New(name).Funcs(funcs).Option("missingkey=error").Parse(text)
	if err != nil {
		return nil, badRequest(err, fmt.Sprintf("the %s template does not parse", name))
	}
	return tmpl, nil
}

// renderIdentity runs a template for one host. tag looks up a key=value tag and fails when the host lacks it, so a node missing its rack tag is reported rather than given an empty rack
func renderIdentity(tmpl *template.Template, host *model.Host, pos position) (string, error) {
	data := identityData{Index: pos.index, Count: pos.count, Name: host.Name, Parts: strings.Split(host.Name, "-"), Tags: map[string]string{}, TagList: host.Tags}
	for _, t := range host.Tags {
		if k, v, ok := strings.Cut(t, "="); ok {
			data.Tags[k] = v
		}
	}

	t, err := tmpl.Clone()
	if err != nil {
		return "", err
	}
	t.Funcs(template.FuncMap{"tag": func(key string) (string, error) {
		v, ok := data.Tags[key]
		if !ok {
			return "", fmt.Errorf("node has no %s= tag", key)
		}
		return v, nil
	}})

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template %s: %w", tmpl.Name(), err)
	}
	return strings.TrimSpace(buf.String()), nil
}
