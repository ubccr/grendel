// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package beacon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOSRelease(t *testing.T) {
	got := parseOSRelease([]byte(`# comment
NAME="Rocky Linux"
ID=rocky
VERSION_ID='9.4'
PRETTY_NAME="Rocky \"Blue\" Onyx \$5"
IMAGE_VERSION=2026.11.3

broken line
`))
	assert.Equal(t, map[string]string{
		"NAME":          "Rocky Linux",
		"ID":            "rocky",
		"VERSION_ID":    "9.4",
		"PRETTY_NAME":   `Rocky "Blue" Onyx $5`,
		"IMAGE_VERSION": "2026.11.3",
	}, got)
}

func TestCollect(t *testing.T) {
	dir := t.TempDir()
	osRelease := filepath.Join(dir, "os-release")
	require.NoError(t, os.WriteFile(osRelease, []byte("ID=rocky\nIMAGE_VERSION=2026.11.3\n"), 0o644))
	kernel := filepath.Join(dir, "osrelease")
	require.NoError(t, os.WriteFile(kernel, []byte("5.14.0-427.el9.x86_64\n"), 0o644))

	got := Collect(FactSource{OSRelease: []string{filepath.Join(dir, "missing"), osRelease}, Kernel: kernel})
	assert.Equal(t, map[string]string{
		"os.id":            "rocky",
		"os.image_version": "2026.11.3",
		"kernel.release":   "5.14.0-427.el9.x86_64",
	}, got)

	assert.Empty(t, Collect(FactSource{}))
}
