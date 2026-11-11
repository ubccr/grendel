// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-faster/jx"
	"github.com/ubccr/grendel/pkg/client"
	"github.com/ubccr/grendel/pkg/model"
)

func TestBiosChanges(t *testing.T) {
	names, err := biosChanges(BmcSetBiosBody{Attributes: map[string]any{"ProcCStates": "Disabled", "SriovGlobalEnable": "Enabled", "MemTest": true, "Port": float64(3)}, ApplyTime: "OnReset"})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(names, ","); got != "MemTest,Port,ProcCStates,SriovGlobalEnable" {
		t.Errorf("names = %s, want them sorted for the event log", got)
	}

	refused := map[string]BmcSetBiosBody{
		"no attributes":     {},
		"path in a name":    {Attributes: map[string]any{"../Accounts": "x"}},
		"space in a name":   {Attributes: map[string]any{"Proc CStates": "x"}},
		"object value":      {Attributes: map[string]any{"ProcCStates": map[string]any{"a": 1}}},
		"null value":        {Attributes: map[string]any{"ProcCStates": nil}},
		"unknown applytime": {Attributes: map[string]any{"ProcCStates": "x"}, ApplyTime: "Later"},
	}
	for name, body := range refused {
		if _, err := biosChanges(body); err == nil {
			t.Errorf("%s: expected a 400", name)
		}
	}
}

func TestGeneratedClientDecodesBios(t *testing.T) {
	upper := int64(7)
	body, err := json.Marshal(model.BmcResponse[model.BmcBiosRegistry]{RunID: "run", Results: []model.BmcResult[model.BmcBiosRegistry]{
		{Node: "cpn-1", Success: true, Data: &model.BmcBiosRegistry{Available: true, Attributes: []model.BmcBiosAttribute{
			{Name: "Port", Type: "Integer", Values: []model.BmcBiosValue{}, UpperBound: &upper},
			{Name: "SriovGlobalEnable", Type: "Enumeration", Values: []model.BmcBiosValue{{Name: "Enabled", DisplayName: "Enabled"}}, DefaultValue: "Disabled"},
		}}},
	}})
	if err != nil {
		t.Fatal(err)
	}

	var reg client.BmcResponseModelBmcBiosRegistry
	if err := reg.Decode(jx.DecodeBytes(body)); err != nil {
		t.Fatalf("decoding the registry %s: %v", body, err)
	}

	body, err = json.Marshal(model.BmcResponse[model.BmcBios]{RunID: "run", Results: []model.BmcResult[model.BmcBios]{
		{Node: "cpn-1", Success: true, Data: &model.BmcBios{Attributes: map[string]any{"Port": float64(3), "MemTest": true, "NumLock": "On"}, Pending: map[string]any{}, ApplyTimes: []string{"OnReset"}}},
		{Node: "cpn-2", Message: "boom"},
	}})
	if err != nil {
		t.Fatal(err)
	}

	var bios client.BmcResponseModelBmcBios
	if err := bios.Decode(jx.DecodeBytes(body)); err != nil {
		t.Fatalf("decoding the attributes %s: %v", body, err)
	}
	if got := string(bios.Results.Value[0].Value.Data.Value.Attributes.Value["Port"]); got != "3" {
		t.Errorf("Port = %s", got)
	}
}
