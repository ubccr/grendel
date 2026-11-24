// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package render

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ubccr/grendel/internal/settings"
	"github.com/ubccr/grendel/pkg/model"
)

// beaconTimeoutMargin is how much of a boot token's life is left once a beacon started with it gives up on the boot, for the time between the config being rendered and the beacon starting, and for its last report to be retried
const beaconTimeoutMargin = 10 * time.Minute

// Request is one node asking the provision server for something, which is everything a template can see
type Request struct {
	Host  *model.Host
	NIC   *model.NetInterface
	Image *model.BootImage
	// Server is the address the node reaches the provision server at, with a port only when the node gave one
	Server  string
	Token   string
	Headers http.Header
}

// Data is what a template rendered for r sees
func Data(r Request) map[string]any {
	endpoints := NewEndpoints(r.Server, r.Token)

	return map[string]any{
		"token":      r.Token,
		"endpoints":  endpoints,
		"bootimage":  r.Image,
		"nic":        r.NIC,
		"host":       r.Host,
		"headers":    r.Headers,
		"extra_vars": settings.ProvisionExtraVars.Get(),
		"beacon": map[string]any{
			"enabled":  settings.ProvisionBeaconEnabled.Get(),
			"served":   settings.ProvisionBeaconDir.Get() != "",
			"path":     settings.ProvisionBeaconPath.Get(),
			"complete": settings.ProvisionBeaconComplete.Get(),
			"url":      endpoints.BeaconURL(r.Host.ArchDebian()),
			"timeout":  beaconTimeout(r.Token, time.Now()),
		},
	}
}

// beaconTimeout is how long a beacon given this token should wait for the boot before reporting it failed, as a duration for its --timeout. The report is only taken while the token is valid, so the wait ends before the token does. It is empty when tokens never expire, which leaves the beacon its own default
func beaconTimeout(token string, now time.Time) string {
	expires, err := model.BootTokenExpiry(token)
	if err != nil || expires.IsZero() {
		return ""
	}

	left := expires.Sub(now)
	left -= min(beaconTimeoutMargin, left/2)
	if left < time.Second {
		return ""
	}
	return left.Round(time.Second).String()
}

// AddCommandLine renders the image's kernel command line into data, where the iPXE templates read it
func AddCommandLine(image *model.BootImage, data map[string]any) error {
	commandLine, err := RenderString("cmdline", image.CommandLine, data)
	if err != nil {
		return err
	}

	data["commandLine"] = commandLine
	return nil
}

// ErrNoFile is a boot file an image does not have, such as an initrd past its last
var ErrNoFile = errors.New("no such boot file")

// FilePath is where on disk the file an image serves as name is for the node data describes. name is kernel or initrd-N, counting from 0
func FilePath(image *model.BootImage, name string, data map[string]any) (string, error) {
	if name == "kernel" {
		return RenderString(name, image.KernelPath, data)
	}

	index, isInitrd := strings.CutPrefix(name, "initrd-")
	i, err := strconv.Atoi(index)
	if !isInitrd || err != nil || i < 0 || i >= len(image.InitrdPaths) {
		return "", fmt.Errorf("%w: %s", ErrNoFile, name)
	}
	return RenderString(name, image.InitrdPaths[i], data)
}

// File is a file a boot image names, with the path its template renders to for one node
type File struct {
	// Name is the file as the provision server serves it: kernel, initrd-0 and so on
	Name string
	Path string
	Err  error
}

// Files renders the image's kernel and initrd paths with data
func Files(image *model.BootImage, data map[string]any) []File {
	names := []string{"kernel"}
	for i := range image.InitrdPaths {
		names = append(names, fmt.Sprintf("initrd-%d", i))
	}

	files := make([]File, 0, len(names))
	for _, name := range names {
		path, err := FilePath(image, name, data)
		files = append(files, File{Name: name, Path: path, Err: err})
	}
	return files
}
