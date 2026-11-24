// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settingstest

import (
	"testing"

	"github.com/ubccr/grendel/internal/settings"
)

// Set pins a setting for the rest of the test
func Set(t testing.TB, key string, value any) {
	t.Helper()
	restore, err := settings.SetOverride(key, value)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(restore)
}
