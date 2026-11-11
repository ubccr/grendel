// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import "time"

// Ack is the payload of an operation that reports nothing but its outcome. Note explains any place where what was done differs from what was asked
type Ack struct {
	Note string
}

func (a Ack) note() string { return a.Note }

// noter lets an operation attach a sentence to a successful result without the engine knowing the payload type
type noter interface{ note() string }

type NodeResult[T any] struct {
	Node     string
	Success  bool
	Message  string
	Duration time.Duration
	Data     *T
}
