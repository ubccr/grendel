// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

import "time"

var EventsRetention = register[time.Duration](Option{
	Key:     "events.retention",
	Section: "events",
	Summary: "Retention",
	Help:    "Events older than this are deleted. Set to 0 to keep them forever.",
	Kind:    KindDuration,
	Default: 2160 * time.Hour,
	Apply:   ApplyLive,
})
