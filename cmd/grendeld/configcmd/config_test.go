// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package configcmd

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readOnlyDir(t *testing.T) string {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("root can write anywhere")
	}
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "grendel.toml"), []byte("dbpath = 'x'\n"), 0o660))
	require.NoError(t, os.Chmod(dir, 0o555))
	t.Cleanup(func() { os.Chmod(dir, 0o755) })
	return dir
}

func names(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	out := []string{}
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out
}

func TestCanReplace(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "grendel.toml")
	require.NoError(t, os.WriteFile(source, []byte("dbpath = 'x'\n"), 0o660))

	require.NoError(t, canReplace(source))
	assert.Equal(t, []string{"grendel.toml"}, names(t, dir), "the check leaves nothing behind")

	require.NoError(t, os.WriteFile(source+".bak", nil, 0o600))
	assert.ErrorContains(t, canReplace(source), "grendel.toml.bak is already there")
}

func TestCanReplaceRefusesADirectoryItCannotWriteIn(t *testing.T) {
	dir := readOnlyDir(t)

	err := canReplace(filepath.Join(dir, "grendel.toml"))
	assert.ErrorIs(t, err, fs.ErrPermission)
	assert.ErrorContains(t, err, "sudo")
	assert.Equal(t, []string{"grendel.toml"}, names(t, dir))
}

func TestCanCreate(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, canCreate(filepath.Join(dir, "remaining.toml")))
	assert.Empty(t, names(t, dir), "the check leaves nothing behind")

	assert.ErrorIs(t, canCreate(filepath.Join(readOnlyDir(t), "remaining.toml")), fs.ErrPermission)
	assert.ErrorIs(t, canCreate(filepath.Join(dir, "missing", "remaining.toml")), fs.ErrNotExist)
}

func TestCheckAccess(t *testing.T) {
	dir := t.TempDir()
	config, db := filepath.Join(dir, "grendel.toml"), filepath.Join(dir, "grendel.db")
	require.NoError(t, os.WriteFile(config, nil, 0o600))

	err := checkAccess(config, db)
	assert.ErrorContains(t, err, "there is no database at "+db)
	assert.NoFileExists(t, db, "a missing database is not created")

	require.NoError(t, os.WriteFile(db, nil, 0o600))
	assert.NoError(t, checkAccess(config, db))
	assert.NoError(t, checkAccess("", ":memory:"), "the in-memory database is refused later, with its own message")

	if os.Geteuid() != 0 {
		require.NoError(t, os.Chmod(config, 0))
		err := checkAccess(config, db)
		assert.ErrorIs(t, err, fs.ErrPermission)
		assert.ErrorContains(t, err, "sudo")
	}
}
