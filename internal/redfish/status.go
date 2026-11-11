// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"errors"

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/schemas"
	"github.com/ubccr/grendel/pkg/model"
)

func (s *Service) statusOne(ctx context.Context, t Target) (*model.RedfishSystem, error) {
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

	sys := systems[0]

	system := &model.RedfishSystem{
		Name:           t.Name,
		HostName:       sys.HostName,
		BIOSVersion:    sys.BiosVersion,
		SerialNumber:   c.serialNumber(sys),
		Manufacturer:   sys.Manufacturer,
		Model:          sys.Model,
		PowerStatus:    string(sys.PowerState),
		Health:         string(sys.Status.Health),
		TotalMemory:    float32(gofish.Deref(sys.MemorySummary.TotalSystemMemoryGiB)),
		ProcessorCount: int(gofish.Deref(sys.ProcessorSummary.LogicalProcessorCount)),
		ProcessorModel: sys.ProcessorSummary.Model,

		ProcessorSockets: int(gofish.Deref(sys.ProcessorSummary.Count)),
		ProcessorCores:   int(gofish.Deref(sys.ProcessorSummary.CoreCount)),

		BootProgress:     bootProgress(sys.BootProgress),
		BootProgressTime: sys.BootProgress.LastStateTime,

		BootOverrideTarget:  string(sys.Boot.BootSourceOverrideTarget),
		BootOverrideEnabled: string(sys.Boot.BootSourceOverrideEnabled),
		BootMode:            string(sys.Boot.BootSourceOverrideMode),
		BootNext:            sys.Boot.BootNext,
		BootOrder:           sys.Boot.BootOrder,
	}

	// a failure parsing Dell's extra block should not cost the operator the standard fields above
	if c.isDell() {
		oem, err := dellOEMSystem(sys)
		if err != nil {
			log.Debugf("node %s reports Dell but its oem block did not parse: %s", t.Name, err)
		} else {
			system.OEMDell = &oem
		}
	}

	return system, nil
}

// bootProgress names the last boot state. A vendor state arrives as LastState OEM with its name in OEMLastState
func bootProgress(p schemas.BootProgress) string {
	if p.LastState == schemas.OEMBootProgressTypes && p.OEMLastState != "" {
		return p.OEMLastState
	}
	return string(p.LastState)
}
