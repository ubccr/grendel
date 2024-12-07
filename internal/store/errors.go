// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package store

import "errors"

var (
	// ErrNotFound is returned when a model is not found in the store
	ErrNotFound = errors.New("not found")

	// ErrInvalidData is returned when a model is is missing required data
	ErrInvalidData = errors.New("invalid data")

	// ErrDuplicateEntry is returned when attempting to store a model with the same ID or Name
	ErrDuplicateEntry = errors.New("duplicate entry")
)
