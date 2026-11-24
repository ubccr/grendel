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
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
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

// the sample is written by hand and installed as the config of a packaged or container install, so it has to load cleanly and hold nothing the database could
func TestSampleConfigHoldsOnlyFileSettings(t *testing.T) {
	text, err := os.ReadFile(filepath.Join("..", "..", "configs", "grendel.toml.sample"))
	require.NoError(t, err)

	// the settings it leaves commented out are checked too, by reading it with them turned on
	on := regexp.MustCompile(`(?m)^#(\[|[a-z_]+ = )`).ReplaceAllString(string(text), "$1")
	var doc map[string]any
	require.NoError(t, toml.Unmarshal([]byte(on), &doc))

	var keys []string
	var walk func(prefix string, m map[string]any)
	walk = func(prefix string, m map[string]any) {
		for name, v := range m {
			if sub, ok := v.(map[string]any); ok {
				walk(prefix+name+".", sub)
				continue
			}
			keys = append(keys, prefix+name)
		}
	}
	walk("", doc)
	lookup := func(key string) (any, bool) {
		var v any = doc
		for _, part := range strings.Split(key, ".") {
			m, ok := v.(map[string]any)
			if !ok {
				return nil, false
			}
			if v, ok = m[part]; !ok {
				return nil, false
			}
		}
		return v, true
	}

	Reset()
	t.Cleanup(Reset)
	plan, err := PlanImport(File{Path: "grendel.toml.sample", Keys: keys, Lookup: lookup}, nil)
	require.NoError(t, err)
	require.NotEmpty(t, plan.Items)
	for _, item := range plan.Items {
		switch item.Action {
		case ImportUnknown, ImportInvalid, ImportDeprecated:
			t.Errorf("%s: %s", item.Key, item.Note)
		case ImportReady:
			t.Errorf("%s can be stored in the database, so it does not belong in the sample", item.Key)
		}
	}
}
