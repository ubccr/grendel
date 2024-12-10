// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package migrations

import "errors"

var (
	ErrNoChange   = errors.New("no change")
	ErrNilVersion = errors.New("no migration")
)
