// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

// build holds the web UI that pnpm builds into it; git tracks only build/.gitkeep, so a fresh clone compiles and serves notBuiltPage until the UI is built. all: keeps files whose names start with _ or ., which plain embed drops; route chunks like _rack-*.js and the .gitkeep are named that way
//
//go:embed all:build
var build embed.FS

// assetDir holds the files Vite names by content hash, so a given path never changes and can be cached for good
const assetDir = "ui/assets/"

// notBuiltPage is served in place of the UI when the binary was built without one, which is the case for a plain go build of a fresh clone
const notBuiltPage = `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Grendel</title></head>
<body style="font-family: system-ui, sans-serif; max-width: 40rem; margin: 4rem auto; padding: 0 1rem; line-height: 1.5">
<h1>The Grendel web UI is not built</h1>
<p>This grendeld was compiled without the web UI. Build it into the binary with:</p>
<pre style="background: #f4f4f5; padding: 1rem; border-radius: 0.5rem">cd internal/frontend
pnpm install
pnpm build
cd -
go build ./cmd/grendeld</pre>
<p>Or, to work on the UI, run <code>pnpm dev</code> in <code>internal/frontend</code>, which serves it with live reload and proxies the API to a grendeld on 127.0.0.1:8080.</p>
</body>
</html>
`

// minGzipSize is the smallest body worth compressing; below it the gzip header and the CPU cost outweigh the saving
const minGzipSize = 1024

// frontend serves the embedded web UI. Hashed assets are cached by the browser for a year, everything else revalidates against an ETag, and text is gzipped once and kept in memory
type frontend struct {
	fsys  fs.FS
	files sync.Map // name -> *frontendFile
}

// frontendFile is one embedded file with its validator and, when worth it, its gzipped body
type frontendFile struct {
	data   []byte
	ctype  string
	etag   string
	gz     []byte
	gzEtag string
}

func setupFrontend() http.Handler {
	buildFs, err := fs.Sub(build, "build")
	if err != nil {
		log.Error(err)
	}

	return newFrontend(buildFs)
}

func newFrontend(fsys fs.FS) *frontend {
	return &frontend{fsys: fsys}
}

func (f *frontend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	file, err := f.file(name)
	if err != nil {
		// a missing asset is a stale chunk from an older build, so it gets a 404 rather than index.html, which the browser would otherwise cache for a year under a script's name
		if strings.HasPrefix(name, assetDir) {
			http.NotFound(w, r)
			return
		}
		// any other path is a client-side route, which the UI's index.html resolves
		name = "ui/index.html"
		if file, err = f.file(name); err != nil {
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, notBuiltPage)
			return
		}
	}

	h := w.Header()
	if strings.HasPrefix(name, assetDir) {
		h.Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		h.Set("Cache-Control", "no-cache")
	}
	h.Set("Content-Type", file.ctype)

	body, etag := file.data, file.etag
	if file.gz != nil {
		h.Add("Vary", "Accept-Encoding")
		if acceptsGzip(r.Header.Get("Accept-Encoding")) {
			body, etag = file.gz, file.gzEtag
			h.Set("Content-Encoding", "gzip")
		}
	}
	h.Set("ETag", etag)

	// the zero modtime leaves Last-Modified out, as embedded files have none, so the ETag is the only validator
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(body))
}

// file loads an embedded file once and keeps it, with its ETag and gzipped body, for later requests
func (f *frontend) file(name string) (*frontendFile, error) {
	if v, ok := f.files.Load(name); ok {
		return v.(*frontendFile), nil
	}

	data, err := fs.ReadFile(f.fsys, name)
	if err != nil {
		return nil, err
	}

	ctype := mime.TypeByExtension(path.Ext(name))
	if ctype == "" {
		ctype = http.DetectContentType(data)
	}

	sum := sha256.Sum256(data)
	tag := hex.EncodeToString(sum[:16])
	file := &frontendFile{data: data, ctype: ctype, etag: `"` + tag + `"`}

	if len(data) >= minGzipSize && compressible(ctype) {
		var buf bytes.Buffer
		zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
		if _, err := zw.Write(data); err == nil && zw.Close() == nil && buf.Len() < len(data) {
			file.gz = buf.Bytes()
			file.gzEtag = `"` + tag + `-gzip"`
		}
	}

	v, _ := f.files.LoadOrStore(name, file)
	return v.(*frontendFile), nil
}

// compressible reports whether a content type is text that gzip shrinks; images, fonts and the like are already compressed
func compressible(ctype string) bool {
	mediaType, _, _ := mime.ParseMediaType(ctype)
	switch {
	case strings.HasPrefix(mediaType, "text/"):
		return true
	case mediaType == "application/javascript", mediaType == "application/json", mediaType == "image/svg+xml", mediaType == "application/manifest+json":
		return true
	}
	return false
}

// acceptsGzip reports whether an Accept-Encoding header allows gzip, honoring an explicit q=0 refusal
func acceptsGzip(header string) bool {
	for _, part := range strings.Split(header, ",") {
		coding, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		coding = strings.ToLower(strings.TrimSpace(coding))
		if coding != "gzip" && coding != "*" {
			continue
		}
		q := strings.ReplaceAll(strings.ToLower(params), " ", "")
		return !(q == "q=0" || q == "q=0.0" || q == "q=0.00" || q == "q=0.000")
	}
	return false
}
