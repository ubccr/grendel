// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubccr/grendel/internal/tors"
	"github.com/ubccr/grendel/pkg/model"
)

type panickingSwitch struct {
	fakeSwitch
	inMACs, inDescriptions bool
}

func (p panickingSwitch) GetMACTable(ctx context.Context) (model.MACTable, error) {
	if p.inMACs {
		panic("bad mac table")
	}
	return p.fakeSwitch.GetMACTable(ctx)
}

func (p panickingSwitch) GetLLDPNeighbors(context.Context) (model.LLDPNeighbors, error) {
	if p.inMACs {
		panic("bad lldp table")
	}
	return nil, nil
}

func (p panickingSwitch) GetInterfaceDescriptions(ctx context.Context) (map[string]string, error) {
	if p.inDescriptions {
		panic("bad interface table")
	}
	return p.fakeSwitch.GetInterfaceDescriptions(ctx)
}

func TestScanSwitchSurvivesDriverPanic(t *testing.T) {
	old := newSwitch
	t.Cleanup(func() { newSwitch = old })
	host := &model.Host{Name: "swe-a01", Tags: []string{"arista"}}

	newSwitch = func(*model.Host, tors.Config) (tors.NetworkSwitch, error) {
		return panickingSwitch{inMACs: true}, nil
	}
	res := scanSwitch(context.Background(), host, tors.Config{}, nil)
	assert.Equal(t, "failed to read the MAC table: panic: bad mac table", res.Error)
	assert.Equal(t, tors.CodeUnreachable, res.ErrorCode)

	newSwitch = func(*model.Host, tors.Config) (tors.NetworkSwitch, error) {
		return panickingSwitch{fakeSwitch: fakeSwitch{macs: model.MACTable{{Ifname: "Ethernet1", MAC: mac("d0:8e:79:00:00:01")}}}, inDescriptions: true}, nil
	}
	res = scanSwitch(context.Background(), host, tors.Config{}, nil)
	assert.Empty(t, res.Error)
	assert.Equal(t, []string{"interface status: panic: bad interface table"}, res.Warnings)
	assert.Len(t, res.Ports, 1)

	newSwitch = func(*model.Host, tors.Config) (tors.NetworkSwitch, error) { panic("bad config") }
	res = scanSwitch(context.Background(), host, tors.Config{}, nil)
	assert.Equal(t, "swe-a01", res.Name)
	assert.Equal(t, "panic: bad config", res.Error)
}
