// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package provision

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchRecoversFromFailedReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.tmpl")
	require.NoError(t, os.WriteFile(path, []byte("v1"), 0o644))

	viper.Set("provision.templates_dir", dir)
	t.Cleanup(func() { viper.Set("provision.templates_dir", nil) })

	renderer, err := NewTemplateRenderer()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	require.NoError(t, renderer.Watch(ctx))

	render := func() string {
		var buf bytes.Buffer
		if err := renderer.templates.Load().ExecuteTemplate(&buf, "custom.tmpl", nil); err != nil {
			return err.Error()
		}
		return buf.String()
	}

	require.NoError(t, os.WriteFile(path, []byte("v2"), 0o644))
	require.Eventually(t, func() bool { return render() == "v2" }, 5*time.Second, 20*time.Millisecond, "valid edit should be reloaded")

	// Wait out the debounce so the broken and fixed writes are handled as separate reloads
	require.NoError(t, os.WriteFile(path, []byte("{{ if"), 0o644))
	time.Sleep(time.Second)
	assert.Equal(t, "v2", render(), "last good templates should keep serving after a failed reload")

	require.NoError(t, os.WriteFile(path, []byte("v3"), 0o644))
	assert.Eventually(t, func() bool { return render() == "v3" }, 5*time.Second, 20*time.Millisecond, "fixed template should be reloaded after a failed reload")
}
