// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package model

import "time"

type EventList []Event

type Event struct {
	Severity    string
	Time        time.Time
	User        string
	Message     string
	JobMessages JobMessageList
}

type Severity string

const (
	SeveritySuccess Severity = "Success"
	SeverityInfo    Severity = "Info"
	SeverityWarning Severity = "Warning"
	SeverityError   Severity = "Error"
)

func (s Severity) String() string {
	return string(s)
}
