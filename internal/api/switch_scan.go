// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/facette/natsort"
	"github.com/go-fuego/fuego"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/internal/tors"
	"github.com/ubccr/grendel/pkg/model"
	"golang.org/x/sync/errgroup"
)

// physicalPort matches front-panel port names on the supported NOSes, leaving out VLANs, port-channels, loopbacks and the management port
var physicalPort = regexp.MustCompile(`(?i)^(ethernet|eth|et)\d`)

type SwitchScanMAC struct {
	MAC   string `json:"mac" example:"d0:8e:79:00:00:07"`
	VLAN  string `json:"vlan" example:"3001"`
	Owner string `json:"owner,omitempty" description:"The node that already has this MAC on one of its ports"`
	BMC   bool   `json:"bmc" description:"The owner has this MAC on its BMC port"`
}

type SwitchScanPort struct {
	Ifname      string          `json:"ifname" example:"Ethernet7"`
	Description string          `json:"description,omitempty"`
	MACs        []SwitchScanMAC `json:"macs"`
}

type SwitchScanResult struct {
	Name      string           `json:"name" example:"swe-a01"`
	NOS       string           `json:"nos" example:"arista"`
	Address   string           `json:"address,omitempty"`
	Error     string           `json:"error,omitempty" description:"Why the MAC table could not be read. The switch's ports are empty"`
	ErrorCode string           `json:"error_code,omitempty" description:"unknown_nos, no_credentials, no_interfaces or unreachable"`
	Warnings  []string         `json:"warnings" oai3:"nullable" description:"An interface status failure, which leaves out the ports with no MACs"`
	Ports     []SwitchScanPort `json:"ports"`
}

type SwitchScanResponse struct {
	Switches []SwitchScanResult `json:"switches"`
}

// SwitchScan reads the MAC table and interface list of each selected switch, and reports per port which MACs it has learned and whether Grendel already knows them
func (h *Handler) SwitchScan(c fuego.ContextNoBody) (*SwitchScanResponse, error) {
	selected, err := h.bmcSelect(c)
	if err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return nil, notFound("no nodes matched the given nodeset and tags")
	}

	all, err := h.DB.Hosts()
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Title: "Error", Detail: "failed to load nodes"}
	}
	owners := macOwners(all)

	if err := http.NewResponseController(c.Response()).SetWriteDeadline(time.Time{}); err != nil {
		log.Debugf("could not clear the write deadline: %s", err)
	}

	cfg := tors.ConfigFromViper()
	out := make([]SwitchScanResult, len(selected))
	var g errgroup.Group
	g.SetLimit(max(viper.GetInt("bmc.fanout"), 1))
	for i, host := range selected {
		g.Go(func() error {
			ctx, cancel := context.WithTimeout(c.Context(), cfg.Timeout)
			defer cancel()
			out[i] = scanSwitch(ctx, host, cfg, owners)
			return nil
		})
	}
	g.Wait()

	slices.SortFunc(out, func(a, b SwitchScanResult) int { return natCmp(a.Name, b.Name) })
	return &SwitchScanResponse{Switches: out}, nil
}

func natCmp(a, b string) int {
	switch {
	case a == b:
		return 0
	case natsort.Compare(a, b):
		return -1
	}
	return 1
}

// newSwitch is tors.New, swapped out in tests
var newSwitch = func(host *model.Host, cfg tors.Config) (tors.NetworkSwitch, error) { return tors.New(host, cfg) }

// recovered reports a panic as an error. A scan runs on goroutines outside net/http's recovery, where a panic in a driver's parser would end the daemon
func recovered(r any) error {
	log.Errorf("switch scan: panic: %v\n%s", r, debug.Stack())
	return fmt.Errorf("panic: %v", r)
}

func guard[T any](read func() (T, error)) (out T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recovered(r)
		}
	}()
	return read()
}

func scanSwitch(ctx context.Context, host *model.Host, cfg tors.Config, owners map[string]macOwner) (res SwitchScanResult) {
	res = SwitchScanResult{Name: host.Name, NOS: tors.HostNOS(host), Ports: []SwitchScanPort{}, Warnings: []string{}}
	if len(host.Interfaces) > 0 {
		res.Address = host.Interfaces[0].AddrString()
	}
	fail := func(err error) SwitchScanResult {
		res.Error, res.ErrorCode = err.Error(), tors.ErrorCode(err)
		return res
	}
	defer func() {
		if r := recover(); r != nil {
			res = fail(recovered(r))
		}
	}()

	sw, err := newSwitch(host, cfg)
	if err != nil {
		return fail(err)
	}
	defer sw.Close()

	var (
		wg           sync.WaitGroup
		macs         model.MACTable
		descriptions map[string]string
		macErr       error
		descErr      error
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		macs, macErr = guard(func() (model.MACTable, error) { return sw.GetMACTable(ctx) })
	}()
	go func() {
		defer wg.Done()
		descriptions, descErr = guard(func() (map[string]string, error) { return sw.GetInterfaceDescriptions(ctx) })
	}()
	wg.Wait()

	if macErr != nil {
		log.Errorf("switch scan %s: mac table: %v", host.Name, macErr)
		return fail(fmt.Errorf("failed to read the MAC table: %w", macErr))
	}
	if descErr != nil {
		res.Warnings = append(res.Warnings, "interface status: "+descErr.Error())
	}

	res.Ports = scanPorts(macs, descriptions, owners)
	return res
}

// scanPorts merges what the switch reported into one row per physical port. Ports with no MACs are kept, since the import numbers ports over them so a BMC that hasn't come up doesn't shift the names after it
func scanPorts(macs model.MACTable, descriptions map[string]string, owners map[string]macOwner) []SwitchScanPort {
	ports := map[string]*SwitchScanPort{}
	port := func(name string) *SwitchScanPort {
		if !physicalPort.MatchString(name) {
			return nil
		}
		p, ok := ports[name]
		if !ok {
			p = &SwitchScanPort{Ifname: name, MACs: []SwitchScanMAC{}}
			ports[name] = p
		}
		return p
	}

	for name, description := range descriptions {
		if p := port(name); p != nil {
			p.Description = description
		}
	}
	for _, e := range macs {
		p := port(e.Ifname)
		if p == nil {
			continue
		}
		m := SwitchScanMAC{MAC: e.MAC.String(), VLAN: e.VLAN}
		if o, ok := owners[m.MAC]; ok {
			m.Owner, m.BMC = o.node, o.bmc
		}
		p.MACs = append(p.MACs, m)
	}

	out := make([]SwitchScanPort, 0, len(ports))
	for _, p := range ports {
		slices.SortFunc(p.MACs, func(a, b SwitchScanMAC) int {
			return cmp.Or(natCmp(a.VLAN, b.VLAN), strings.Compare(a.MAC, b.MAC))
		})
		out = append(out, *p)
	}
	slices.SortFunc(out, func(a, b SwitchScanPort) int { return natCmp(a.Ifname, b.Ifname) })
	return out
}
