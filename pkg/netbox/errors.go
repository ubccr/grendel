// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package netbox

import "errors"

var (
	// ErrNotFound is returned when an object is not found in NetBox
	ErrNotFound = errors.New("not found")

	// ErrBadHttpStatus is returned when an api call to NetBox does not return 200
	ErrBadHttpStatus = errors.New("bad http status code")
)
