// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package settings

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// viper only loads the config file, in cmd/shared. A read of it anywhere else would miss the database and race with a setting being changed.
func TestNothingElseReadsViper(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, dir := range []string{"cmd", "internal", "pkg", "test"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == "node_modules" || d.Name() == "frontend" {
					return filepath.SkipDir
				}
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			if !strings.HasSuffix(path, ".go") || strings.HasPrefix(rel, filepath.Join("cmd", "shared")) {
				return nil
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, imp := range file.Imports {
				if name, _ := strconv.Unquote(imp.Path.Value); name == "github.com/spf13/viper" {
					t.Errorf("%s imports viper, read the value through internal/settings", rel)
				}
			}
			return nil
		})
		require.NoError(t, err)
	}
}

// the sample is written from the registry, so a new option shows up in it without anyone remembering to add it
func TestSampleConfigIsCurrent(t *testing.T) {
	committed, err := os.ReadFile(filepath.Join("..", "..", "configs", "grendel.toml.sample"))
	require.NoError(t, err)
	assert.Equal(t, sampleFile(), string(committed), "run `go generate ./internal/settings` to rewrite configs/grendel.toml.sample")
}
