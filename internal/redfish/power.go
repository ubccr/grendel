// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"errors"

	"github.com/stmcginnis/gofish/schemas"
)

func (s *Service) powerOne(ctx context.Context, t Target, req PowerRequest) (*Ack, error) {
	c, err := s.connect(ctx, t)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	if req.Target == TargetBMC {
		return s.resetManagers(c)
	}
	return s.resetSystems(c, req)
}

func (s *Service) resetSystems(c *conn, req PowerRequest) (*Ack, error) {
	systems, err := c.service.Systems()
	if err != nil {
		return nil, err
	}
	if len(systems) == 0 {
		return nil, errors.New("bmc exposes no computer systems")
	}

	if req.Boot != "" && req.Boot != schemas.NoneBootSource {
		boot := schemas.Boot{
			BootSourceOverrideTarget:  req.Boot,
			BootSourceOverrideEnabled: schemas.OnceBootSourceOverrideEnabled,
		}
		for _, sys := range systems {
			if err := sys.SetBoot(&boot); err != nil {
				return nil, err
			}
		}
	}

	substituted := false
	for _, sys := range systems {
		reset := req.Reset

		if sys.PowerState == schemas.OffPowerState && reset == schemas.ForceRestartResetType {
			reset = schemas.OnResetType
			substituted = true
		}

		if _, err := sys.Reset(reset); err != nil {
			return nil, err
		}
	}

	if substituted {
		return doneWith("node was powered off, sent %s instead of %s", schemas.OnResetType, req.Reset)
	}
	return done()
}

func (s *Service) resetManagers(c *conn) (*Ack, error) {
	managers, err := c.service.Managers()
	if err != nil {
		return nil, err
	}
	if len(managers) == 0 {
		return nil, errors.New("bmc exposes no managers")
	}

	for _, m := range managers {
		if _, err := m.Reset(schemas.GracefulRestartResetType); err != nil {
			return nil, err
		}
	}

	return done()
}
