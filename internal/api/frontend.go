package api

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
)

//go:embed build
var build embed.FS

type frontendFileSystem struct {
	root http.FileSystem
}

func (fsys *frontendFileSystem) Open(name string) (http.File, error) {
	f, err := fsys.root.Open(name)
	if os.IsNotExist(err) {
		return fsys.root.Open("ui/index.html")
	}
	return f, err
}

func setupFrontend() http.Handler {
	buildFs, err := fs.Sub(build, "build")
	if err != nil {
		log.Error(err)
	}

	ffsys := &frontendFileSystem{http.FS(buildFs)}

	return http.FileServer(ffsys)
}
