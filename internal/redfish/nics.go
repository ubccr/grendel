// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"errors"
	"net"

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/schemas"
	"github.com/ubccr/grendel/pkg/model"
)

func (s *Service) nicsOne(ctx context.Context, t Target) (*model.BmcNICList, error) {
	c, err := s.connect(ctx, t)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	systems, err := c.service.Systems()
	if err != nil {
		return nil, err
	}
	if len(systems) == 0 {
		return nil, errors.New("bmc exposes no computer systems")
	}

	ifaces, err := systems[0].EthernetInterfaces()
	if err != nil {
		log.Debugf("node %s: reading ethernet interfaces: %s", t.Name, err)
	}
	out := &model.BmcNICList{NICs: systemNICs(ifaces)}
	if len(out.NICs) > 0 {
		return out, nil
	}

	// iDRACs can report no system ethernet interfaces while the host is off, but still list the adapters under the chassis
	chassis, err := systems[0].Chassis()
	if err != nil {
		return out, nil
	}
	var adapters []*schemas.NetworkAdapter
	for _, ch := range chassis {
		a, err := ch.NetworkAdapters()
		if err != nil {
			log.Debugf("node %s: reading network adapters of %s: %s", t.Name, ch.ID, err)
			continue
		}
		adapters = append(adapters, a...)
	}
	for _, a := range adapters {
		funcs, err := a.NetworkDeviceFunctions()
		if err != nil {
			continue
		}
		for _, f := range funcs {
			nic := deviceFunctionNIC(f)
			if nic.MAC == "" {
				continue
			}
			if port, err := f.PhysicalPortAssignment(); err == nil && port != nil {
				nic.Link = string(port.LinkStatus)
			}
			out.NICs = append(out.NICs, nic)
		}
	}
	return out, nil
}

func systemNICs(ifaces []*schemas.EthernetInterface) []model.BmcNIC {
	out := []model.BmcNIC{}
	for _, i := range ifaces {
		mac := normalizeMAC(i.PermanentMACAddress, i.MACAddress)
		if mac == "" {
			continue
		}
		out = append(out, model.BmcNIC{
			ID:        i.ID,
			Name:      i.Name,
			MAC:       mac,
			Link:      string(i.LinkStatus),
			SpeedMbps: gofish.Deref(i.SpeedMbps),
			Enabled:   i.InterfaceEnabled,
		})
	}
	return out
}

func deviceFunctionNIC(f *schemas.NetworkDeviceFunction) model.BmcNIC {
	return model.BmcNIC{
		ID:      f.ID,
		Name:    f.Name,
		MAC:     normalizeMAC(f.Ethernet.PermanentMACAddress, f.Ethernet.MACAddress),
		Enabled: f.DeviceEnabled,
	}
}

// normalizeMAC picks the first MAC that parses and writes it lower case with colons, the form Grendel stores, since BMCs report upper case or dashes
func normalizeMAC(macs ...string) string {
	for _, m := range macs {
		if hw, err := net.ParseMAC(m); err == nil && len(hw) == 6 {
			return hw.String()
		}
	}
	return ""
}
