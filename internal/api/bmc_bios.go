// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/stmcginnis/gofish/schemas"
	"github.com/ubccr/grendel/internal/redfish"
	"github.com/ubccr/grendel/pkg/model"
)

type BmcSetBiosBody struct {
	Attributes map[string]any `json:"attributes" description:"Attribute names and their new values. A string value is converted to the number or boolean the attribute holds"`
	ApplyTime  string         `json:"apply_time" description:"When the bmc applies the change. Immediate, OnReset, AtMaintenanceWindowStart or InMaintenanceWindowOnReset, limited to what each bmc accepts. Empty for the bmc's default" example:"OnReset"`
}

func (h *Handler) BmcBios(c fuego.ContextNoBody) (model.BmcResponse[model.BmcBios], error) {
	var names []string
	if raw := c.QueryParam("names"); raw != "" {
		for _, name := range strings.Split(raw, ",") {
			if name = strings.TrimSpace(name); name != "" {
				names = append(names, name)
			}
		}
	}

	return runBmc(h, c, "bmc bios", nil,
		func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[model.BmcBios] {
			return h.Redfish.Bios(ctx, t, redfish.BiosRequest{Names: names})
		},
		func(data *model.BmcBios) *model.BmcBios { return data },
	)
}

func (h *Handler) BmcSetBios(c fuego.ContextWithBody[BmcSetBiosBody]) (model.BmcResponse[model.BmcAck], error) {
	body, err := c.Body()
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, err
	}

	names, err := biosChanges(body)
	if err != nil {
		return model.BmcResponse[model.BmcAck]{}, err
	}

	// the event names the attributes but never their values, which can include a bios password
	operation := "bmc bios set " + strings.Join(names, ",")
	if body.ApplyTime != "" {
		operation += ", apply time " + body.ApplyTime
	}

	req := redfish.SetBiosRequest{Attributes: body.Attributes, ApplyTime: body.ApplyTime}
	return runAck(h, c, "bmc.bios.set", operation, func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return h.Redfish.SetBios(ctx, t, req)
	})
}

func (h *Handler) BmcDiscardBios(c fuego.ContextNoBody) (model.BmcResponse[model.BmcAck], error) {
	return runAck(h, c, "bmc.bios.discard", "bmc bios discard", func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[redfish.Ack] {
		return h.Redfish.DiscardPendingBios(ctx, t)
	})
}

// A single node route because a registry runs to thousands of entries, and the nodes of one model share it
func (h *Handler) BmcBiosRegistry(c fuego.ContextNoBody) (model.BmcResponse[model.BmcBiosRegistry], error) {
	nodes, err := h.bmcNodes(c)
	if err != nil {
		return model.BmcResponse[model.BmcBiosRegistry]{}, err
	}
	if n := len(nodes.hosts) + len(nodes.noBMC); n != 1 {
		return model.BmcResponse[model.BmcBiosRegistry]{}, badRequest(fmt.Errorf("%d nodes selected", n), "the bios registry is read from one node at a time")
	}

	return runBmcOn(h, c, nodes, "bmc bios registry", nil,
		func(ctx context.Context, t []redfish.Target) []redfish.NodeResult[model.BmcBiosRegistry] {
			return h.Redfish.BiosRegistry(ctx, t)
		},
		func(data *model.BmcBiosRegistry) *model.BmcBiosRegistry { return data },
	)
}

var biosNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

var settingsApplyTimes = []schemas.SettingsApplyTime{
	schemas.ImmediateSettingsApplyTime,
	schemas.OnResetSettingsApplyTime,
	schemas.AtMaintenanceWindowStartSettingsApplyTime,
	schemas.InMaintenanceWindowOnResetSettingsApplyTime,
}

// biosChanges validates a set request before any node is contacted, and returns the attribute names in order
func biosChanges(body BmcSetBiosBody) ([]string, error) {
	if len(body.Attributes) == 0 {
		return nil, badRequest(errors.New("no attributes given"), "attributes must name at least one bios attribute")
	}

	names := make([]string, 0, len(body.Attributes))
	for name, value := range body.Attributes {
		if !biosNamePattern.MatchString(name) {
			return nil, badRequest(fmt.Errorf("invalid attribute name %q", name), "bios attribute names may only contain letters, numbers, and . _ -")
		}
		switch value.(type) {
		case string, float64, bool:
		default:
			return nil, badRequest(fmt.Errorf("attribute %s has a %T value", name, value), "bios attribute values must be a string, number or boolean")
		}
		names = append(names, name)
	}
	sort.Strings(names)

	if body.ApplyTime != "" && !slices.Contains(settingsApplyTimes, schemas.SettingsApplyTime(body.ApplyTime)) {
		return nil, badRequest(fmt.Errorf("unknown apply time %q", body.ApplyTime), "apply_time must be Immediate, OnReset, AtMaintenanceWindowStart or InMaintenanceWindowOnReset")
	}

	return names, nil
}
