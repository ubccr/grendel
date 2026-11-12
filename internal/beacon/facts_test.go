// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package beacon

import (
	"context"
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

func TestBootFacts(t *testing.T) {
	assert.Nil(t, bootFacts(map[string]uint64{"UserspaceTimestampMonotonic": 5_000_000}), "nothing until the boot finishes")

	got := bootFacts(map[string]uint64{
		"FirmwareTimestampMonotonic":  20_000_000,
		"LoaderTimestampMonotonic":    3_000_000,
		"InitRDTimestampMonotonic":    2_500_000,
		"UserspaceTimestampMonotonic": 10_000_000,
		"FinishTimestampMonotonic":    40_123_456,
	})
	assert.Equal(t, map[string]string{
		"boot.firmware":  "17s",
		"boot.loader":    "3s",
		"boot.kernel":    "2.5s",
		"boot.initrd":    "7.5s",
		"boot.userspace": "30.123s",
		"boot.total":     "1m0.123s",
	}, got)

	// a boot without an initrd and on a bios that doesn't report its own time
	got = bootFacts(map[string]uint64{"UserspaceTimestampMonotonic": 1_000_000, "FinishTimestampMonotonic": 4_000_000})
	assert.Equal(t, map[string]string{"boot.kernel": "1s", "boot.userspace": "3s", "boot.total": "4s"}, got)
}

func TestCollect(t *testing.T) {
	dir := t.TempDir()
	osRelease := filepath.Join(dir, "os-release")
	require.NoError(t, os.WriteFile(osRelease, []byte("ID=rocky\nIMAGE_VERSION=2026.11.3\n"), 0o644))
	kernel := filepath.Join(dir, "osrelease")
	require.NoError(t, os.WriteFile(kernel, []byte("5.14.0-427.el9.x86_64\n"), 0o644))

	dropIns := filepath.Join(dir, "beacon.d")
	require.NoError(t, os.Mkdir(dropIns, 0o755))
	write := func(name, content string, mode os.FileMode) {
		require.NoError(t, os.WriteFile(filepath.Join(dropIns, name), []byte(content), mode))
	}
	write("site", "# static facts\nrack_pdu = pdu-f01-a\nBad Key=x\n", 0o644)
	write("slurm", "#!/bin/sh\necho version=24.05\necho partition=general\n", 0o755)
	write("failing", "#!/bin/sh\necho value=1\nexit 3\n", 0o755)
	write("slurm.rpmsave", "version=old\n", 0o644)
	write("os", "image_version=spoofed\n", 0o644)

	src := FactSource{OSRelease: []string{filepath.Join(dir, "missing"), osRelease}, Kernel: kernel, DropIns: dropIns}
	got := Collect(context.Background(), src, nil)

	assert.Equal(t, map[string]string{
		"os.id":            "rocky",
		"os.image_version": "2026.11.3",
		"kernel.release":   "5.14.0-427.el9.x86_64",
		"site.rack_pdu":    "pdu-f01-a",
		"slurm.version":    "24.05",
		"slurm.partition":  "general",
	}, got, "a failing script, a backup file and a drop-in shadowing a built in fact are all left out")
}

func TestCollectWithoutDropIns(t *testing.T) {
	got := Collect(context.Background(), FactSource{DropIns: filepath.Join(t.TempDir(), "none")}, nil)
	assert.Empty(t, got)
}
