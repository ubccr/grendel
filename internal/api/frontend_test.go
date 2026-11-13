// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testUI = fstest.MapFS{
	"ui/index.html":            {Data: []byte("<!doctype html><title>Grendel</title>" + strings.Repeat(" ", 2048))},
	"ui/assets/index-abc.js":   {Data: []byte("console.log('grendel');" + strings.Repeat("// padding\n", 200))},
	"ui/assets/tiny-abc.js":    {Data: []byte("1")},
	"ui/assets/font-abc.woff2": {Data: []byte(strings.Repeat("\x00", 4096))},
}

func serveUI(t *testing.T, target string, header map[string]string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	newFrontend(testUI).ServeHTTP(rec, req)
	return rec.Result()
}

func TestFrontendAssetsAreImmutable(t *testing.T) {
	res := serveUI(t, "/ui/assets/index-abc.js", nil)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "public, max-age=31536000, immutable", res.Header.Get("Cache-Control"))
	assert.Contains(t, res.Header.Get("Content-Type"), "javascript")
	assert.NotEmpty(t, res.Header.Get("ETag"))
}

func TestFrontendIndexRevalidates(t *testing.T) {
	for _, target := range []string{"/ui/", "/ui/index.html", "/ui/nodes/cpn-01/node"} {
		res := serveUI(t, target, nil)
		assert.Equal(t, http.StatusOK, res.StatusCode, target)
		assert.Equal(t, "no-cache", res.Header.Get("Cache-Control"), target)
		assert.Contains(t, res.Header.Get("Content-Type"), "text/html", target)
	}
}

func TestFrontendMissingAssetIs404(t *testing.T) {
	res := serveUI(t, "/ui/assets/index-old.js", nil)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	assert.Empty(t, res.Header.Get("Cache-Control"))
}

func TestFrontendGzip(t *testing.T) {
	plain := serveUI(t, "/ui/assets/index-abc.js", nil)
	plainBody, _ := io.ReadAll(plain.Body)
	assert.Empty(t, plain.Header.Get("Content-Encoding"))
	assert.Equal(t, "Accept-Encoding", plain.Header.Get("Vary"))

	res := serveUI(t, "/ui/assets/index-abc.js", map[string]string{"Accept-Encoding": "br, gzip"})
	require.Equal(t, "gzip", res.Header.Get("Content-Encoding"))
	assert.NotEqual(t, plain.Header.Get("ETag"), res.Header.Get("ETag"))
	zr, err := gzip.NewReader(res.Body)
	require.NoError(t, err)
	body, err := io.ReadAll(zr)
	require.NoError(t, err)
	assert.Equal(t, plainBody, body)

	refused := serveUI(t, "/ui/assets/index-abc.js", map[string]string{"Accept-Encoding": "gzip;q=0"})
	assert.Empty(t, refused.Header.Get("Content-Encoding"))
}

func TestFrontendSkipsGzipWhenNotWorthIt(t *testing.T) {
	for _, target := range []string{"/ui/assets/tiny-abc.js", "/ui/assets/font-abc.woff2"} {
		res := serveUI(t, target, map[string]string{"Accept-Encoding": "gzip"})
		assert.Empty(t, res.Header.Get("Content-Encoding"), target)
		assert.Empty(t, res.Header.Get("Vary"), target)
	}
}

func TestFrontendETagNotModified(t *testing.T) {
	first := serveUI(t, "/ui/", map[string]string{"Accept-Encoding": "gzip"})
	etag := first.Header.Get("ETag")
	require.NotEmpty(t, etag)

	res := serveUI(t, "/ui/", map[string]string{"Accept-Encoding": "gzip", "If-None-Match": etag})
	assert.Equal(t, http.StatusNotModified, res.StatusCode)
}

// TestFrontendEmbedsEveryBuiltFile guards the embed pattern: plain go:embed drops names starting with _ or ., and Vite names route chunks like _rack-*.js
func TestFrontendEmbedsEveryBuiltFile(t *testing.T) {
	entries, err := os.ReadDir("build/ui/assets")
	if os.IsNotExist(err) {
		t.Skip("web UI not built; run pnpm build in internal/frontend")
	}
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	ui := setupFrontend()
	for _, e := range entries {
		req := httptest.NewRequest(http.MethodHead, "/ui/assets/"+e.Name(), nil)
		rec := httptest.NewRecorder()
		ui.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, e.Name())
	}
}

func TestFrontendNotBuilt(t *testing.T) {
	ui := newFrontend(fstest.MapFS{".gitkeep": {}})
	for _, target := range []string{"/ui/", "/ui/nodes"} {
		rec := httptest.NewRecorder()
		ui.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
		assert.Equal(t, http.StatusNotFound, rec.Code, target)
		assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"), target)
		assert.Contains(t, rec.Body.String(), "web UI is not built", target)
	}
}
