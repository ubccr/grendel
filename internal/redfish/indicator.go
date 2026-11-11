// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"errors"

	"github.com/stmcginnis/gofish/schemas"
)

// indicatorOne turns the location indicator on or off. LocationIndicatorActive is the current property and IndicatorLED the deprecated one older firmware still uses, so whichever the system reports is the one written. A system reporting neither falls back to its chassis, where some BMCs keep the indicator
func (s *Service) indicatorOne(ctx context.Context, t Target, req IndicatorRequest) (*Ack, error) {
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
	if payload := indicatorPayload(sys.LocationIndicatorActive, sys.IndicatorLED, req.On); payload != nil {
		return patchIndicator(c, sys.ODataID, payload, req.On)
	}

	chassis, err := sys.Chassis()
	if err != nil {
		return nil, err
	}
	for _, ch := range chassis {
		if payload := indicatorPayload(ch.LocationIndicatorActive, ch.IndicatorLED, req.On); payload != nil {
			return patchIndicator(c, ch.ODataID, payload, req.On)
		}
	}

	return nil, errors.New("bmc reports no location indicator")
}

// indicatorPayload builds the PATCH for whichever indicator property a resource reports, or nil when it reports neither
func indicatorPayload(active *bool, led schemas.IndicatorLED, on bool) map[string]any {
	if active != nil {
		return map[string]any{"LocationIndicatorActive": on}
	}
	if led != "" {
		state := schemas.OffIndicatorLED
		if on {
			state = schemas.BlinkingIndicatorLED
		}
		return map[string]any{"IndicatorLED": state}
	}
	return nil
}

func patchIndicator(c *conn, uri string, payload map[string]any, on bool) (*Ack, error) {
	res, err := c.client.Patch(uri, payload)
	schemas.DeferredCleanupHTTPResponse(res)
	if err != nil {
		return nil, err
	}

	if on {
		return doneWith("location indicator on")
	}
	return doneWith("location indicator off")
}

// indicatorState reads the location indicator as On or Off, empty when the system reports neither property
func indicatorState(sys *schemas.ComputerSystem) string {
	switch {
	case sys.LocationIndicatorActive != nil && *sys.LocationIndicatorActive:
		return "On"
	case sys.LocationIndicatorActive != nil:
		return "Off"
	case sys.IndicatorLED == schemas.LitIndicatorLED || sys.IndicatorLED == schemas.BlinkingIndicatorLED:
		return "On"
	case sys.IndicatorLED == schemas.OffIndicatorLED:
		return "Off"
	}
	return ""
}
